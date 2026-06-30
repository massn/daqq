---
title: "Known limitations"
weight: 5
---

This page collects open issues, fairness caveats, and security trade-offs in daqq's current design. Each entry states **what the issue is**, **how severe it is today**, and **what would change if the network grew or if a different problem class were added**.

Severity ratings:

- **Low** — observable in principle, no realistic impact at current scale or with shipped problems.
- **Medium** — could affect fairness or correctness once a specific feature is introduced (e.g. submission deadlines).
- **High** — breaks a core property (seed unpredictability, ledger integrity, etc.).

There are currently **no High-severity issues** in the shipped code. Everything below is Low or Medium and is documented so future contributors avoid stepping on the same rake.

## 1. Simultaneity: who actually starts computing first?

**Issue.** When the beacon finalises `Seeds[R]` in the EndBlocker of block `H = 50·(R+1)`, the seed is *logically* available to all nodes at the same height. In reality, every node receives block `H` over the CometBFT p2p gossip network with **different latency** (well-connected nodes receive it tens to hundreds of milliseconds before peripheral ones). A node that finishes processing block `H` first can start running the quantum algorithm seeded by `Seeds[R]` first.

**How big is it today?**

- The shipped `random_circuit` problem has **no submission deadline**. As long as the result lands before the next round (or even later), nothing is gained by starting early. Severity: **Low**.
- Quantum simulation/execution typically takes orders of magnitude longer (seconds to minutes) than block propagation jitter (~hundreds of ms). The head start is in the noise.
- daqq has **no rewards**, so "who gets there first" doesn't translate into economic advantage.

**When it would matter.**

- A future problem with a **per-round submission deadline** would advantage centrally-located nodes.
- A future problem that records **wall-clock latency** (e.g. randomized benchmarking that compares hardware speeds) would conflate "fast hardware" with "fast gossip path".

**Mitigations if it ever becomes a real problem.**

- Delay submission acceptance by `k` blocks past the seed finalisation. Every node has block `H` by `H+k` for any realistic `k≥2`, so the gossip jitter is masked.
- Require submissions to commit at block `H+m` for a fixed `m`, with the actual payload revealed at `H+m+n` — i.e. mirror commit-reveal on the result side.
- For latency-sensitive experiments, record `(seed_available_at_block, submitted_at_block)` per node and treat block deltas, not wall-clock, as the latency metric.

## 2. Predictable-seed window (block offsets 46 – 49)

**Issue.** Reveals are rejected after offset 45, but the seed is not officially stored in `Seeds[roundID]` until the EndBlocker at offset 50. During the 4 blocks in between, every accepted reveal is already on-chain and the seed is just `SHA256(XOR(reveals))` — anyone can compute it locally before the chain announces it. See [Concepts → Lifecycle of one round]({{< relref "concepts#lifecycle-of-one-round" >}}).

**How big is it today?**

- Same logic as Issue #1: no shipped problem has a deadline, so being "4 blocks early on the seed" buys nothing. Severity: **Low**.

**When it would matter.**

- Identical conditions to Issue #1 (deadlines or latency-sensitive problems).

**Mitigations.**

- Move `RevealEnd` to `RoundDuration - 1` (offset 49) so the gap shrinks to zero.
- Or accept the window and gate all problem submissions on `Seeds[R]` being **written** (which the code already does via `GetSeed`).

## 3. Last-revealer withholding (RANDAO bias)

**Issue.** A participant can see what the seed *would* be by computing `SHA256(XOR(others' reveals XOR my secret))` before they reveal. If they dislike the result, they can simply **not reveal**. Their secret is then excluded from the XOR, and the seed becomes `SHA256(XOR(others))` instead. This gives the withholder a binary choice over each of their identities.

**How big is it today?**

- An attacker with `m` controlled identities can choose among `2^m` possible seeds per round.
- Severity for cryptographic unpredictability: **Low** — `2^m` is tiny next to `2^256`.
- Severity for problem-specific bias: **depends on the problem**. For `random_circuit` the bias is essentially undetectable in the output distribution. For a future problem whose "interesting outcome" lives in a small subset of seed space, the attacker could nudge toward it.

**Mitigations.**

