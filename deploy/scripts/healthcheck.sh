#!/usr/bin/env bash
#
# Liveness check + self-heal for the daqq node. Confirms the local REST API
# answers AND that the block height has advanced since the previous run; if the
# chain looks stuck (API down, or height not moving), it restarts the systemd
# service. systemd's own Restart= handles process crashes; this catches the
# subtler "process alive but not producing blocks" case.
#
# Run it periodically via the timer in deploy/systemd/ (see bottom of this file)
# or cron. Idempotent and side-effect-free unless a restart is warranted.
#
# Usage:
#   deploy/scripts/healthcheck.sh
#   API_URL=http://127.0.0.1:1317 SERVICE=quantumchaind deploy/scripts/healthcheck.sh
#
set -euo pipefail

API_URL="${API_URL:-http://127.0.0.1:1317}"
SERVICE="${SERVICE:-quantumchaind}"
STATE_FILE="${STATE_FILE:-/tmp/daqq-healthcheck.height}"

ts() { date '+%Y-%m-%dT%H:%M:%S%z'; }
log() { echo "$(ts) healthcheck: $*"; }

restart() {
	log "UNHEALTHY ($1) -> restarting $SERVICE"
	if command -v systemctl >/dev/null 2>&1; then
		sudo systemctl restart "$SERVICE"
	else
		log "systemctl not available; cannot auto-restart" >&2
	fi
	rm -f "$STATE_FILE"
	exit 1
}

# 1. Is the GUI/REST endpoint answering at all?
SEEDS_HTTP="$(curl -s -o /dev/null -w '%{http_code}' "$API_URL/gui/seeds" 2>/dev/null || true)"
if [ "$SEEDS_HTTP" != "200" ]; then
	restart "GET /gui/seeds returned '${SEEDS_HTTP:-no-response}'"
fi

# 2. Is the chain still advancing? Compare height to the previous run.
HEIGHT="$(curl -s "$API_URL/gui/net_info" 2>/dev/null | jq -r '.self.height // empty' 2>/dev/null || true)"
if ! [ "$HEIGHT" -ge 1 ] 2>/dev/null; then
	restart "could not read a numeric block height (got '${HEIGHT:-empty}')"
fi

PREV=""
[ -f "$STATE_FILE" ] && PREV="$(cat "$STATE_FILE" 2>/dev/null || true)"
echo "$HEIGHT" >"$STATE_FILE"

if [ -n "$PREV" ] && [ "$HEIGHT" -le "$PREV" ] 2>/dev/null; then
	restart "block height stuck at $HEIGHT (was $PREV last run)"
fi

log "OK (height $HEIGHT${PREV:+, prev $PREV})"

# ---------------------------------------------------------------------------
# Run every 5 minutes via systemd. Install alongside the node:
#
#   # /etc/systemd/system/daqq-healthcheck.service
#   [Unit]
#   Description=daqq node healthcheck
#   [Service]
#   Type=oneshot
#   User=ubuntu
#   ExecStart=/home/ubuntu/daqq/deploy/scripts/healthcheck.sh
#
#   # /etc/systemd/system/daqq-healthcheck.timer
#   [Unit]
#   Description=Run daqq healthcheck every 5 minutes
#   [Timer]
#   OnBootSec=2min
#   OnUnitActiveSec=5min
#   [Install]
#   WantedBy=timers.target
#
#   sudo systemctl daemon-reload && sudo systemctl enable --now daqq-healthcheck.timer
#
# The restart path needs passwordless sudo for `systemctl restart $SERVICE`
# (or run the unit as root).
# ---------------------------------------------------------------------------
