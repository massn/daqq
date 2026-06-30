package keeper

import (
	"context"
	"sort"

	"quantum-chain/x/problems/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
//
// The Params are written first so subsequent Register-style helpers see the
// correct NextProblemID. Genesis problems are then inserted directly with
// their pre-assigned IDs (no NextProblemID consumption) so that exported and
// imported state round-trip identically.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		return err
	}

	for _, p := range genState.Problems {
		if err := k.Problems.Set(ctx, p.Id, p); err != nil {
			return err
		}
		if err := k.ProblemsByName.Set(ctx, p.Name, p.Id); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	iter, err := k.Problems.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	problems, err := iter.Values()
	if err != nil {
		return nil, err
	}

	// Deterministic ordering by ID for reproducible export.
	sort.Slice(problems, func(i, j int) bool { return problems[i].Id < problems[j].Id })

	return &types.GenesisState{Params: params, Problems: problems}, nil
}
