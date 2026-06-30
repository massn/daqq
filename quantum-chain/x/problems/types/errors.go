package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/problems module sentinel errors
var (
	ErrInvalidSigner       = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrProblemNotFound     = errors.Register(ModuleName, 1101, "problem not found")
	ErrProblemNameTaken    = errors.Register(ModuleName, 1102, "problem name already registered")
	ErrInvalidProblem      = errors.Register(ModuleName, 1103, "invalid problem")
)