- Slash withholders (would require introducing a stake/penalty mechanism — at odds with daqq's no-reward design).
- Layer a VDF on top: the seed becomes `VDF(SHA256(XOR(reveals)))` with a delay longer than the reveal window, so the withholder can no longer predict the outcome of withholding before deciding. Out of scope for MVP.

## 4. Empty rounds (no reveals at all)

**Issue.** If a round closes with **zero** valid reveals, `abci.go` skips the `Seeds.Set` call entirely — `Seeds[roundID]` simply doesn't exist for that round. Any problem submission for that round will be rejected with `ErrSeedNotReady`.

**How big is it today?**

- Severity: **Low**. The chain keeps progressing; only that round produces no seed and therefore no problem results. Round R+1 starts immediately.
- It is silent: there is no explicit "round skipped" event today.

**When it would matter.**

- Statistical analyses that assume "one seed per round" need to filter out missing rounds.
- Long-running tooling should not infinite-wait for `Seeds[R]` of a skipped round.

**Mitigations.**

- Emit a `RoundSkipped{R}` event in EndBlocker when `count == 0`.
- Document the "skipped round" semantics in the SDK clients before adding any code that assumes continuity.

## 5. Per-round participation cost

**Issue.** Every participating node must broadcast **two transactions per round** (one `MsgCommit`, one `MsgReveal`) just to contribute to the seed, before doing any actual quantum work. With `RoundDuration = 50` and a few-second block time, that is one round per ~minutes — manageable but not free.

**How big is it today?**

- Severity: **Low** on a small experimental network. Tx volume is dominated by problem submissions, not by commits/reveals.
- daqq has no fee market, so the cost is operational (uptime, key custody) rather than monetary.

**When it would matter.**

- A network with thousands of validators would multiply on-chain state proportionally. `Commits` and `Reveals` collections grow with `(rounds × participants)`.

**Mitigations.**

- Prune `Commits[r]` and `Reveals[r]` after `Seeds[r]` is finalised (the seed is the only durable artifact).
- Allow batched commit-reveal across rounds.

## 6. No submission deadline on problem modules

**Issue.** `random_circuit.MsgSubmitResult` accepts a result for round `R` at any later block, as long as `Seeds[R]` exists. There is no late-cutoff.

**How big is it today?**

- Severity: **Low**. It makes the ledger maximally inclusive: a node that was offline can still backfill round 42's distribution next week. State storage cost is the only downside.

**When it would matter.**

- A problem that compares "live" results (e.g. quantum hardware availability windows) would want deadlines.

**Mitigations.**

- Add a per-problem `submission_deadline_blocks` parameter in `x/problems`. When introducing one, **first** read Issues #1 and #2 — they become Medium-severity the moment deadlines are real.

## 7. Reveal hashing convention

**Issue.** `msg_server_reveal.go` computes `sha256.Sum256([]byte(msg.Secret))` — i.e. it hashes the **hex-string bytes** of the secret, not the **raw 32 bytes**. The aggregation in `abci.go` then hex-decodes the same secret to 32 raw bytes for XOR. So the on-chain check is "did you commit to the hex string?" while the on-chain use is "I'll XOR the raw bytes". Functionally consistent, but the two representations are intermingled.

**How big is it today?**

- Severity: **Low**. It works. No security loss — committing to the hex form fixes the raw form just as well.
- The risk is **clarity**: a future contributor who assumes "the commit covers the raw bytes" could subtly break the protocol.

**Mitigations.**

- Pick one representation (raw bytes recommended) and convert at the SDK boundary, not inside the keeper.
- Add a test that pins `commit = sha256_hex(secret_hex)` so the convention is locked.

## Out of scope / not problems

A few things readers sometimes ask about but that are not issues here:

- **No proof-of-quantumness.** daqq does not verify that a participant actually ran a quantum computer; it just records what they submitted. This is intentional — the value is the **shared, reproducible trail**, not adjudication.
- **Validator centralisation.** Cosmos SDK governance / staking applies as normal. daqq does not change validator economics because there is no reward to distribute.
- **No native token.** Treated as a feature, not a bug. See [Overview → Why no rewards?]({{< relref "overview#why-no-rewards" >}}).
