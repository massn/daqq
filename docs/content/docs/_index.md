---
title: "Documentation"
weight: 1
sidebar:
  open: true
---

Welcome to the daqq documentation.

- [Overview]({{< relref "overview" >}}) — what daqq is and the design intent
- [Concepts]({{< relref "concepts" >}}) — terminology (block, commit, reveal, round, seed) and the round lifecycle
- [Architecture]({{< relref "architecture" >}}) — module structure and execution order
- [Problem System]({{< relref "problem-system" >}}) — design spec for the multi-problem framework
- [Known limitations]({{< relref "limitations" >}}) — open issues, fairness caveats, and severity assessments
- Modules
  - [beacon]({{< relref "modules/beacon" >}}) — RANDAO commit-reveal randomness beacon
  - [random_circuit]({{< relref "modules/random_circuit" >}}) — Problem #1: theoretical output distribution ledger
  - [problems]({{< relref "modules/problems" >}}) — on-chain problem registry
  - [quantumchain]({{< relref "modules/quantumchain" >}}) — base module
- [Operations]({{< relref "operations/quickstart" >}}) — quickstart and localnet
