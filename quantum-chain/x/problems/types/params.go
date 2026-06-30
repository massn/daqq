package types

import "fmt"

// FirstProblemID is the ID assigned to the first registered problem.
const FirstProblemID uint64 = 1

// NewParams creates a new Params instance.
func NewParams(nextProblemID uint64) Params {
	return Params{NextProblemId: nextProblemID}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(FirstProblemID)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.NextProblemId == 0 {
		return fmt.Errorf("next_problem_id must be >= 1")
	}
	return nil
}
