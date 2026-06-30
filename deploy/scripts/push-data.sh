#!/usr/bin/env bash
#
# push-data.sh — ConoHa side of the Worker relay (Epic C, lightweight variant).
#
# Snapshots this node's read-only /gui surfaces, sanitizes net_info (drops peer
# IPs / node IDs), bundles all four into one JSON object, and POSTs it to the
# daqq-data Worker's /ingest endpoint. The Worker stores it in KV and serves it
# to the public GUI. Requires only curl + jq — no node, hugo, or wrangler here.
#
# Env (from the systemd EnvironmentFile /root/.daqq-data.env):
#   WORKER_URL    e.g. https://daqq-data.<you>.workers.dev   (no trailing slash)
#   INGEST_TOKEN  shared bearer token (matches the Worker secret)
#   API           node REST base, default http://localhost:1317
set -euo pipefail

API="${API:-http://localhost:1317}"
RPC="${RPC:-http://localhost:26657}"
QC_BIN="${QC_BIN:-/root/go/bin/quantumchaind}"
: "${WORKER_URL:?set WORKER_URL (see deploy/systemd/daqq-data.env.example)}"
: "${INGEST_TOKEN:?set INGEST_TOKEN (see deploy/systemd/daqq-data.env.example)}"

ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# fetch each surface to a file (results can be large — avoid arg-length limits)
curl -fsS "$API/gui/seeds"    >"$tmp/seeds.json"
curl -fsS "$API/gui/problems" >"$tmp/problems.json"
curl -fsS "$API/gui/results"  >"$tmp/results.json"
# net_info: keep only monikers / counts / height + a snapshot timestamp
curl -fsS "$API/gui/net_info" | jq --arg ts "$ts" '{
  self: { moniker: .self.moniker, height: .self.height },
  peers: [ .peers[]? | { moniker: .moniker, outbound: .outbound } ],
  n_peers: .n_peers,
  snapshot_at: $ts
}' >"$tmp/net.json"

# Recent solved circuits: pull the full output distributions (the actual answers,
# not just hashes) for the last SOL_KEEP rounds from on-chain submit-result txs,
# PER submitting node: { round_id: { creator_address: entries } }. `q txs`
# returns oldest-first, so read the last page(s).
SOL_KEEP="${SOL_KEEP:-30}"
Q="message.action='/quantumchain.random_circuit.v1.MsgSubmitResult'"
echo '{}' >"$tmp/submissions.json"
PT="$("$QC_BIN" q txs --query "$Q" --limit 50 --page 1 --node "$RPC" -o json 2>/dev/null | jq -r '.page_total // 1' || echo 1)"
{
  "$QC_BIN" q txs --query "$Q" --limit 50 --page "$PT" --node "$RPC" -o json 2>/dev/null
  [ "${PT:-1}" -gt 1 ] && "$QC_BIN" q txs --query "$Q" --limit 50 --page "$((PT-1))" --node "$RPC" -o json 2>/dev/null
} | jq -s -c \
  '[ .[].txs[]? as $tx | $tx.tx.body.messages[]? | select(.["@type"] | test("SubmitResult"))
     | { round: (.round_id | tonumber? // (.round_id|tostring|tonumber)),
         addr: .creator,
         rec: { entries: (.distribution.entries // []),
                height: ($tx.height | tonumber?), time: $tx.timestamp } } ]
   | group_by(.round) | sort_by(.[0].round) | reverse | .[0:'"$SOL_KEEP"']
   | reduce .[] as $g ({}; .[($g[0].round|tostring)] = (reduce $g[] as $x ({}; .[$x.addr] = $x.rec)))' \
  >"$tmp/submissions.json" 2>/dev/null || echo '{}' >"$tmp/submissions.json"
# agreed/first distribution per round (default display + backward compat)
jq -c 'with_entries(.value = ((.value | to_entries[0].value.entries) // []))' \
  "$tmp/submissions.json" >"$tmp/solutions.json" 2>/dev/null || echo '{}' >"$tmp/solutions.json"
# latest single (backward compat with /gui/solution)
jq -c 'to_entries | (max_by(.key|tonumber) // {key:null,value:[]})
       | {round_id:(.key|tonumber?), entries:.value}' \
  "$tmp/solutions.json" >"$tmp/solution.json" 2>/dev/null || echo '{}' >"$tmp/solution.json"

# validator monikers: map each operator's bech32 data part (shared with the
# submitter's account address) to its self-declared moniker, so the GUI can name
# which node submitted each result.
echo '{}' >"$tmp/monikers.json"
"$QC_BIN" q staking validators --node "$RPC" -o json 2>/dev/null \
  | jq -c '[.validators[] | {key: .operator_address[10:42], value: .description.moniker}] | from_entries' \
  >"$tmp/monikers.json" 2>/dev/null || echo '{}' >"$tmp/monikers.json"

# bundle into one object via --slurpfile (reads files, no arg-length limit)
jq -n \
  --slurpfile s "$tmp/seeds.json" --slurpfile p "$tmp/problems.json" \
  --slurpfile r "$tmp/results.json" --slurpfile n "$tmp/net.json" \
  --slurpfile sol "$tmp/solution.json" --slurpfile sols "$tmp/solutions.json" \
  --slurpfile subm "$tmp/submissions.json" --slurpfile mon "$tmp/monikers.json" \
  '{seeds: $s[0], problems: $p[0], results: $r[0], net_info: $n[0], solution: $sol[0], solutions: $sols[0], submissions: $subm[0], monikers: $mon[0]}' >"$tmp/combined.json"

curl -fsS -X POST "$WORKER_URL/ingest" \
  -H "Authorization: Bearer $INGEST_TOKEN" \
  -H 'Content-Type: application/json' \
  --data @"$tmp/combined.json" >/dev/null

echo ">> pushed snapshot at $ts ($(wc -c <"$tmp/combined.json") bytes)"
