package problems

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"quantum-chain/x/problems/types"
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
				{
					RpcMethod: "ListProblems",
					Use:       "list-problems",
					Short:     "List all registered problems",
				},
				{
					RpcMethod:      "GetProblem",
					Use:            "get-problem [id]",
					Short:          "Get a registered problem by ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "GetProblemByName",
					Use:            "get-problem-by-name [name]",
					Short:          "Get a registered problem by name",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "name"}},
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
					RpcMethod: "SetProblemEnabled",
					Skip:      true, // authority-gated; invoked via gov proposal
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}
