package beacon

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	beaconsimulation "quantum-chain/x/beacon/simulation"
	"quantum-chain/x/beacon/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	beaconGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&beaconGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgCommit          = "op_weight_msg_beacon"
		defaultWeightMsgCommit int = 100
	)

	var weightMsgCommit int
	simState.AppParams.GetOrGenerate(opWeightMsgCommit, &weightMsgCommit, nil,
		func(_ *rand.Rand) {
			weightMsgCommit = defaultWeightMsgCommit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCommit,
		beaconsimulation.SimulateMsgCommit(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgReveal          = "op_weight_msg_beacon"
		defaultWeightMsgReveal int = 100
	)

	var weightMsgReveal int
	simState.AppParams.GetOrGenerate(opWeightMsgReveal, &weightMsgReveal, nil,
		func(_ *rand.Rand) {
			weightMsgReveal = defaultWeightMsgReveal
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgReveal,
		beaconsimulation.SimulateMsgReveal(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
