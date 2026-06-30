---
title: "daqq"
toc: false
---

<p style="text-align: center; font-size: 1.25rem; font-weight: 600; margin: 1.5rem 0;">Shared inputs, shared records — on a blockchain.</p>

**daqq** (Distributed Agreement on Quantum Queries) is a no-reward distributed ledger that lets P2P nodes **agree on the same fresh random value at the same block height**, then run **identical quantum algorithms** seeded by it and **record each node's result on-chain** for the whole network to compare and audit.

It is not a payment network, not a smart-contract platform, and not a benchmarking service for one specific algorithm. It is a **shared, tamper-evident trail** of "what happened when every node was given the same random input" — useful for cross-validating quantum hardware and simulators, for reproducible randomized benchmarks, and for distributed scientific bookkeeping.

<div style="display:flex; gap:0.75rem; justify-content:center; flex-wrap:wrap; margin:2rem 0;">
  <a href="/gui/" style="display:inline-block; padding:0.75rem 1.6rem; border-radius:10px; background:#3b82f6; color:#fff; font-weight:700; text-decoration:none;">▶ Live dashboard (GUI)</a>
  <a href="/docs/" style="display:inline-block; padding:0.75rem 1.6rem; border-radius:10px; border:2px solid #3b82f6; color:#3b82f6; font-weight:700; text-decoration:none;">📖 Documentation</a>
  <a href="https://discord.gg/pxrjYJKKF" style="display:inline-block; padding:0.75rem 1.6rem; border-radius:10px; background:#5865F2; color:#fff; font-weight:700; text-decoration:none;">💬 Discord</a>
</div>

<p style="text-align:center; color:#6b7280; font-size:0.9rem; margin-top:-1rem;">The live dashboard shows shared seeds, registered problems, the node network, and per-round cross-validation across independent validators.</p>

{{< cards >}}
  {{< card link="docs/overview" title="Overview" subtitle="What daqq is and why it exists." >}}
  {{< card link="docs/concepts" title="Concepts" subtitle="Block, commit, reveal, round, seed, problem." >}}
  {{< card link="docs/architecture" title="Architecture" subtitle="Cosmos SDK modules and execution order." >}}
  {{< card link="docs/problem-system" title="Problem System" subtitle="How multiple algorithms coexist on one beacon." >}}
  {{< card link="docs/modules/beacon" title="Beacon Protocol" subtitle="RANDAO commit-reveal seed agreement." >}}
  {{< card link="docs/limitations" title="Known limitations" subtitle="Fairness, simultaneity, and security caveats." >}}
  {{< card link="docs/operations/quickstart" title="Quickstart" subtitle="Run a single-node or 3-node localnet." >}}
{{< /cards >}}

## Key properties

- **No reward token.** Participants run the ledger together for its own sake, not for incentives. No MEV, no fee market, no inflation schedule.
- **Shared randomness.** Every 50 blocks, all nodes derive the same 256-bit seed via RANDAO-style commit-reveal + XOR aggregation. As long as one honest participant contributes a high-entropy secret, the seed is unpredictable to everyone in advance.
- **Multi-problem by design.** The chain is a platform: each quantum algorithm lives in its own Cosmos SDK module, registered in the on-chain `problems` registry. New algorithms ship as new modules via gov upgrade. The first one — `random_circuit` — generates a random circuit from each round's seed and records every participant's theoretical output distribution.
- **Auditable, reproducible.** Anyone can re-derive the seed from on-chain reveals and re-derive each algorithm's input from the seed. Disagreements between nodes are visible and replayable.
