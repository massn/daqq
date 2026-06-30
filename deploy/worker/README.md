# daqq-data Worker (data relay)

Decouples "the node produces data" from "Pages serves the site". The node POSTs a
sanitized snapshot to this Worker; the Worker keeps it in KV and serves it to the
public GUI with CORS. The static site (docs + GUI) is deployed to Pages **once**;
data updates flow through the Worker with no Pages redeploy.

```
ConoHa node ──curl POST /ingest──▶ Worker (KV) ──GET /gui/*──▶ Pages GUI (fetch + CORS)
  curl + jq only                   free *.workers.dev, stable URL      deployed once
```

## One-time setup (run from this directory, `deploy/worker/`)

You need an authenticated `wrangler` (the same one you use for Pages).

```bash
cd deploy/worker

# 1. create the KV namespace, copy the printed id into wrangler.toml (id = "...")
wrangler kv namespace create DAQQ_DATA

# 2. set the shared ingest token (paste the value generated for you)
wrangler secret put INGEST_TOKEN

# 3. deploy — prints the URL, e.g. https://daqq-data.<you>.workers.dev
wrangler deploy
```

Tell me the printed Worker URL. Then I:

- point the static GUI's data URLs at the Worker, rebuild the combined site, and
  hand you the one Pages redeploy command;
- install the matching `WORKER_URL` + `INGEST_TOKEN` on ConoHa (`/root/.daqq-data.env`)
  and enable the `daqq-data-push.timer` (pushes every 2 min).

## Sanity checks

```bash
curl https://daqq-data.<you>.workers.dev/            # {ok:true, last_ingest:...}
curl https://daqq-data.<you>.workers.dev/gui/seeds   # {seeds:[...]} once data flows
```

## Files

- `src/index.js` — the Worker (ingest + serve).
- `wrangler.toml` — name/KV binding (`INGEST_TOKEN` is a secret, not stored here).
- `../scripts/push-data.sh` — ConoHa side (snapshot + POST), curl/jq only.
- `../systemd/daqq-data-push.{service,timer}` + `daqq-data.env.example` — the 2-min push.
