---
title: "Overview"
weight: 1
---

## What is daqq?

**daqq** (Distributed Agreement on Quantum Queries) is a Cosmos SDK chain whose purpose is **not** to transfer value, but to give a P2P network of independent participants a way to **cooperate on quantum-algorithm experiments without trusting any one party**. Concretely, daqq lets nodes:

1. **Agree on the same fresh random seed at the same block height**, with no single point of trust.
2. **Feed that seed into a quantum algorithm** that every node runs identically.
3. **Commit each node's results back to a shared ledger** so the whole network has an auditable, reproducible trail.

## Why no rewards?

Most public blockchains pay block producers in a native token. daqq deliberately does not:

- The ledger is intended for **collaborative scientific or experimental use**, not for economic activity.
- Participation is its own reward — running a node means contributing to (and being able to verify) a shared randomness beacon and a shared archive of quantum-algorithm results.
- Removing the reward token simplifies the trust model: there is no MEV to extract, no fee market to design, no inflation schedule to argue about.

{{< callout type="info" >}}
A `stake` denomination still exists for validator bonding (this is a Cosmos SDK requirement for proof-of-stake consensus), but there is no token issuance for users or block production.
{{< /callout >}}

## Why shared randomness?

Most interesting quantum algorithms (random-circuit sampling, variational ansatz training, randomized benchmarking, random Hamiltonian generation, …) take a **random input** somewhere. In a distributed setting, "use a random number" usually means each node picks its own — and they disagree, so nothing they produce is comparable.

For experiments that need every node to do the **same** randomized thing at the **same** logical moment, you need an agreed-upon random value.

Common bad options:

- **Trust one node** — single point of failure.
- **Use the block hash** — biasable by the proposer.
- **Use timestamps** — divergent across nodes, also biasable.

daqq's beacon module uses a **RANDAO-style commit-reveal scheme**: every participant commits to a secret they choose, later reveals it, and the chain XORs all reveals together. As long as **at least one honest participant** picks an unpredictable secret, the resulting seed is unpredictable to everyone in advance.

## What can you do with the seed?

Anything that benefits from "every node, same random input, same logical moment, results recorded for everyone to compare". Some general patterns:

- **Cross-validation of quantum hardware vs. simulator.** Every node simulates / executes the same parameterised quantum algorithm and submits its outcome; divergences are immediately visible.
- **Reproducible benchmarking.** A canonical sequence of randomized instances, archived on-chain, that anyone can replay against new hardware or new software.
- **Distributed scientific bookkeeping.** Researchers running independent experiments can anchor each run to a beacon round, so the random input is provably not cherry-picked.

## Many algorithms, one beacon

daqq is structured as a **multi-problem platform**: the [`problems`]({{< relref "modules/problems" >}}) module keeps an on-chain registry of every algorithm the network supports, and participants are free to pick any of them to run against each round's shared seed. Adding a new algorithm means shipping a new Cosmos SDK module that consumes `Seeds[roundID]`, computes its result, and writes to its own ledger. New modules are rolled out via gov upgrade. See [Problem System]({{< relref "problem-system" >}}) for the full design.

## Example: Problem #1, `random_circuit`

The chain ships with one concrete problem to make the platform tangible. The [`random_circuit`]({{< relref "modules/random_circuit" >}}) module:

1. Takes the round's seed.
2. Generates a random quantum circuit deterministically from it (parameterised by `(seed, width, depth)`).
3. Has every participant compute the circuit's **theoretical output probability distribution** locally.
4. Records each participant's distribution on-chain so divergences across nodes become visible.

`random_circuit` is just one example consumer of the beacon. Other algorithms — VQE on random Hamiltonians, randomized benchmarking sequences, random Clifford sampling, etc. — can be added as additional problem modules without touching the beacon or the existing ledger.
