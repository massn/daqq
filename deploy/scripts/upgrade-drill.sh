#!/usr/bin/env bash
#
# Local hot-swap rehearsal for the daqq quantum-chain node.
#
# Proves the full Cosmovisor binary swap end to end on a single local node:
#
#   1. genesis binary = a build that does NOT know the "v1-1" upgrade handler
#      (auto-built from the commit before app/upgrades.go was added).
#   2. upgrade binary = the current HEAD build, which DOES handle "v1-1".
#   3. a governance software-upgrade proposal for "v1-1" is submitted and voted in
#      with a short voting period.
#   4. at the planned height the genesis binary halts ("UPGRADE NEEDED"); Cosmovisor
#      swaps to the upgrade binary and restarts; the chain resumes past that height.
#
# Everything runs in a throwaway home; it never touches ~/.quantumchain. Fully
# local — no Oracle Cloud, no real funds.
#
# Usage:
#   deploy/scripts/upgrade-drill.sh
#   DRILL_GENESIS_REF=<git-ref> deploy/scripts/upgrade-drill.sh
#
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHAIN_ID="daqq-drill"
DRILL_HOME="${DRILL_HOME:-/tmp/daqq-upgrade-drill}"
UPGRADE_NAME="v1-1"
APP=quantumchaind
KR=(--keyring-backend test)
LOG="$DRILL_HOME/cosmovisor.log"
WT="$DRILL_HOME/genesis-src"

export DAEMON_NAME="$APP"
export DAEMON_HOME="$DRILL_HOME"
export DAEMON_ALLOW_DOWNLOAD_BINARIES=false
export DAEMON_RESTART_AFTER_UPGRADE=true

COSMOVISOR_PID=""
cleanup() {
	[ -n "$COSMOVISOR_PID" ] && kill "$COSMOVISOR_PID" 2>/dev/null || true
	pkill -f "cosmovisor run start --home $DRILL_HOME" 2>/dev/null || true
	pkill -f "$DRILL_HOME/cosmovisor" 2>/dev/null || true
}
trap cleanup EXIT

say() { printf '\n>> %s\n' "$*"; }
# Tolerant against the node not being up yet (pipefail + set -e would otherwise
# abort the whole script when curl can't connect during the poll loops).
rpc_height() { { curl -s localhost:26657/status 2>/dev/null | jq -r '.result.sync_info.latest_block_height // empty' 2>/dev/null; } || true; }

# ---- 0. clean slate -------------------------------------------------------
pkill -f "cosmovisor run start --home $DRILL_HOME" 2>/dev/null || true
rm -rf "$DRILL_HOME"
mkdir -p "$DRILL_HOME"

# ---- 1. build the two binaries -------------------------------------------
say "building upgrade (HEAD) binary — has the $UPGRADE_NAME handler"
( cd "$REPO/quantum-chain" && go build -o "$DRILL_HOME/upgrade-$APP" ./cmd/$APP )

# The genesis binary must NOT know the v1-1 handler. The chain entrypoint
# (cmd/quantumchaind) is not tracked in git, so we cannot use `git worktree`;
# instead copy the on-disk source and strip the upgrade-handler wiring.
say "building genesis binary (handler wiring removed) — must NOT know $UPGRADE_NAME"
# Anchored excludes (leading /) so we drop only the root-level built binaries,
# NOT the cmd/quantumchaind source directory.
rsync -a --delete \
	--exclude '/.git' --exclude '*.log' \
	--exclude '/quantumchaind' --exclude '/quantum-chaind' --exclude '/qc-client' \
	"$REPO/quantum-chain/" "$WT/"
rm -f "$WT/app/upgrades.go"
if sed --version >/dev/null 2>&1; then
	sed -i '/app.setupUpgradeHandlers()/d; /app.setupStoreLoaders()/d' "$WT/app/app.go"
else
	sed -i '' '/app.setupUpgradeHandlers()/d; /app.setupStoreLoaders()/d' "$WT/app/app.go"
fi
# Sanity: the stripped source must no longer wire the handler.
if grep -q "setupUpgradeHandlers" "$WT/app/app.go"; then
	echo "error: failed to strip upgrade-handler wiring from genesis source" >&2
	exit 1
