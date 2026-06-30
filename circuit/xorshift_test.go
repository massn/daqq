package circuit

import (
	"testing"
)

func TestNewXorShift32(t *testing.T) {
	seed := uint32(12345)
	gateSet := []string{"H", "S", "T"}

	generator := NewXorShift32(seed, gateSet)

	if generator == nil {
		t.Fatal("NewXorShift32 returned nil")
	}

	if generator.state != seed {
		t.Errorf("state = %d, want %d", generator.state, seed)
	}

	if len(generator.gateSet) != len(gateSet) {
		t.Errorf("gateSet length = %d, want %d", len(generator.gateSet), len(gateSet))
	}

	if generator.gateSetSize != uint32(len(gateSet)) {
		t.Errorf("gateSetSize = %d, want %d", generator.gateSetSize, len(gateSet))
	}
}

func TestXorShift32_Next(t *testing.T) {
	seed := uint32(12345)
	gateSet := []string{"H", "S", "T"}
	generator := NewXorShift32(seed, gateSet)

	// Generate several gates and check they are from the gate set
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		gate := generator.Next()

		// Check gate is in the gate set
		found := false
		for _, g := range gateSet {
			if gate == g {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Next() returned %s, which is not in gate set", gate)
		}

		seen[gate] = true
	}

	// Check that we've seen multiple different gates (probabilistic test)
	if len(seen) < 2 {
		t.Error("Expected to see multiple different gates, but only saw one type")
	}
}

func TestXorShift32_Deterministic(t *testing.T) {
	seed := uint32(12345)
	gateSet := []string{"H", "S", "T"}

	// Create two generators with the same seed
	gen1 := NewXorShift32(seed, gateSet)
	gen2 := NewXorShift32(seed, gateSet)

	// They should produce the same sequence
	for i := 0; i < 20; i++ {
		gate1 := gen1.Next()
		gate2 := gen2.Next()
		if gate1 != gate2 {
			t.Errorf("Iteration %d: gen1 = %s, gen2 = %s (should be equal)", i, gate1, gate2)
		}
	}
}

func TestXorShift32_DifferentSeeds(t *testing.T) {
	gateSet := []string{"H", "S", "T"}

	gen1 := NewXorShift32(12345, gateSet)
	gen2 := NewXorShift32(54321, gateSet)

	// Different seeds should eventually produce different sequences
	same := 0
	total := 20
	for i := 0; i < total; i++ {
		if gen1.Next() == gen2.Next() {
			same++
		}
	}

	// Not all should be the same (probabilistic test)
	if same == total {
		t.Error("Different seeds produced identical sequences (very unlikely)")
	}
}

func TestXorShift32_SingleGate(t *testing.T) {
	seed := uint32(12345)
	gateSet := []string{"H"}
	generator := NewXorShift32(seed, gateSet)

	// With only one gate, should always return that gate
	for i := 0; i < 10; i++ {
		gate := generator.Next()
		if gate != "H" {
			t.Errorf("Expected H, got %s", gate)
		}
	}
}
