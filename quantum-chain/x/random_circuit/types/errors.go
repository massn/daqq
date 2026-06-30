package types

import "cosmossdk.io/errors"

var (
	ErrInvalidSigner    = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrSample           = errors.Register(ModuleName, 1101, "sample error")
	ErrSeedNotReady     = errors.Register(ModuleName, 1102, "seed not ready")
	ErrAlreadySubmitted = errors.Register(ModuleName, 1103, "already submitted")
	ErrNotValidator     = errors.Register(ModuleName, 1104, "submitter is not an active validator")
)
