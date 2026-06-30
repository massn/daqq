package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"quantum-chain/x/beacon/types"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker executes the Randao logic at the end of each block.
func (k Keeper) EndBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := uint64(sdkCtx.BlockHeight())

	// Example Round Duration: 50 blocks
	const RoundDuration = 50
	const RevealDuration = 15 // Blocks 0-30: Commit, 31-45: Reveal, 50: Finalize

	// Check if this block is the end of a round
	if blockHeight > 0 && blockHeight%RoundDuration == 0 {
		roundID := (blockHeight / RoundDuration) - 1
		// Aggregation Phase

		// Iterate over reveals for this round
		// Note: Collections iteration.
		// We need to fetch all reveals for the current roundID.
		rng := collections.NewPrefixedPairRange[uint64, string](roundID)
		iter, err := k.Reveals.Iterate(ctx, rng)
		if err != nil {
			return err
		}
		defer iter.Close()

		var combinedSecret []byte
		count := 0

		for ; iter.Valid(); iter.Next() {
			secret, err := iter.Value()
			if err != nil {
				return err
			}
			secretBytes, _ := hex.DecodeString(secret) // Assume validated hex
			if count == 0 {
				combinedSecret = secretBytes
			} else {
				// XOR
				if len(secretBytes) != len(combinedSecret) {
					// Pad or error? For simplicity assume fixed length sha256
					continue
				}
				for i := range combinedSecret {
					combinedSecret[i] ^= secretBytes[i]
				}
			}
			count++
		}

		if count > 0 {
			finalHash := sha256.Sum256(combinedSecret)
			finalSeed := hex.EncodeToString(finalHash[:])

			// Store Seed
			if err := k.Seeds.Set(ctx, roundID, finalSeed); err != nil {
				return err
			}

			// Emit Event
			sdkCtx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeNewRound,
					sdk.NewAttribute(types.AttributeKeyRoundID, fmt.Sprintf("%d", roundID)),
					sdk.NewAttribute(types.AttributeKeySeed, finalSeed),
				),
			)
		}
	}
	return nil
}
