# quantumchain

**quantumchain** is a blockchain built using Cosmos SDK and Tendermint and created with [Ignite CLI](https://ignite.com/cli).

## Quick Start

### Prerequisites

- [Go](https://go.dev/doc/install) 1.22 or higher
- [Ignite CLI](https://ignite.com/cli)

### Start the Chain

To start the blockchain in development mode, run:

```bash
cd quantum-chain
ignite chain serve
```

The `serve` command downloads dependencies, compiles the source code, initializes the blockchain configuration, and starts the node.

### Install and Run Binary

To install the `quantumchaind` binary:

```bash
cd quantum-chain
make install
```

To run the node:

```bash
quantumchaind start
```

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

### Web Frontend (Visualization)

 The node embeds a web visualizer for the network-shared random seeds produced by the beacon module.

 1. Enable the REST API (`api.enable = true` in `~/.quantumchain/config/app.toml`), then start the chain: `quantumchaind start`.
 2. Open `http://localhost:1317/gui/` in your web browser.
 3. The node serves the page, the seed endpoint (`/gui/seeds`), and the network endpoint (`/gui/net_info`) from the same origin, so no CORS configuration is needed.
 4. The page shows the latest shared random seed and draws the node's peer network.

### Web Frontend (Ignite Scaffolding)

 Additionally, Ignite CLI offers a frontend scaffolding feature (based on Vue) to help you quickly build a web frontend for your blockchain:

Use: `ignite scaffold vue`
This command can be run within your scaffolded blockchain project.

For more information see the [monorepo for Ignite front-end development](https://github.com/ignite/web).

## Release

To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```bash
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

### Install

To install the latest version of your blockchain node's binary, execute the following command on your machine:

```bash
curl https://get.ignite.com/username/quantum-chain@latest! | sudo bash
```

`username/quantum-chain` should match the `username` and `repo_name` of the Github repository to which the source code was pushed. Learn more about [the install process](https://github.com/ignite/installer).

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
- [Developer Chat](https://discord.com/invite/ignitecli)