fi
( cd "$WT" && go build -o "$DRILL_HOME/genesis-$APP" ./cmd/$APP )
GENBIN="$DRILL_HOME/genesis-$APP"
UPBIN="$DRILL_HOME/upgrade-$APP"

# ---- 2. install cosmovisor -------------------------------------------------
# cosmovisor pulls an old bytedance/sonic that fails to link under Go 1.24
# (encoding/json.unquoteBytes was removed); build it with an older toolchain.
COSMOVISOR_VERSION="${COSMOVISOR_VERSION:-v1.7.1}"
COSMOVISOR_TOOLCHAIN="${COSMOVISOR_TOOLCHAIN:-go1.23.6}"
if ! command -v cosmovisor >/dev/null 2>&1; then
	say "installing cosmovisor ($COSMOVISOR_VERSION via $COSMOVISOR_TOOLCHAIN)"
	GOTOOLCHAIN="$COSMOVISOR_TOOLCHAIN" go install "cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@${COSMOVISOR_VERSION}"
fi
command -v cosmovisor >/dev/null 2>&1 || { echo "error: cosmovisor not on PATH after install" >&2; exit 1; }

# ---- 3. init the chain with the genesis binary ----------------------------
say "initializing throwaway chain at $DRILL_HOME"
"$GENBIN" init drill --chain-id "$CHAIN_ID" --home "$DRILL_HOME" >/dev/null 2>&1
"$GENBIN" keys add val "${KR[@]}" --home "$DRILL_HOME" >/dev/null 2>&1
VAL_ADDR="$("$GENBIN" keys show val -a "${KR[@]}" --home "$DRILL_HOME")"
"$GENBIN" genesis add-genesis-account "$VAL_ADDR" 1000000000stake --home "$DRILL_HOME" >/dev/null 2>&1
"$GENBIN" genesis gentx val 700000000stake --chain-id "$CHAIN_ID" "${KR[@]}" --home "$DRILL_HOME" >/dev/null 2>&1
"$GENBIN" genesis collect-gentxs --home "$DRILL_HOME" >/dev/null 2>&1

# fast blocks + short governance so the drill finishes in a couple of minutes
G="$DRILL_HOME/config/genesis.json"
tmp="$(mktemp)"
jq '.app_state.gov.params.voting_period="20s"
  | .app_state.gov.params.max_deposit_period="20s"
  | .app_state.gov.params.expedited_voting_period="10s"
  | .app_state.gov.params.min_deposit=[{"denom":"stake","amount":"1"}]
  | .app_state.gov.params.expedited_min_deposit=[{"denom":"stake","amount":"1"}]' \
  "$G" >"$tmp" && mv "$tmp" "$G"
sed -i '' 's/^timeout_commit = .*/timeout_commit = "1s"/' "$DRILL_HOME/config/config.toml" 2>/dev/null \
  || sed -i 's/^timeout_commit = .*/timeout_commit = "1s"/' "$DRILL_HOME/config/config.toml"
# the node refuses to start with an empty minimum-gas-prices
sed -i '' 's/^minimum-gas-prices = .*/minimum-gas-prices = "0stake"/' "$DRILL_HOME/config/app.toml" 2>/dev/null \
  || sed -i 's/^minimum-gas-prices = .*/minimum-gas-prices = "0stake"/' "$DRILL_HOME/config/app.toml"

# ---- 4. lay out cosmovisor: genesis binary + staged upgrade binary --------
say "staging cosmovisor (genesis + upgrades/$UPGRADE_NAME)"
cosmovisor init "$GENBIN" >/dev/null 2>&1
mkdir -p "$DRILL_HOME/cosmovisor/upgrades/$UPGRADE_NAME/bin"
cp "$UPBIN" "$DRILL_HOME/cosmovisor/upgrades/$UPGRADE_NAME/bin/$APP"

# ---- 5. start the node through cosmovisor ---------------------------------
say "starting node under cosmovisor (log: $LOG)"
cosmovisor run start --home "$DRILL_HOME" >"$LOG" 2>&1 &
COSMOVISOR_PID=$!

