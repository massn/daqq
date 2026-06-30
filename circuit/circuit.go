package circuit

import "fmt"

// RandomGateGenerator yields a deterministic stream of gate names, one per
// call to Next, used to fill the single-qubit layers of a random circuit.
type RandomGateGenerator interface {
	Next() string // return gate name
}

// Gate is a single quantum gate in a circuit. TargetQubit is always set;
// ControlQubit is set only for two-qubit gates such as CNOT.
type Gate struct {
	ControlQubit *uint32 `json:"control_qubit,omitempty"`
	TargetQubit  *uint32 `json:"target_qubit"`
	GateName     string  `json:"gate_name"`
}

// Validate checks if the gate configuration is valid
func (g *Gate) Validate(width uint32) error {
	if g.TargetQubit == nil {
		return fmt.Errorf("target qubit cannot be nil")
	}
	if *g.TargetQubit >= width {
		return fmt.Errorf("target qubit %d is out of range (width: %d)", *g.TargetQubit, width)
	}
	if g.GateName == "" {
		return fmt.Errorf("gate name cannot be empty")
	}

	// For two-qubit gates, validate control qubit
	if g.GateName == "CNOT" {
		if g.ControlQubit == nil {
			return fmt.Errorf("control qubit cannot be nil for CNOT gate")
		}
		if *g.ControlQubit >= width {
			return fmt.Errorf("control qubit %d is out of range (width: %d)", *g.ControlQubit, width)
		}
		if *g.ControlQubit == *g.TargetQubit {
			return fmt.Errorf("control and target qubits must be different")
		}
	}

	return nil
}

// QC is a generated random quantum circuit: its gate set, dimensions
// (Width qubits over Depth layers), and the ordered list of gates to apply.
type QC struct {
	GateSet []string `json:"gate_set"`
	Width   uint32   `json:"width"`
	Depth   uint32   `json:"depth"`
	Gates   []*Gate  `json:"gates"`
}

// applySingleQubitGates applies a random single-qubit gate to each qubit
func applySingleQubitGates(width uint32, generator RandomGateGenerator) []*Gate {
	gates := make([]*Gate, 0, width)
	for j := 0; j < int(width); j++ {
		tq := uint32(j)
		gates = append(gates, &Gate{
			TargetQubit: &tq,
			GateName:    generator.Next(),
		})
	}
	return gates
}

// applyCNOTGatesEven applies CNOT gates to even-indexed pairs (0-1, 2-3, ...)
func applyCNOTGatesEven(width uint32) []*Gate {
	gates := make([]*Gate, 0, width/2)
	for j := 0; j < int(width/2); j++ {
		cq := uint32(j * 2)
		tq := uint32(j*2 + 1)
		gates = append(gates, &Gate{
			GateName:     "CNOT",
			ControlQubit: &cq,
			TargetQubit:  &tq,
		})
	}
	return gates
}

// applyCNOTGatesOdd applies CNOT gates to odd-indexed pairs (1-2, 3-4, ...)
func applyCNOTGatesOdd(width uint32) []*Gate {
	gates := make([]*Gate, 0, width/2)
	for j := 0; j < int(width/2); j++ {
		if j*2+2 >= int(width) {
			continue
		}
		cq := uint32(j*2 + 1)
		tq := uint32(j*2 + 2)
		gates = append(gates, &Gate{
			GateName:     "CNOT",
			ControlQubit: &cq,
			TargetQubit:  &tq,
		})
	}
	return gates
}

// MakeRandomQC builds a random quantum circuit of the given width (qubits) and
// depth (layers) from seed. Layers cycle through single-qubit gates (H/S/T),
// even CNOT pairs, single-qubit gates again, then odd CNOT pairs, so the same
// seed always produces the same circuit.
func MakeRandomQC(seed, width, depth uint32) *QC {
	oneQubitGateNames := []string{"H", "S", "T"}
	generator := NewXorShift32(seed, oneQubitGateNames)

	gates := []*Gate{}
	for i := 0; i < int(depth); i++ {
		switch i % 4 {
		case 0, 2:
			gates = append(gates, applySingleQubitGates(width, generator)...)
		case 1:
			gates = append(gates, applyCNOTGatesEven(width)...)
		case 3:
			gates = append(gates, applyCNOTGatesOdd(width)...)
		}
	}

	return &QC{
		GateSet: append(oneQubitGateNames, "CNOT"),
		Width:   width,
		Depth:   depth,
		Gates:   gates,
	}
}
