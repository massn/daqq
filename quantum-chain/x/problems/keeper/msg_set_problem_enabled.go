package keeper

import (
	"bytes"
	"context"

	errorsmod "cosmossdk.io/errors"

	"quantum-chain/x/problems/types"
)

// SetProblemEnabled flips the Enabled flag of a registered problem.
// Authority-gated: only the configured authority (defaults to x/gov) may call.
func (k msgServer) SetProblemEnabled(ctx context.Context, req *types.MsgSetProblemEnabled) (*types.MsgSetProblemEnabledResponse, error) {
	authority, err := k.addressCodec.StringToBytes(req.Authority)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}
	if !bytes.Equal(k.GetAuthority(), authority) {
		expected, _ := k.addressCodec.BytesToString(k.GetAuthority())
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", expected, req.Authority)
	}

	if err := k.SetEnabled(ctx, req.Id, req.Enabled); err != nil {
		return nil, err
	}
	return &types.MsgSetProblemEnabledResponse{}, nil
}
