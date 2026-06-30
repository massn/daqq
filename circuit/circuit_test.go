package circuit

import (
	"testing"
)

func TestMakeRandomQC(t *testing.T) {
	tests := []struct {
		name  string
		seed  uint32
		width uint32
		depth uint32
	}{
		{"basic circuit", 12345, 4, 8},
		{"single qubit", 54321, 1, 4},
		{"large circuit", 99999, 8, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qc := MakeRandomQC(tt.seed, tt.width, tt.depth)

			if qc == nil {
				t.Fatal("MakeRandomQC returned nil")
			}

			if qc.Width != tt.width {
				t.Errorf("Width = %d, want %d", qc.Width, tt.width)
			}

			if qc.Depth != tt.depth {
				t.Errorf("Depth = %d, want %d", qc.Depth, tt.depth)
			}

			if len(qc.GateSet) == 0 {
				t.Error("GateSet is empty")
			}

			if len(qc.Gates) == 0 {
				t.Error("Gates is empty")
			}

			// Validate all gates
			for i, gate := range qc.Gates {
				if err := gate.Validate(qc.Width); err != nil {
					t.Errorf("Gate %d is invalid: %v", i, err)
				}
			}
		})
	}
}

func TestGateValidate(t *testing.T) {
	tests := []struct {
		name    string
		gate    *Gate
		width   uint32
		wantErr bool
	}{
		{
			name: "valid H gate",
			gate: &Gate{
				GateName:    "H",
				TargetQubit: uint32Ptr(0),
			},
			width:   4,
			wantErr: false,
		},
		{
			name: "valid CNOT gate",
			gate: &Gate{
				GateName:     "CNOT",
				ControlQubit: uint32Ptr(0),
				TargetQubit:  uint32Ptr(1),
			},
			width:   4,
			wantErr: false,
		},
		{
			name: "nil target qubit",
			gate: &Gate{
				GateName:    "H",
				TargetQubit: nil,
			},
			width:   4,
			wantErr: true,
		},
		{
			name: "target qubit out of range",
			gate: &Gate{
				GateName:    "H",
				TargetQubit: uint32Ptr(5),
			},
			width:   4,
			wantErr: true,
		},
		{
			name: "CNOT without control qubit",
			gate: &Gate{
				GateName:     "CNOT",
				ControlQubit: nil,
				TargetQubit:  uint32Ptr(1),
			},
			width:   4,
			wantErr: true,
		},
		{
			name: "CNOT with same control and target",
			gate: &Gate{
				GateName:     "CNOT",
				ControlQubit: uint32Ptr(1),
				TargetQubit:  uint32Ptr(1),
			},
			width:   4,
			wantErr: true,
		},
		{
			name: "empty gate name",
			gate: &Gate{
				GateName:    "",
				TargetQubit: uint32Ptr(0),
			},
			width:   4,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.gate.Validate(tt.width)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplySingleQubitGates(t *testing.T) {
	generator := NewXorShift32(12345, []string{"H", "S", "T"})
	width := uint32(4)

	gates := applySingleQubitGates(width, generator)

	if len(gates) != int(width) {
		t.Errorf("Expected %d gates, got %d", width, len(gates))
	}

	for i, gate := range gates {
		if gate.TargetQubit == nil {
			t.Errorf("Gate %d has nil target qubit", i)
			continue
		}
		if *gate.TargetQubit != uint32(i) {
			t.Errorf("Gate %d target qubit = %d, want %d", i, *gate.TargetQubit, i)
		}
	}
}

func TestApplyCNOTGatesEven(t *testing.T) {
	width := uint32(8)
	gates := applyCNOTGatesEven(width)

	expectedCount := int(width / 2)
	if len(gates) != expectedCount {
		t.Errorf("Expected %d gates, got %d", expectedCount, len(gates))
	}

	for i, gate := range gates {
		if gate.GateName != "CNOT" {
			t.Errorf("Gate %d name = %s, want CNOT", i, gate.GateName)
		}
		if gate.ControlQubit == nil || gate.TargetQubit == nil {
			t.Errorf("Gate %d has nil qubit", i)
			continue
		}
		expectedControl := uint32(i * 2)
		expectedTarget := uint32(i*2 + 1)
		if *gate.ControlQubit != expectedControl {
			t.Errorf("Gate %d control = %d, want %d", i, *gate.ControlQubit, expectedControl)
		}
		if *gate.TargetQubit != expectedTarget {
			t.Errorf("Gate %d target = %d, want %d", i, *gate.TargetQubit, expectedTarget)
		}
	}
}

func TestApplyCNOTGatesOdd(t *testing.T) {
	width := uint32(8)
	gates := applyCNOTGatesOdd(width)

	// Should be (width/2 - 1) gates for odd pairs
	expectedCount := int(width/2) - 1
	if len(gates) != expectedCount {
		t.Errorf("Expected %d gates, got %d", expectedCount, len(gates))
	}

	for i, gate := range gates {
		if gate.GateName != "CNOT" {
			t.Errorf("Gate %d name = %s, want CNOT", i, gate.GateName)
		}
		if gate.ControlQubit == nil || gate.TargetQubit == nil {
			t.Errorf("Gate %d has nil qubit", i)
			continue
		}
		expectedControl := uint32(i*2 + 1)
		expectedTarget := uint32(i*2 + 2)
		if *gate.ControlQubit != expectedControl {
			t.Errorf("Gate %d control = %d, want %d", i, *gate.ControlQubit, expectedControl)
		}
		if *gate.TargetQubit != expectedTarget {
			t.Errorf("Gate %d target = %d, want %d", i, *gate.TargetQubit, expectedTarget)
		}
	}
}

// Helper function to create uint32 pointer
func uint32Ptr(v uint32) *uint32 {
	return &v
}
