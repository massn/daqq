package types_test

import (
	"testing"

	"quantum-chain/x/problems/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc:     "zero-value genesis is invalid (next_problem_id must be >= 1)",
			genState: &types.GenesisState{},
			valid:    false,
		},
		{
			desc: "genesis with pre-seeded problem matching NextProblemID",
			genState: &types.GenesisState{
				Params: types.NewParams(2),
				Problems: []types.Problem{{
					Id:         1,
					Name:       "random_circuit",
					ModuleName: "random_circuit",
					Kind:       types.ProblemKind_PROBLEM_KIND_BUILTIN,
					Enabled:    true,
				}},
			},
			valid: true,
		},
		{
			desc: "duplicate problem name is invalid",
			genState: &types.GenesisState{
				Params: types.NewParams(3),
				Problems: []types.Problem{
					{Id: 1, Name: "a", ModuleName: "a", Enabled: true},
					{Id: 2, Name: "a", ModuleName: "b", Enabled: true},
				},
			},
			valid: false,
		},
		{
			desc: "problem id >= NextProblemID is invalid",
			genState: &types.GenesisState{
				Params: types.NewParams(1),
				Problems: []types.Problem{
					{Id: 1, Name: "a", ModuleName: "a", Enabled: true},
				},
			},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
