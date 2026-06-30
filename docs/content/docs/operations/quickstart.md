---
title: "Quickstart"
weight: 1
---

All commands assume you are in the repository root and have [Task](https://taskfile.dev/) installed.

## Install the node binary

```bash
task -t Taskfile.quickstart.yml install
```

This compiles and installs `quantumchaind` into `$GOPATH/bin`.

## Single-node chain

```bash
task -t Taskfile.quickstart.yml quickstart
```

Equivalent to running `quickstart:init` (one-time) and then `quickstart:start`. Home directory: `~/.quantumchain`.

To reset:

```bash
task -t Taskfile.quickstart.yml quickstart:clean
```

## Produce shared randomness

The beacon module derives a network-shared random seed every 50 blocks from the
participants' commit/reveal contributions. Even a single node can drive this:
run the beacon agent in its own terminal while the node is up, and `alice` will
commit and reveal each round automatically.

```bash
task -t Taskfile.quickstart.yml beacon:loop
```

A round spans 50 blocks (commit in offset 0–30, reveal in 31–45, finalize at the
next boundary), so the first seed appears within a few minutes. Stop the agent
with `Ctrl+C`.

## See the node in a GUI

The node ships with a built-in web visualizer, served by its own REST API server at `/gui`. It shows the network-shared random seeds (the beacon output) and draws the node's peer network.

`quickstart:init` already enables the REST API, so with the node running just open the visualizer in your browser:

```bash
task -t Taskfile.quickstart.yml gui
```

This opens `http://localhost:1317/gui/`. Because the page, the seed endpoint (`/gui/seeds`), and the network endpoint (`/gui/net_info`) are all served from the same origin by the node itself, no CORS configuration or separate web server is needed. The status badge turns green once it connects, the top panel shows the latest shared random seed (and recent rounds), and the network panel shows this node and its connected peers.

## 3-node localnet (alice / bob / carol)

```bash
task -t Taskfile.quickstart.yml localnet:init
```

Then start each node in its own terminal:

```bash
task -t Taskfile.quickstart.yml localnet:start:alice
task -t Taskfile.quickstart.yml localnet:start:bob
task -t Taskfile.quickstart.yml localnet:start:carol
```

Check propagation:

```bash
task -t Taskfile.quickstart.yml localnet:status
```

Reset:

```bash
task -t Taskfile.quickstart.yml localnet:clean
```

## Run a random quantum circuit (standalone)

```bash
task -t Taskfile.quickstart.yml circuit
```

Generates a random quantum circuit using a timestamp seed. (Wiring this generator to the beacon seed is a planned integration — see [Beacon → Integration](../modules/beacon#integration).)
