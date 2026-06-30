#!/usr/bin/env bash
#
# Snapshot the node's read-only /gui data into static JSON for the Cloudflare
# Pages site (Epic C). Fetches localhost /gui/*, strips peer IPs / node IDs, and
# writes <OUT>/data/*.json. The static index.html (deploy/pages/index.html) reads
# these files, so the node never exposes a live inbound endpoint to the public —
# it only pushes a sanitized snapshot.
#
# Usage:
#   API=http://localhost:1317 OUT=deploy/pages ./deploy/scripts/gui-snapshot.sh
#   DEPLOY=1 PROJECT=daqq ... ./deploy/scripts/gui-snapshot.sh   # also deploy to Pages
#
# Env: API (node REST, default http://localhost:1317), OUT (pages dir),
#      DEPLOY (1 to run `wrangler pages deploy`), PROJECT (Pages project name).
set -euo pipefail

API="${API:-http://localhost:1317}"
OUT="${OUT:-deploy/pages}"
PROJECT="${PROJECT:-daqq}"
mkdir -p "$OUT/data"
ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# read-only data surfaces (no peer details here)
curl -fsS "$API/gui/seeds"    >"$OUT/data/seeds.json"
curl -fsS "$API/gui/problems" >"$OUT/data/problems.json"
curl -fsS "$API/gui/results"  >"$OUT/data/results.json"

# net_info: drop peer IPs / node IDs and the self node id; keep only counts,
# monikers and height plus a snapshot timestamp.
curl -fsS "$API/gui/net_info" | jq --arg ts "$ts" '{
  self: { moniker: .self.moniker, height: .self.height },
  peers: [ .peers[]? | { moniker: .moniker, outbound: .outbound } ],
  n_peers: .n_peers,
  snapshot_at: $ts
}' >"$OUT/data/net_info.json"

echo ">> snapshot written to $OUT/data (at $ts)"

if [ "${DEPLOY:-0}" = "1" ]; then
  echo ">> deploying to Cloudflare Pages (project: $PROJECT)"
  npx --yes wrangler pages deploy "$OUT" --project-name "$PROJECT" --commit-dirty=true
fi
