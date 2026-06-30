package keeper_test

import "context"

// stubProblemsKeeper is a minimal in-memory ProblemsKeeper used by unit tests
// that don't exercise the registration path. It always claims problems are
// enabled and assigns ID 1 on registration.
type stubProblemsKeeper struct{}

func (stubProblemsKeeper) RegisterProblem(_ context.Context, _, _, _ string) (uint64, error) {
	return 1, nil
}

func (stubProblemsKeeper) IsEnabled(_ context.Context, _ uint64) (bool, error) {
	return true, nil
}
