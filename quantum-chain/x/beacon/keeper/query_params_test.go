package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quantum-chain/x/beacon/keeper"
	"quantum-chain/x/beacon/types"
)

func TestParamsQuery(t *testing.T) {
	f := initFixture(t)

	qs := keeper.NewQueryServerImpl(f.keeper)
	params := types.DefaultParams()
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	response, err := qs.Params(f.ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestParamsQueryInvalidRequest(t *testing.T) {
	f := initFixture(t)

	qs := keeper.NewQueryServerImpl(f.keeper)

	// A nil request must be rejected with InvalidArgument rather than panicking.
	response, err := qs.Params(f.ctx, nil)
	require.Nil(t, response)
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}