say "waiting for first block"
for _ in $(seq 1 90); do
	h="$(rpc_height)"
	if [ -n "$h" ] && [ "$h" -ge 1 ] 2>/dev/null; then break; fi
	sleep 1
done
h0="$(rpc_height)"
if [ -z "$h0" ]; then echo "error: node never produced a block" >&2; tail -30 "$LOG" >&2; exit 1; fi
UPGRADE_HEIGHT=$((h0 + 35))
say "node live at height $h0; scheduling upgrade '$UPGRADE_NAME' for height $UPGRADE_HEIGHT"

# ---- 6. submit + vote the software-upgrade proposal -----------------------
GOV_ADDR="$("$UPBIN" q auth module-account gov -o json --home "$DRILL_HOME" 2>/dev/null | jq -r '.account.value.address // .account.address // empty')"
[ -n "$GOV_ADDR" ] || { echo "error: could not resolve gov module address" >&2; exit 1; }

PROP="$DRILL_HOME/proposal.json"
cat >"$PROP" <<JSON
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "$GOV_ADDR",
      "plan": { "name": "$UPGRADE_NAME", "height": "$UPGRADE_HEIGHT", "info": "drill" }
    }
  ],
  "metadata": "drill",
  "deposit": "10000000stake",
  "title": "$UPGRADE_NAME hot-swap drill",
  "summary": "no-op protocol bump to rehearse the cosmovisor swap"
}
JSON

say "submitting proposal"
"$UPBIN" tx gov submit-proposal "$PROP" --from val "${KR[@]}" --chain-id "$CHAIN_ID" \
	--home "$DRILL_HOME" --gas auto --gas-adjustment 1.5 --gas-prices 0stake -y >/dev/null 2>&1
sleep 4
say "voting yes"
"$UPBIN" tx gov vote 1 yes --from val "${KR[@]}" --chain-id "$CHAIN_ID" \
	--home "$DRILL_HOME" --gas auto --gas-adjustment 1.5 --gas-prices 0stake -y >/dev/null 2>&1

say "waiting for proposal to pass"
st=""
for _ in $(seq 1 50); do
	st="$({ "$UPBIN" q gov proposal 1 -o json --home "$DRILL_HOME" 2>/dev/null | jq -r '.proposal.status // empty' 2>/dev/null; } || true)"
	if [ "$st" = "PROPOSAL_STATUS_PASSED" ]; then break; fi
	if [ "$st" = "PROPOSAL_STATUS_REJECTED" ] || [ "$st" = "PROPOSAL_STATUS_FAILED" ]; then
		echo "error: proposal $st" >&2; exit 1
	fi
	sleep 1
done
say "proposal status: ${st:-unknown}"

# ---- 7. wait for the swap and verify the chain survived -------------------
say "waiting for cosmovisor to swap at height $UPGRADE_HEIGHT and resume"
target=$((UPGRADE_HEIGHT + 3))
for _ in $(seq 1 120); do
	h="$(rpc_height)"
	if [ -n "$h" ] && [ "$h" -ge "$target" ] 2>/dev/null; then break; fi
	sleep 1
done
hf="$(rpc_height)"

echo
echo "================ DRILL RESULT ================"
SWAP_OK=no; ALIVE=no
if grep -qiE "upgrade.*$UPGRADE_NAME|pre-upgrade|applying upgrade|halt application|upgrade-info" "$LOG"; then SWAP_OK=yes; fi
if [ -n "$hf" ] && [ "$hf" -gt "$UPGRADE_HEIGHT" ] 2>/dev/null; then ALIVE=yes; fi
echo "upgrade height      : $UPGRADE_HEIGHT"
echo "final height        : ${hf:-?}"
echo "chain alive past H  : $ALIVE"
echo "cosmovisor swapped  : $SWAP_OK"
echo "cosmovisor swap log :"
grep -iE "upgrade|swap|moniker" "$LOG" | tail -8 | sed 's/^/    /'
echo "=============================================="

if [ "$ALIVE" = yes ] && [ "$SWAP_OK" = yes ]; then
	echo "DRILL PASSED: governance upgrade '$UPGRADE_NAME' hot-swapped with no downtime."
	exit 0
fi
echo "DRILL FAILED — see $LOG" >&2
exit 1
