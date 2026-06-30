package types

import "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenIDs := make(map[uint64]struct{}, len(gs.Problems))
	seenNames := make(map[string]struct{}, len(gs.Problems))
	for i, p := range gs.Problems {
		if p.Id == 0 {
			return fmt.Errorf("genesis problems[%d]: id must be >= 1", i)
		}
		if p.Id >= gs.Params.NextProblemId {
			return fmt.Errorf("genesis problems[%d]: id %d must be < next_problem_id %d",
				i, p.Id, gs.Params.NextProblemId)
		}
		if _, ok := seenIDs[p.Id]; ok {
			return fmt.Errorf("genesis problems[%d]: duplicate id %d", i, p.Id)
		}
		seenIDs[p.Id] = struct{}{}

		if p.Name == "" {
			return fmt.Errorf("genesis problems[%d]: name is required", i)
		}
		if _, ok := seenNames[p.Name]; ok {
			return fmt.Errorf("genesis problems[%d]: duplicate name %q", i, p.Name)
		}
		seenNames[p.Name] = struct{}{}

		if p.ModuleName == "" {
			return fmt.Errorf("genesis problems[%d]: module_name is required", i)
		}
	}

	return nil
}
