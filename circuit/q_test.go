package circuit

import (
	"os"
	"testing"
)

func TestQC_ExecuteCircuit(t *testing.T) {
	qc := MakeRandomQC(12345, 4, 8)

	qsim, qubits, err := qc.executeCircuit()
	if err != nil {
		t.Fatalf("executeCircuit() error = %v", err)
	}

	if qsim == nil {
		t.Error("executeCircuit() returned nil simulator")
	}

	if len(qubits) != int(qc.Width) {
		t.Errorf("executeCircuit() returned %d qubits, want %d", len(qubits), qc.Width)
	}

	state := qsim.State()
	if len(state) == 0 {
		t.Error("executeCircuit() produced empty state")
	}

	// Check that probabilities sum to ~1
	sum := 0.0
	for _, s := range state {
		sum += s.Probability()
	}
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("Total probability = %f, want ~1.0", sum)
	}
}

func TestQC_ExecuteCircuit_InvalidGate(t *testing.T) {
	// Create a QC with an invalid gate
	invalidTarget := uint32(10)
	qc := &QC{
		Width: 4,
		Gates: []*Gate{
			{
				GateName:    "H",
				TargetQubit: &invalidTarget, // Out of range
			},
		},
	}

	_, _, err := qc.executeCircuit()
	if err == nil {
		t.Error("executeCircuit() should return error for invalid gate")
	}
}

func TestQC_RunShots(t *testing.T) {
	qc := MakeRandomQC(12345, 3, 4)
	shots := 1000

	results, err := qc.RunShots(shots)
	if err != nil {
		t.Fatalf("RunShots() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("RunShots() returned empty results")
	}

	// Count total shots
	totalShots := uint64(0)
	for _, count := range results {
		totalShots += count
	}

	if totalShots != uint64(shots) {
		t.Errorf("Total shots = %d, want %d", totalShots, shots)
	}

	// Check that all result strings are binary strings of correct length
	expectedLength := int(qc.Width)
	for state := range results {
		if len(state) != expectedLength {
			t.Errorf("State %s has length %d, want %d", state, len(state), expectedLength)
		}
		for _, c := range state {
			if c != '0' && c != '1' {
				t.Errorf("State %s contains non-binary character %c", state, c)
			}
		}
	}
}

func TestQC_RunShots_InvalidCircuit(t *testing.T) {
	// Create a QC with an invalid gate
	invalidTarget := uint32(10)
	qc := &QC{
		Width: 4,
		Gates: []*Gate{
			{
				GateName:    "H",
				TargetQubit: &invalidTarget,
			},
		},
	}

	_, err := qc.RunShots(100)
	if err == nil {
		t.Error("RunShots() should return error for invalid circuit")
	}
}

func TestQC_State(t *testing.T) {
	qc := MakeRandomQC(12345, 3, 4)

	// Remove existing output file if it exists
	os.Remove(DefaultOutputFile)

	err := qc.State()
	if err != nil {
		t.Fatalf("State() error = %v", err)
	}

	// Check that output file was created
	if _, err := os.Stat(DefaultOutputFile); os.IsNotExist(err) {
		t.Errorf("State() did not create output file %s", DefaultOutputFile)
	}

	// Clean up
	os.Remove(DefaultOutputFile)
}

func TestQC_State_InvalidCircuit(t *testing.T) {
	// Create a QC with an invalid gate
	invalidTarget := uint32(10)
	qc := &QC{
		Width: 4,
		Gates: []*Gate{
			{
				GateName:    "H",
				TargetQubit: &invalidTarget,
			},
		},
	}

	err := qc.State()
	if err == nil {
		t.Error("State() should return error for invalid circuit")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants are defined with reasonable values
	if DefaultGrids <= 0 {
		t.Errorf("DefaultGrids = %d, want positive value", DefaultGrids)
	}

	if DefaultOutputFile == "" {
		t.Error("DefaultOutputFile is empty")
	}

	if DefaultChartTitle == "" {
		t.Error("DefaultChartTitle is empty")
	}

	if DefaultChartSubtitle == "" {
		t.Error("DefaultChartSubtitle is empty")
	}
}
