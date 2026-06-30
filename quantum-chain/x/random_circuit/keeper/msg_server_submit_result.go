package keeper

import (
	"context"

	"quantum-chain/x/random_circuit/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SubmitResult(ctx context.Context, msg *types.MsgSubmitResult) (*types.MsgSubmitResultResponse, error) {
	creatorBz, err := k.addressCodec.StringToBytes(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	// 0. Only active (bonded) validators may submit results. This caps the
	// participant set at the staking module's MaxValidators and is
	// sybil-resistant: an attacker cannot flood results from throwaway
	// addresses without bonding stake into the active set. The validator
	// operator address shares the creator account's bytes.
	validator, err := k.stakingKeeper.GetValidator(ctx, sdk.ValAddress(creatorBz))
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNotValidator, "submitter %s is not a validator", msg.Creator)
	}
	if !validator.IsBonded() {
		return nil, errorsmod.Wrapf(types.ErrNotValidator, "validator %s is not in the active set", msg.Creator)
	}

	// Logic
	// 1. Check if seed exists for the round
	_, err = k.beaconKeeper.GetSeed(ctx, msg.RoundId)
	if err != nil {
		// If seed not found, it returns error usually or empty.
		// Collections Get returns ErrNotFound if strictly checking, but GetSeed wrapper might return error.
		return nil, errorsmod.Wrapf(types.ErrSeedNotReady, "seed for round %d not ready", msg.RoundId)
	}

	// 2. Check if already submitted
	exists, err := k.Results.Has(ctx, collections.Join(msg.RoundId, msg.Creator))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errorsmod.Wrap(types.ErrAlreadySubmitted, "result already submitted")
	}

	// 3. Store
	//
	// We persist the theoretical probability distribution (case A) that the
	// participant computed for this round. A separate sampling-based problem
	// (case B) will live in a future `random_circuit_sampling` module and
	// store ShotCounts instead.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	resultData := types.ResultData{
		Address:      msg.Creator,
		Distribution: msg.Distribution,
		SubmittedAt:  sdkCtx.BlockTime(),
		BlockHeight:  sdkCtx.BlockHeight(),
	}

	if err := k.Results.Set(ctx, collections.Join(msg.RoundId, msg.Creator), resultData); err != nil {
		return nil, err
	}

	return &types.MsgSubmitResultResponse{}, nil
}
