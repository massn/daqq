package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"quantum-chain/x/beacon/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	// Randao State
	RoundInfo collections.Item[uint64]
	Commits   collections.Map[collections.Pair[uint64, string], string]
	Reveals   collections.Map[collections.Pair[uint64, string], string]
	Seeds     collections.Map[uint64, string]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

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

		Params:    collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		RoundInfo: collections.NewItem(sb, types.BoundInfoKey, "round_info", collections.Uint64Value),
		Commits:   collections.NewMap(sb, types.CommitsKey, "commits", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), collections.StringValue),
		Reveals:   collections.NewMap(sb, types.RevealsKey, "reveals", collections.PairKeyCodec(collections.Uint64Key, collections.StringKey), collections.StringValue),
		Seeds:     collections.NewMap(sb, types.SeedsKey, "seeds", collections.Uint64Key, collections.StringValue),
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
