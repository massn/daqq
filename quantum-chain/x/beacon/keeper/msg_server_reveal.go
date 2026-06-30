package keeper

import (
	"context"

	"quantum-chain/x/beacon/types"

	"crypto/sha256"
	"encoding/hex"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Reveal(ctx context.Context, msg *types.MsgReveal) (*types.MsgRevealResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// Logic
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := uint64(sdkCtx.BlockHeight())
	const RoundDuration = 50
	const CommitEnd = 30
	const RevealEnd = 45

	roundID := blockHeight / RoundDuration
	offset := blockHeight % RoundDuration

	if msg.RoundId != roundID {
		return nil, errorsmod.Wrapf(types.ErrInvalidPhase, "msg round id %d does not match current round %d", msg.RoundId, roundID)
	}

	if offset <= CommitEnd || offset > RevealEnd {
		return nil, errorsmod.Wrap(types.ErrInvalidPhase, "not in reveal phase")
	}

	// Verify Commit exists
	committedHash, err := k.Commits.Get(ctx, collections.Join(msg.RoundId, msg.Creator))
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrNoCommit, "commit not found")
	}

	// Verify Hash
	hash := sha256.Sum256([]byte(msg.Secret)) // Hex string or raw bytes? User input is usually hex string if cli, but internal logic says Hash(Secret).
	// Assuming Secret is hex string, we decode it first? Or just hash the string bytes?
	// Design says: Hash(Secret).
	// Let's assume Secret is passed as string hex.
	// We need to match what Commit submitted.
	// Simplify: Hash of string bytes.
	computedHash := hex.EncodeToString(hash[:])

	if computedHash != committedHash {
		return nil, errorsmod.Wrap(types.ErrInvalidReveal, "hash mismatch")
	}

	// Store Reveal
	if err := k.Reveals.Set(ctx, collections.Join(msg.RoundId, msg.Creator), msg.Secret); err != nil {
		return nil, err
	}

	return &types.MsgRevealResponse{}, nil
}
