package types

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type BeaconKeeper interface {
	// TODO Add methods imported from beacon should be defined here
	GetSeed(context.Context, uint64) (string, error)
}

// StakingKeeper is the slice of x/staking that x/random_circuit depends on to
// gate result submission to active validators. Restricting submitters to the
// bonded validator set caps the participant count at the staking module's
// MaxValidators and makes submission sybil-resistant (an attacker cannot flood
// results from throwaway addresses without bonding stake into the active set).
type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
}

// ProblemsKeeper is the slice of x/problems' API that x/random_circuit
// depends on. Kept primitive-typed so this module does not import
// x/problems' proto types.
type ProblemsKeeper interface {
	RegisterProblem(ctx context.Context, name, moduleName, description string) (uint64, error)
	IsEnabled(ctx context.Context, id uint64) (bool, error)
}

// AuthKeeper defines the expected interface for the Auth module.
type AuthKeeper interface {
	AddressCodec() address.Codec
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI // only used for simulation
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// ParamSubspace defines the expected Subspace interface for parameters.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
