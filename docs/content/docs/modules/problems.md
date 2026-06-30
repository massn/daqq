---
title: "problems"
weight: 3
---

The `x/problems` module is the on-chain **registry** of daqq problems. It does not implement any problem itself — it tracks which problem modules exist, the IDs they have been assigned, and whether each one is currently accepting submissions.

See the [Problem System]({{< relref "problem-system" >}}) page for the overall design rationale.

## What it stores

```go
type Problem struct {
    ID           uint64
    Name         string      // unique across all problems (e.g. "random_circuit")
    ModuleName   string      // implementing Cosmos module name
    Kind         ProblemKind // BUILTIN for the MVP; WASM reserved for the future
    Enabled      bool        // false => no new submissions, history kept
    AddedAtRound uint64      // beacon round at which the problem became selectable
    Description  string
}

type Params struct {
    NextProblemID uint64      // monotonically increasing; IDs are never reused
}
```

State invariants:

- IDs are assigned sequentially and **never reused**, even after a problem is disabled.
- `Name` is unique across all problems, including disabled ones.
- Problems are **only added or toggled**, never deleted.

## How problem modules register themselves

Each problem module (e.g. `x/random_circuit`) calls `ProblemsKeeper.RegisterProblem(ctx, name, moduleName, description)` from its own `InitGenesis`. `Register` is idempotent on `name`, so re-running genesis (e.g. in tests or after `quickstart:clean`) is safe.

To make this work, the `InitGenesis` order in `app/app_config.go` puts `problems` **before** any problem module:

```go
InitGenesis: []string{
    // ...
    problemsmoduletypes.ModuleName,
    randomcircuitmoduletypes.ModuleName,
}
```

## Queries

| Command | Purpose |
|---|---|
| `quantumchaind query problems params` | Show `NextProblemID`. |
| `quantumchaind query problems list-problems` | List every registered problem (paginated). |
| `quantumchaind query problems get-problem [id]` | Look up by numeric ID. |
| `quantumchaind query problems get-problem-by-name [name]` | Look up by unique name. |

## Disabling a problem

Disabling does not delete the entry — it only flips the `enabled` flag. The implementing module's message handler is responsible for rejecting new submissions when its problem is disabled (`ProblemsKeeper.IsEnabled(ctx, id)`).

The flip itself is gov-gated: a `MsgSetProblemEnabled` is sent through `x/gov` as a regular governance proposal. On acceptance:

```yaml
tx:
  body:
    messages:
      - "@type": /quantumchain.problems.v1.MsgSetProblemEnabled
        authority: "<gov module address>"
        id: 1
        enabled: false
```

Historical submissions to the now-disabled problem remain queryable.

## Adding a new problem

New problems are added by:

1. Writing a new `x/<your_problem>` module that depends on `BeaconKeeper.GetSeed` and `ProblemsKeeper.RegisterProblem`.
2. Cutting a release with the new binary.
3. Submitting a `software-upgrade` gov proposal to roll the network onto the release.
4. On the upgrade height, the new module's `InitGenesis`/upgrade handler runs `RegisterProblem`, and the registry entry appears.

See [Problem System → Adding a new problem]({{< relref "problem-system#adding-a-new-problem-community-contribution" >}}).
