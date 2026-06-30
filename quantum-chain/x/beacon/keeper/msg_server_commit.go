package keeper

import (
	"context"

	"quantum-chain/x/beacon/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Commit(ctx context.Context, msg *types.MsgCommit) (*types.MsgCommitResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// Logic
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := uint64(sdkCtx.BlockHeight())
	const RoundDuration = 50
	const CommitEnd = 30

	roundID := blockHeight / RoundDuration
	offset := blockHeight % RoundDuration

	if msg.RoundId != roundID {
		return nil, errorsmod.Wrapf(types.ErrInvalidPhase, "msg round id %d does not match current round %d", msg.RoundId, roundID)
	}

	if offset > CommitEnd {
		return nil, errorsmod.Wrap(types.ErrInvalidPhase, "not in commit phase")
	}

	// Check existing
	exists, err := k.Commits.Has(ctx, collections.Join(msg.RoundId, msg.Creator))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errorsmod.Wrap(types.ErrAlreadyCommitted, "already committed for this round")
	}

	// Store
	if err := k.Commits.Set(ctx, collections.Join(msg.RoundId, msg.Creator), msg.Hash); err != nil {
		return nil, err
	}

	return &types.MsgCommitResponse{}, nil
}
