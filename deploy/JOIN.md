# Join the daqq network as a full node (by application)

You can run a node that follows the live `quantum-chain` network — syncing every
block from genesis and serving the `/gui` visualizer locally. It is a
**non-validating full node**: it holds no validator key and no stake, so there is
nothing to register and no reward to earn (daqq is a no-reward ledger). It simply
keeps its own verified copy of the ledger.

> **Participation is by application (allowlist).** daqq is a small, curated
> research network. The public sentry's peer-to-peer port is **not** open to the
> whole internet — only source IPs that have been approved can reach it. This
> keeps the network's attack surface minimal. To join you first request access,
> then connect. See [Request access](#request-access) below.

## Request access

Access is requested through the daqq Discord: **<https://discord.gg/pxrjYJKKF>**

1. Find your node's **public source IP** (the address you connect out from):

   ```sh
   curl -s https://api.ipify.org
   ```

2. Join the Discord and open the **`#run-a-node`** channel to say you'd like to
   run a node.

3. **Send your details by direct message to a `@core` operator** — do **not**
   post them in a public channel (the server rules ask you not to share node IP
   addresses publicly). In the DM include:
   - your **public IP** (from step 1),
   - a **moniker** (a name for your node),
   - optionally your **node ID** (`quantumchaind comet show-node-id`).

4. Once the operator confirms your IP is on the allowlist, continue below.

> Your IP must be reasonably stable. If your connection uses a dynamic or
> CGNAT-shared address, it may change and stop working — request access again
> with the new IP. (Behind the scenes the operator runs
> [`deploy/scripts/sentry-allowlist.sh`](scripts/sentry-allowlist.sh) to grant
> each approved IP.)

## What you need (provided here)

| Fact | Value |
| ---- | ----- |
| **Sentry peer** (dial this, once allowlisted) | `fe59695b49bbd872f60a0a32be1067c1db0861cf@133.88.118.11:26636` |
| **Genesis** | [`deploy/genesis.json`](genesis.json) — chain-id `quantum-chain`, sha256 `95e97b7cc61f807420059486ae2b07a3bf870e2e6c0368d3a612bce249492541` |
| **Code** | this repository's default branch |

> **Binary compatibility.** A full node must run the same on-chain state machine
> the network runs, or it halts on an app-hash mismatch. The published repo is
> state-machine-compatible with the live network — the only differences from the
> running binary are the embedded GUI page and the off-chain `qc-client`, neither
> of which affects consensus (identical `go.mod`, `app.go`, and `x/` modules).
> So building from this repo as-is is sufficient; you do not need a special
> commit.

## Option A — Docker / Apple container (no Go toolchain)

`deploy/docker/join-entrypoint.sh` installs the network genesis and dials the
sentry for you.

```sh
# from the repo root
docker build -f deploy/docker/Dockerfile -t daqq-quantumchaind .

docker run --rm \
  -e PERSISTENT_PEERS="fe59695b49bbd872f60a0a32be1067c1db0861cf@133.88.118.11:26636" \
  -e MONIKER="my-fullnode" \
  -v "$PWD/deploy/genesis.json:/genesis.json:ro" \
  -e GENESIS_SRC=/genesis.json \
  -p 1317:1317 -p 26657:26657 \
  --entrypoint /usr/local/bin/join-entrypoint.sh \
  daqq-quantumchaind
```

Apple `container` (macOS, Apple Silicon) is the same with `container run -d ...`
— see [`deploy/docker/README.md`](docker/README.md) for the `container` variant
and requirements.

## Option B — Native binary

```sh
# 1. build & install quantumchaind
cd quantum-chain && make install && cd ..

# 2. init a home (this throwaway genesis is overwritten in step 3)
quantumchaind init my-fullnode --chain-id quantum-chain --home "$HOME/.quantumchain"

# 3. install the network genesis
cp deploy/genesis.json "$HOME/.quantumchain/config/genesis.json"

# 4. point it at the sentry
CFG="$HOME/.quantumchain/config/config.toml"
sed -i.bak 's#^persistent_peers = .*#persistent_peers = "fe59695b49bbd872f60a0a32be1067c1db0861cf@133.88.118.11:26636"#' "$CFG"

# 5. sync
quantumchaind start --home "$HOME/.quantumchain"
```

## Verify you are syncing

```sh
# connected to the sentry (n_peers >= 1) — 0 usually means your IP is not (yet)
# on the allowlist, or it changed
curl -s localhost:26657/net_info | jq '.result.n_peers'

# catching_up flips true -> false as latest_block_height climbs to the tip
curl -s localhost:26657/status | jq '.result.sync_info'
```

Then open the local visualizer at <http://localhost:1317/gui/>.

## What this model does and doesn't hide

- Your node makes only an **outbound** connection to the sentry, so you never
  listen on a public port and never appear in any peer crawl.
- The sentry keeps the validators in `private_peer_ids`, so their addresses are
  never gossiped — the validators stay hidden behind the single public entry
  point, and there is no reachable network path from the internet to them.
- Because access is allowlisted, you **do** disclose your source IP to the
  operator when you request access. That is the trade-off of the permissioned
  model: the network sees far fewer strangers, at the cost of each participant
  identifying their IP to be let in.

## For network operators

The public sentry is produced by [`deploy/scripts/init-seed.sh`](scripts/init-seed.sh)
and kept alive by [`deploy/systemd/daqq-seed.service`](systemd/daqq-seed.service).
It is a non-validating full node that binds p2p to `0.0.0.0`, advertises its
public address, dials the validators over localhost, and lists them in
`private_peer_ids` so their addresses are never gossiped.

Access is controlled with [`deploy/scripts/sentry-allowlist.sh`](scripts/sentry-allowlist.sh):

```sh
sudo deploy/scripts/sentry-allowlist.sh add 203.0.113.9 "alice's laptop"
sudo deploy/scripts/sentry-allowlist.sh list
sudo deploy/scripts/sentry-allowlist.sh remove 203.0.113.9
```

The sentry port is **never** opened to `Anywhere`. The cloud firewall (e.g.
ConoHa security group) forwards the port to the host, and ufw allows only the
per-IP rules the script adds; everyone else is dropped.
