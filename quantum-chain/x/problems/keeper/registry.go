package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	"quantum-chain/x/problems/types"
)

// RegistrationRequest is the input for Register.
type RegistrationRequest struct {
	Name         string
	ModuleName   string
	Kind         types.ProblemKind
	Description  string
	AddedAtRound uint64
}

// Register inserts a new Problem and bumps NextProblemID.
//
// If a problem with the same name already exists, Register returns it without
// modifying state. This makes the call idempotent and safe to invoke from
// genesis init or upgrade handlers across restarts.
//
// New problems are created with Enabled=true. Use SetEnabled to disable.
func (k Keeper) Register(ctx context.Context, req RegistrationRequest) (types.Problem, error) {
	if req.Name == "" {
		return types.Problem{}, errorsmod.Wrap(types.ErrInvalidProblem, "name is required")
	}
	if req.ModuleName == "" {
		return types.Problem{}, errorsmod.Wrap(types.ErrInvalidProblem, "module_name is required")
	}

	// Idempotency: if the name is already registered, return the existing entry.
	if existingID, err := k.ProblemsByName.Get(ctx, req.Name); err == nil {
		existing, err := k.Problems.Get(ctx, existingID)
		if err != nil {
			return types.Problem{}, err
		}
		return existing, nil
	} else if !errorsIsNotFound(err) {
		return types.Problem{}, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return types.Problem{}, err
	}

	id := params.NextProblemId
	problem := types.Problem{
		Id:           id,
		Name:         req.Name,
		ModuleName:   req.ModuleName,
		Kind:         req.Kind,
		Enabled:      true,
		AddedAtRound: req.AddedAtRound,
		Description:  req.Description,
	}

	if err := k.Problems.Set(ctx, id, problem); err != nil {
		return types.Problem{}, err
	}
	if err := k.ProblemsByName.Set(ctx, problem.Name, id); err != nil {
		return types.Problem{}, err
	}

	params.NextProblemId = id + 1
	if err := k.Params.Set(ctx, params); err != nil {
		return types.Problem{}, err
	}

	return problem, nil
}

// GetByID fetches a problem by its registry ID.
func (k Keeper) GetByID(ctx context.Context, id uint64) (types.Problem, error) {
	p, err := k.Problems.Get(ctx, id)
	if err != nil {
		if errorsIsNotFound(err) {
			return types.Problem{}, errorsmod.Wrapf(types.ErrProblemNotFound, "id %d", id)
		}
		return types.Problem{}, err
	}
	return p, nil
}

// GetByName fetches a problem by its unique name.
func (k Keeper) GetByName(ctx context.Context, name string) (types.Problem, error) {
	id, err := k.ProblemsByName.Get(ctx, name)
	if err != nil {
		if errorsIsNotFound(err) {
			return types.Problem{}, errorsmod.Wrapf(types.ErrProblemNotFound, "name %q", name)
		}
		return types.Problem{}, err
	}
	return k.GetByID(ctx, id)
}

// SetEnabled flips the enabled flag of an existing problem.
func (k Keeper) SetEnabled(ctx context.Context, id uint64, enabled bool) error {
	p, err := k.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if p.Enabled == enabled {
		return nil
	}
	p.Enabled = enabled
	return k.Problems.Set(ctx, id, p)
}

// IsEnabled is a convenience helper for problem modules that want to gate
// their message handlers on the registry state.
func (k Keeper) IsEnabled(ctx context.Context, id uint64) (bool, error) {
	p, err := k.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return p.Enabled, nil
}

// RegisterProblem is a primitive-typed convenience wrapper around Register
// intended for use by other modules' expected_keepers interfaces. It always
// registers BUILTIN problems (the only Kind in the MVP). Returns the assigned
// problem ID. Idempotent on Name.
func (k Keeper) RegisterProblem(ctx context.Context, name, moduleName, description string) (uint64, error) {
	p, err := k.Register(ctx, RegistrationRequest{
		Name:        name,
		ModuleName:  moduleName,
		Kind:        types.ProblemKind_PROBLEM_KIND_BUILTIN,
		Description: description,
	})
	if err != nil {
		return 0, err
	}
	return p.Id, nil
}

// errorsIsNotFound reports whether err is the collections "not found" sentinel.
func errorsIsNotFound(err error) bool {
	return err != nil && (err == collections.ErrNotFound || errorsmod.IsOf(err, collections.ErrNotFound))
}
