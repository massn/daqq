package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"quantum-chain/x/random_circuit/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema  collections.Schema
	Params  collections.Item[types.Params]
	Results collections.Map[collections.Pair[uint64, string], types.ResultData]

	beaconKeeper   types.BeaconKeeper
	problemsKeeper types.ProblemsKeeper
	stakingKeeper  types.StakingKeeper

	// ProblemID is the registry ID this module owns. Populated during
	// InitGenesis once the entry is registered with x/problems.
	ProblemID collections.Item[uint64]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	beaconKeeper types.BeaconKeeper,
	problemsKeeper types.ProblemsKeeper,
	stakingKeeper types.StakingKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		beaconKeeper:   beaconKeeper,
		problemsKeeper: problemsKeeper,
		stakingKeeper:  stakingKeeper,
		Params:         collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Results:        collections.NewMap(sb, types.ResultsKey, "results", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), codec.CollValue[types.ResultData](cdc)),
		ProblemID:      collections.NewItem(sb, types.ProblemIDKey, "problem_id", collections.Uint64Value),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}
