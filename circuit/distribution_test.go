package circuit

import (
	"math"
	"reflect"
	"strconv"
	"testing"
)

// TestDistributionDeterministic is the property daqq's cross-validation relies
// on: the same seed must yield a byte-identical distribution every time, so
// independent participants agree exactly.
func TestDistributionDeterministic(t *testing.T) {
	const seed, width, depth = 12345, 5, 10

	first, err := MakeRandomQC(seed, width, depth).Distribution()
	if err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}
	for i := range 5 {
		got, err := MakeRandomQC(seed, width, depth).Distribution()
		if err != nil {
			t.Fatalf("Distribution() error on run %d: %v", i, err)
		}
		if !reflect.DeepEqual(first, got) {
			t.Fatalf("distribution not deterministic: run %d differs from run 0", i)
		}
	}
}

// TestDistributionDifferentSeeds guards against a degenerate implementation that
// ignores the seed: different shared seeds should map to different circuits.
func TestDistributionDifferentSeeds(t *testing.T) {
	a, err := MakeRandomQC(1, 5, 10).Distribution()
	if err != nil {
		t.Fatal(err)
	}
	b, err := MakeRandomQC(2, 5, 10).Distribution()
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(a, b) {
		t.Fatal("distinct seeds produced identical distributions")
	}
}

// TestDistributionSumsToOne checks the output is a valid probability
// distribution (probabilities sum to ~1).
func TestDistributionSumsToOne(t *testing.T) {
	dist, err := MakeRandomQC(777, 5, 10).Distribution()
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.0
	for _, p := range dist {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			t.Fatalf("probability %q is not a valid float: %v", p, err)
		}
		sum += v
	}
	if math.Abs(sum-1.0) > 1e-9 {
		t.Fatalf("probabilities sum to %v, want ~1.0", sum)
	}
}
