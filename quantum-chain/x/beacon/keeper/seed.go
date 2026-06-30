package keeper

import (
	"context"
)

// GetSeed returns the seed for a given roundID.
func (k Keeper) GetSeed(ctx context.Context, roundID uint64) (string, error) {
	return k.Seeds.Get(ctx, roundID)
}
