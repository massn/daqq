#!/usr/bin/env bash
#
# sentry-watchdog.sh — liveness check + self-heal for the daqq sentry node.
#
# The sentry occasionally exits block sync early and switches to consensus mode,
# where it can no longer backfill the gap: the process stays up and even reports
# sync_info.catching_up=false, yet the block height stops advancing. systemd's
# Restart= never fires because the process didn't die. Restarting it makes it
# re-enter block sync and catch up. This watchdog detects that "alive but stuck"
# state and restarts the service.
#
# Unlike deploy/scripts/healthcheck.sh (validators, which expose the REST /gui
# API), the sentry disables API/gRPC, so this reads the CometBFT RPC /status
# height instead.
#
# Run periodically via deploy/systemd/daqq-sentry-watchdog.timer. Side-effect
# free unless a restart is warranted.
#
# Usage:
#   deploy/scripts/sentry-watchdog.sh
#   RPC_URL=http://127.0.0.1:26637 SERVICE=daqq-seed deploy/scripts/sentry-watchdog.sh
set -euo pipefail

RPC_URL="${RPC_URL:-http://127.0.0.1:26637}"
SERVICE="${SERVICE:-daqq-seed}"
STATE_FILE="${STATE_FILE:-/tmp/daqq-sentry-watchdog.height}"

ts() { date '+%Y-%m-%dT%H:%M:%S%z'; }
log() { echo "$(ts) sentry-watchdog: $*"; }

restart() {
	log "UNHEALTHY ($1) -> restarting $SERVICE"
	if command -v systemctl >/dev/null 2>&1; then
		systemctl restart "$SERVICE"
	else
		log "systemctl not available; cannot auto-restart" >&2
	fi
	rm -f "$STATE_FILE"
	exit 1
}

# 1. Does the RPC answer with a numeric height?
HEIGHT="$(curl -s "$RPC_URL/status" 2>/dev/null | jq -r '.result.sync_info.latest_block_height // empty' 2>/dev/null || true)"
if ! [ "$HEIGHT" -ge 1 ] 2>/dev/null; then
	restart "could not read a numeric block height from $RPC_URL/status (got '${HEIGHT:-empty}')"
fi

# 2. Has the height advanced since the previous run?
PREV=""
[ -f "$STATE_FILE" ] && PREV="$(cat "$STATE_FILE" 2>/dev/null || true)"
echo "$HEIGHT" >"$STATE_FILE"

if [ -n "$PREV" ] && [ "$HEIGHT" -le "$PREV" ] 2>/dev/null; then
	restart "block height stuck at $HEIGHT (was $PREV last run)"
fi

log "OK (height $HEIGHT${PREV:+, prev $PREV})"
