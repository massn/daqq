---
title: "random_circuit"
weight: 2
---

The `x/random_circuit` module (formerly `x/qcledger`) is **Problem #1** in daqq's Problem System. It records participants' **theoretical output probability distributions** for the random quantum circuit generated from each beacon round's seed.

{{< callout type="info" >}}
**Case A vs Case B** — `x/random_circuit` handles the analytical/theoretical distribution (Case A). A future `x/random_circuit_sampling` module will handle empirical shot histograms (Case B) as a separate problem in the registry.
{{< /callout >}}

## Registration with x/problems

`x/random_circuit` registers itself in the [problems](problems) registry at chain genesis. Look it up by name:

```bash
quantumchaind query problems get-problem-by-name random_circuit
```

```yaml
problem:
  id: "1"
  name: random_circuit
  module_name: random_circuit
  kind: PROBLEM_KIND_BUILTIN
  enabled: true
  description: Theoretical output distribution of a randomly generated quantum circuit (case A).
```

The keeper persists the assigned `ProblemID` in its own store so other modules and clients can derive it without re-querying.

## Dependency on beacon

Every submission carries a `roundID`. Before accepting it, the handler calls:

```go
// quantum-chain/x/random_circuit/keeper/msg_server_submit_result.go
_, err := k.beaconKeeper.GetSeed(ctx, msg.RoundId)
if err != nil {
    return nil, errorsmod.Wrapf(types.ErrSeedNotReady, "seed for round %d not ready", msg.RoundId)
}
```

This guarantees:

1. Every recorded result is tied to a seed every node already agrees on.
2. Honest nodes generating the same circuit from that seed produce the same expected output, making divergent results easy to spot.

## Submission shape

`MsgSubmitResult` carries the participant's full theoretical distribution:

```proto
message MsgSubmitResult {
  string creator = 1;
  uint64 round_id = 2;
  Distribution distribution = 3;
}

message Distribution {
  // basis state (e.g. "00101") -> probability as a decimal string
  map<string, string> probabilities = 1;
}
```

Probabilities are encoded as decimal strings to avoid floating-point divergence across nodes during consensus serialisation.

A submitter may submit at most one solution per `(round, problem)` pair. Re-submissions are rejected.

## TODO

This page still needs:

- Validation rules for `Distribution` (must sum to ~1, dimension = `2^width`, byte cap).
- Query endpoints for stored submissions.
- Worked example of a full submit flow.
