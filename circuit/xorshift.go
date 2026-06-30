package circuit

// XorShift32 is a deterministic RandomGateGenerator backed by a 32-bit xorshift
// PRNG. Each generated value is mapped onto a gate name from its gate set.
type XorShift32 struct {
	gateSet     []string
	gateSetSize uint32
	state       uint32
}

// NewXorShift32 creates a new XorShift32 random gate generator
func NewXorShift32(seed uint32, gs []string) *XorShift32 {
	return &XorShift32{
		state:       seed,
		gateSet:     gs,
		gateSetSize: uint32(len(gs)),
	}
}

// Next advances the xorshift state and returns the next gate name from the
// generator's gate set.
func (x *XorShift32) Next() string {
	x.state ^= x.state << 13
	x.state ^= x.state >> 17
	x.state ^= x.state << 5
	return x.gateSet[x.state%x.gateSetSize]
}

// SeedState folds an arbitrary-length seed (e.g. the beacon's full 256-bit
// shared random) into the 32-bit xorshift state using FNV-1a. Every byte of the
// seed influences the result, so the whole shared random — not just its first
// four bytes — selects the circuit. Mirrored by the GUI's seedState() in JS.
func SeedState(seed []byte) uint32 {
	const (
		offset uint32 = 2166136261
		prime  uint32 = 16777619
	)
	h := offset
	for _, b := range seed {
		h ^= uint32(b)
		h *= prime
	}
	return h
}
