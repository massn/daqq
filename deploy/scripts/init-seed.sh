#!/usr/bin/env bash
#
# init-seed.sh — set up the public sentry node for an existing daqq network.
#
# This is the single documented entry point approved participants dial to join:
# it is a non-validating full node that (a) binds its p2p port to the public
# interface, (b) relays blocks so newcomers can sync from it, and (c) shields the
# real validators — it dials them over localhost and never gossips their
# addresses (private_peer_ids), so exposing this node does NOT expose the
# validators' p2p.
#
# Access is by application, not open to the whole internet: the sentry port is
# admitted per source IP via deploy/scripts/sentry-allowlist.sh (ufw). This
# script configures the node; the allowlist script controls who may reach it.
#
# It differs from init-node.sh (bootstraps a brand-new chain) and matches
# join-entrypoint.sh (installs the network genesis instead of generating one),
# but additionally publishes p2p and pins the upstream validators as private.
#
# Idempotent: safe to re-run; it re-applies config on an existing home.
#
# Typical use, co-located on the ConoHa validator host (dials bob over
# localhost; bob already has allow_duplicate_ip = true so no validator change is
# needed):
#
#   PEERS="8315e979898de25610c5cb19c949cc6e2819f5c5@127.0.0.1:26666" \
#   PRIVATE_PEER_IDS="dadd29045541596177ee6f1929a27c3702e4701e,8315e979898de25610c5cb19c949cc6e2819f5c5" \
#   deploy/scripts/init-seed.sh
#
# Env (all optional except where noted):
#   HOME_DIR          seed node home        (default /root/.quantumchain-seed)
#   GENESIS_SRC       network genesis to install
#                                           (default /root/.quantumchain/config/genesis.json)
#   CHAIN_ID          must match genesis    (default quantum-chain)
#   MONIKER           this node's name      (default daqq-seed)
#   EXTERNAL_IP       public IP to advertise (default: auto-detected)
#   P2P_PORT          public p2p listen port (default 26636)
#   RPC_PORT          localhost RPC port    (default 26637)
#   PPROF_PORT        localhost pprof port  (default 6062)
#   PEERS             persistent_peers to dial for the chain — REQUIRED in
#                     practice (the upstream full node/validator to sync from),
#                     e.g. "<nodeID>@127.0.0.1:26666"
#   PRIVATE_PEER_IDS  comma-separated node IDs this seed must never gossip
#                     (the validators), so PEX cannot leak their addresses
set -euo pipefail

HOME_DIR="${HOME_DIR:-/root/.quantumchain-seed}"
GENESIS_SRC="${GENESIS_SRC:-/root/.quantumchain/config/genesis.json}"
CHAIN_ID="${CHAIN_ID:-quantum-chain}"
MONIKER="${MONIKER:-daqq-seed}"
P2P_PORT="${P2P_PORT:-26636}"
RPC_PORT="${RPC_PORT:-26637}"
PPROF_PORT="${PPROF_PORT:-6062}"
PEERS="${PEERS:-}"
PRIVATE_PEER_IDS="${PRIVATE_PEER_IDS:-}"
APP="${QC_BIN:-quantumchaind}"

say() { printf '\n>> %s\n' "$*"; }
die() { printf '\nerror: %s\n' "$*" >&2; exit 1; }

command -v "$APP" >/dev/null 2>&1 || die "$APP not on PATH (set QC_BIN or install it)"
[ -f "$GENESIS_SRC" ] || die "GENESIS_SRC='$GENESIS_SRC' not found (point it at the network's genesis.json)"
[ -n "$PEERS" ] || say "warning: PEERS is empty — the seed has no upstream to sync from; set PEERS=<nodeID>@host:port"

EXTERNAL_IP="${EXTERNAL_IP:-$(curl -fsS https://api.ipify.org 2>/dev/null || curl -fsS ifconfig.me 2>/dev/null || true)}"
[ -n "$EXTERNAL_IP" ] || die "could not auto-detect EXTERNAL_IP; set it explicitly (the public IP peers dial)"

