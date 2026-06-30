package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "random_circuit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	// It should be synced with the gov module's name if it is ever changed.
	// See: https://github.com/cosmos/cosmos-sdk/blob/v0.52.0-beta.2/x/gov/types/keys.go#L9
	GovModuleName = "gov"
)

// ParamsKey is the prefix to retrieve all Params
var (
	ParamsKey    = collections.NewPrefix("p_random_circuit")
	ResultsKey   = collections.NewPrefix("results")
	ProblemIDKey = collections.NewPrefix("problem_id")
)

// ProblemName is the unique name used to register this module in x/problems.
const ProblemName = "random_circuit"

// ProblemDescription is the human-readable summary stored in x/problems.
const ProblemDescription = "Theoretical output distribution of a randomly generated quantum circuit (case A)."
