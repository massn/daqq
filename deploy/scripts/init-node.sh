#!/usr/bin/env bash
#
# Initialize a single daqq quantum-chain node home, idempotently, with all
# externally-reachable services locked to localhost. Only the embedded /gui
# visualizer is meant to be exposed (via Cloudflare Tunnel); RPC, gRPC and the
# non-GUI REST endpoints stay private.
#
# daqq is a no-reward ledger: no user-balance tokens are minted, only the
# minimum stake a PoS validator needs to participate.
#
# Idempotent: if the home is already initialized it only re-applies the
# bind/cors settings and exits. Safe to re-run.
#
# Usage:
#   deploy/scripts/init-node.sh
#   CHAIN_ID=quantum-chain HOME_DIR=$HOME/.quantumchain deploy/scripts/init-node.sh
#
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-quantum-chain}"
HOME_DIR="${HOME_DIR:-$HOME/.quantumchain}"
MONIKER="${MONIKER:-mynode}"
KEY="${KEY:-alice}"
APP=quantumchaind
KR=(--keyring-backend test)

command -v "$APP" >/dev/null 2>&1 || { echo "error: $APP not on PATH (run: cd quantum-chain && make install)" >&2; exit 1; }

say() { printf '\n>> %s\n' "$*"; }

# ---- 1. one-time genesis setup (skipped if already initialized) -----------
if [ -f "$HOME_DIR/config/genesis.json" ]; then
	say "home $HOME_DIR already initialized — skipping genesis setup"
else
	say "initializing $CHAIN_ID node at $HOME_DIR"
	"$APP" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null
	"$APP" keys add "$KEY" "${KR[@]}" --home "$HOME_DIR" >/dev/null
	ADDR="$("$APP" keys show "$KEY" -a "${KR[@]}" --home "$HOME_DIR")"
	# stake only — no user-balance tokens (no-reward ledger)
	"$APP" genesis add-genesis-account "$ADDR" 200000000stake --home "$HOME_DIR" >/dev/null
	"$APP" genesis gentx "$KEY" 100000000stake --chain-id "$CHAIN_ID" "${KR[@]}" --home "$HOME_DIR" >/dev/null
	"$APP" genesis collect-gentxs --home "$HOME_DIR" >/dev/null
	"$APP" genesis validate --home "$HOME_DIR" >/dev/null
fi

# ---- 2. lock binds to localhost + enable the API (idempotent) -------------
say "locking RPC/gRPC/API to localhost and enabling the API"
CFG="$HOME_DIR/config/config.toml"
APPTOML="$HOME_DIR/config/app.toml"

# config.toml: CometBFT RPC -> localhost
sed_i() { if sed --version >/dev/null 2>&1; then sed -i "$@"; else sed -i '' "$@"; fi; }
sed_i 's#^laddr = "tcp://[^"]*:26657"#laddr = "tcp://127.0.0.1:26657"#' "$CFG"

# app.toml: minimum-gas-prices, [api] enable+address, [grpc] address — section-aware
tmp="$(mktemp)"
awk '
	/^\[/ { sec = $0 }
	/^minimum-gas-prices = / && sec == "" { print "minimum-gas-prices = \"0stake\""; next }
	sec == "[api]"  && /^enable = /  { print "enable = true"; next }
	sec == "[api]"  && /^address = / { print "address = \"tcp://127.0.0.1:1317\""; next }
	sec == "[grpc]" && /^address = / { print "address = \"127.0.0.1:9090\""; next }
	{ print }
' "$APPTOML" >"$tmp" && mv "$tmp" "$APPTOML"

say "done. start with:  $APP start --home $HOME_DIR"
echo "   (or run it under cosmovisor — see deploy/scripts/setup-cosmovisor.sh)"
echo "   GUI will be served at http://127.0.0.1:1317/gui/"
