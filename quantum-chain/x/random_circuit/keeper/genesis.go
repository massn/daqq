package keeper

import (
	"context"

	"quantum-chain/x/random_circuit/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
//
// In addition to writing module params, InitGenesis self-registers
// `random_circuit` as Problem #1 in x/problems. The Register call is
// idempotent, so re-running genesis (e.g. for tests or chain reset) is safe.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		return err
	}

	id, err := k.problemsKeeper.RegisterProblem(ctx, types.ProblemName, types.ModuleName, types.ProblemDescription)
	if err != nil {
		return err
	}
	if err := k.ProblemID.Set(ctx, id); err != nil {
		return err
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	var err error

	genesis := types.DefaultGenesis()
	genesis.Params, err = k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return genesis, nil
}