CFG="$HOME_DIR/config/config.toml"
APPTOML="$HOME_DIR/config/app.toml"

# GNU (Debian/Ubuntu) or BSD (macOS) sed.
sed_i() { if sed --version >/dev/null 2>&1; then sed -i "$@"; else sed -i '' "$@"; fi; }

# ---- 1. one-time init (node key + config; genesis replaced below) ----------
if [ -f "$CFG" ]; then
	say "home $HOME_DIR already initialized — reapplying genesis/config"
else
	say "initializing seed node at $HOME_DIR (chain-id $CHAIN_ID)"
	"$APP" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null
fi

# ---- 2. install the network genesis (never generate a fresh one) -----------
say "installing network genesis from $GENESIS_SRC"
cp "$GENESIS_SRC" "$HOME_DIR/config/genesis.json"

# ---- 3. configure p2p to face the public + shield the validators -----------
say "publishing p2p on 0.0.0.0:$P2P_PORT (advertising $EXTERNAL_IP:$P2P_PORT), RPC on 127.0.0.1:$RPC_PORT"
tmp="$(mktemp)"
awk -v ext="$EXTERNAL_IP:$P2P_PORT" -v p2p="$P2P_PORT" -v rpc="$RPC_PORT" \
    -v pprof="$PPROF_PORT" -v peers="$PEERS" -v priv="$PRIVATE_PEER_IDS" '
	/^\[/ { sec = $0 }
	sec == "[rpc]" && /^laddr = /       { print "laddr = \"tcp://127.0.0.1:" rpc "\""; next }
	sec == "[rpc]" && /^pprof_laddr = / { print "pprof_laddr = \"localhost:" pprof "\""; next }
	sec == "[p2p]" && /^laddr = /              { print "laddr = \"tcp://0.0.0.0:" p2p "\""; next }
	sec == "[p2p]" && /^external_address = /   { print "external_address = \"" ext "\""; next }
	sec == "[p2p]" && /^persistent_peers = /   { print "persistent_peers = \"" peers "\""; next }
	sec == "[p2p]" && /^private_peer_ids = /   { print "private_peer_ids = \"" priv "\""; next }
	sec == "[p2p]" && /^pex = /                { print "pex = true"; next }
	sec == "[p2p]" && /^seed_mode = /          { print "seed_mode = false"; next }
	sec == "[p2p]" && /^allow_duplicate_ip = / { print "allow_duplicate_ip = true"; next }
	{ print }
' "$CFG" >"$tmp" && mv "$tmp" "$CFG"

# ---- 4. app.toml: no-reward gas, and keep API/gRPC off + off the default ---
#         ports so a co-located seed never clashes with the validators.
say "disabling API/gRPC (seed only needs p2p) and setting zero gas"
tmp="$(mktemp)"
awk '
	/^\[/ { sec = $0 }
	/^minimum-gas-prices = / && sec == ""      { print "minimum-gas-prices = \"0stake\""; next }
	sec == "[api]"  && /^enable = /  { print "enable = false"; next }
	sec == "[api]"  && /^address = / { print "address = \"tcp://127.0.0.1:1319\""; next }
	sec == "[grpc]" && /^enable = /  { print "enable = false"; next }
	sec == "[grpc]" && /^address = / { print "address = \"127.0.0.1:9092\""; next }
	{ print }
' "$APPTOML" >"$tmp" && mv "$tmp" "$APPTOML"

NODE_ID="$("$APP" comet show-node-id --home "$HOME_DIR" 2>/dev/null || echo '<run: quantumchaind comet show-node-id>')"

say "done. seed identity for the join docs:"
echo "   $NODE_ID@$EXTERNAL_IP:$P2P_PORT"
echo
echo "   start it:      $APP start --home $HOME_DIR"
echo "   or as a unit:  see deploy/systemd/daqq-seed.service"
echo "   admit a peer:  sudo deploy/scripts/sentry-allowlist.sh add <applicant-IP>"
echo "   (do NOT open $P2P_PORT to Anywhere — access is by application/allowlist)"
