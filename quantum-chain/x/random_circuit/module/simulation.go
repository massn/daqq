package randomcircuit

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	randomcircuitsimulation "quantum-chain/x/random_circuit/simulation"
	"quantum-chain/x/random_circuit/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	randomCircuitGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&randomCircuitGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgSubmitResult          = "op_weight_msg_random_circuit"
		defaultWeightMsgSubmitResult int = 100
	)

	var weightMsgSubmitResult int
	simState.AppParams.GetOrGenerate(opWeightMsgSubmitResult, &weightMsgSubmitResult, nil,
		func(_ *rand.Rand) {
			weightMsgSubmitResult = defaultWeightMsgSubmitResult
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSubmitResult,
		randomcircuitsimulation.SimulateMsgSubmitResult(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
