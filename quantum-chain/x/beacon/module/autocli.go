package beacon

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"quantum-chain/x/beacon/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "Commit",
					Use:            "commit [round-id] [hash]",
					Short:          "Send a commit tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "round_id"}, {ProtoField: "hash"}},
				},
				{
					RpcMethod:      "Reveal",
					Use:            "reveal [round-id] [secret]",
					Short:          "Send a reveal tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "round_id"}, {ProtoField: "secret"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
