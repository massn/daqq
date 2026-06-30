package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"quantum-chain/x/problems/keeper"
	"quantum-chain/x/problems/types"
)

func regReq(name, moduleName, desc string) keeper.RegistrationRequest {
	return keeper.RegistrationRequest{
		Name:        name,
		ModuleName:  moduleName,
		Kind:        types.ProblemKind_PROBLEM_KIND_BUILTIN,
		Description: desc,
	}
}

func TestRegister_AssignsSequentialIDsAndIsIdempotent(t *testing.T) {
	f := initFixture(t)

	p1, err := f.keeper.Register(f.ctx, regReq("random_circuit", "random_circuit", "first"))
	require.NoError(t, err)
	require.Equal(t, uint64(1), p1.Id)
	require.True(t, p1.Enabled)

	p2, err := f.keeper.Register(f.ctx, regReq("foo", "foo", "second"))
	require.NoError(t, err)
	require.Equal(t, uint64(2), p2.Id)

	// Re-registering an existing name returns the same entry, no state changes.
	pAgain, err := f.keeper.Register(f.ctx, regReq("random_circuit", "random_circuit", "ignored"))
	require.NoError(t, err)
	require.Equal(t, p1, pAgain)

	params, err := f.keeper.Params.Get(f.ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(3), params.NextProblemId)
}

func TestRegister_RejectsEmptyFields(t *testing.T) {
	f := initFixture(t)

	_, err := f.keeper.Register(f.ctx, regReq("", "x", ""))
	require.Error(t, err)

	_, err = f.keeper.Register(f.ctx, regReq("x", "", ""))
	require.Error(t, err)
}

func TestSetEnabledAndLookups(t *testing.T) {
	f := initFixture(t)

	p, err := f.keeper.Register(f.ctx, regReq("random_circuit", "random_circuit", ""))
	require.NoError(t, err)

	got, err := f.keeper.GetByID(f.ctx, p.Id)
	require.NoError(t, err)
	require.True(t, got.Enabled)

	gotByName, err := f.keeper.GetByName(f.ctx, "random_circuit")
	require.NoError(t, err)
	require.Equal(t, p.Id, gotByName.Id)

	require.NoError(t, f.keeper.SetEnabled(f.ctx, p.Id, false))
	got, err = f.keeper.GetByID(f.ctx, p.Id)
	require.NoError(t, err)
	require.False(t, got.Enabled)

	// Unknown lookups return ErrProblemNotFound.
	_, err = f.keeper.GetByID(f.ctx, 999)
	require.ErrorIs(t, err, types.ErrProblemNotFound)
	_, err = f.keeper.GetByName(f.ctx, "missing")
	require.ErrorIs(t, err, types.ErrProblemNotFound)
}
