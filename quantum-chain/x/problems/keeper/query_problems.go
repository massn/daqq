package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"quantum-chain/x/problems/types"
)

func (q queryServer) ListProblems(ctx context.Context, req *types.QueryListProblemsRequest) (*types.QueryListProblemsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	problems, pageRes, err := query.CollectionPaginate(ctx, q.k.Problems, req.Pagination,
		func(_ uint64, p types.Problem) (types.Problem, error) { return p, nil },
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryListProblemsResponse{Problems: problems, Pagination: pageRes}, nil
}

func (q queryServer) GetProblem(ctx context.Context, req *types.QueryGetProblemRequest) (*types.QueryGetProblemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	p, err := q.k.GetByID(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "problem id %d not found", req.Id)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryGetProblemResponse{Problem: p}, nil
}

func (q queryServer) GetProblemByName(ctx context.Context, req *types.QueryGetProblemByNameRequest) (*types.QueryGetProblemByNameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	p, err := q.k.GetByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "problem name %q not found", req.Name)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryGetProblemByNameResponse{Problem: p}, nil
}
