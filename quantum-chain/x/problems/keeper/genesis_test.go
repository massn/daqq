package keeper_test

import (
	"testing"

	"quantum-chain/x/problems/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	f := initFixture(t)
	err := f.keeper.InitGenesis(f.ctx, genesisState)
	require.NoError(t, err)
	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.Empty(t, got.Problems)
}

func TestGenesisRoundTripWithPreSeededProblems(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.NewParams(2),
		Problems: []types.Problem{{
			Id:           1,
			Name:         "random_circuit",
			ModuleName:   "random_circuit",
			Kind:         types.ProblemKind_PROBLEM_KIND_BUILTIN,
			Enabled:      true,
			AddedAtRound: 0,
			Description:  "Theoretical output distribution of a random quantum circuit.",
		}},
	}

	f := initFixture(t)
	require.NoError(t, f.keeper.InitGenesis(f.ctx, genesisState))

	got, err := f.keeper.ExportGenesis(f.ctx)
	require.NoError(t, err)
	require.EqualExportedValues(t, genesisState.Params, got.Params)
	require.Len(t, got.Problems, 1)
	require.Equal(t, genesisState.Problems[0], got.Problems[0])

	// Name index is populated and lookup works.
	p, err := f.keeper.GetByName(f.ctx, "random_circuit")
	require.NoError(t, err)
	require.Equal(t, uint64(1), p.Id)
}
