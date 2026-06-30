package types

import "cosmossdk.io/errors"

var (
	ErrInvalidSigner    = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSample           = errors.Register(ModuleName, 1101, "sample error")
	ErrInvalidPhase     = errors.Register(ModuleName, 1102, "invalid phase")
	ErrAlreadyCommitted = errors.Register(ModuleName, 1103, "already committed")
	ErrNoCommit         = errors.Register(ModuleName, 1104, "no commit found")
	ErrInvalidReveal    = errors.Register(ModuleName, 1105, "invalid secret reveal")
)
