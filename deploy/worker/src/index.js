// daqq data relay Worker.
//
// POST /ingest   (Authorization: Bearer <INGEST_TOKEN>)  — ConoHa pushes the
//                combined snapshot {seeds, problems, results, net_info}; stored
//                as a single KV value so one push = one KV write.
// GET  /gui/<kind>                                        — the Pages GUI reads
//                back one slice (seeds|problems|results|net_info) with CORS.
//
// Keeping the node behind this relay means it never exposes a public inbound
// endpoint: it only makes outbound HTTPS POSTs to the Worker.

const KINDS = ['seeds', 'problems', 'results', 'net_info'];
const KEY = 'gui';

const CORS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
  'Access-Control-Allow-Headers': 'Authorization, Content-Type',
};

const json = (obj, extra = {}) =>
  new Response(JSON.stringify(obj), {
    headers: { 'Content-Type': 'application/json', ...CORS, ...extra },
  });

// shape returned when no snapshot has been ingested yet, per kind
const empty = (kind) =>
  kind === 'seeds' ? { seeds: [] }
  : kind === 'net_info' ? { self: {}, peers: [], n_peers: 0 }
  : {};

export default {
  async fetch(req, env) {
    const url = new URL(req.url);

    if (req.method === 'OPTIONS') return new Response(null, { headers: CORS });

    // --- ingest (ConoHa -> Worker) ---
    if (req.method === 'POST' && url.pathname === '/ingest') {
      const expected = `Bearer ${env.INGEST_TOKEN}`;
      if (!env.INGEST_TOKEN || (req.headers.get('Authorization') || '') !== expected)
        return json({ error: 'unauthorized' }, { 'x-status': '401' });
      let data;
      try { data = JSON.parse(await req.text()); }
      catch { return new Response('bad json', { status: 400, headers: CORS }); }
      if (typeof data !== 'object' || data === null)
        return new Response('bad shape', { status: 400, headers: CORS });
      data.ingested_at = new Date().toISOString();
      await env.DAQQ_DATA.put(KEY, JSON.stringify(data));
      return new Response('ok', { headers: CORS });
    }

    // --- serve (Worker -> GUI) ---
    // Serve any key present in the stored snapshot (seeds/problems/results/
    // net_info/solution/…), so new data kinds need no Worker change.
    const m = url.pathname.match(/^\/gui\/([a-z_]+)$/);
    if (req.method === 'GET' && m) {
      const kind = m[1];
      const blob = await env.DAQQ_DATA.get(KEY);
      const all = blob ? JSON.parse(blob) : {};
      // short edge cache; the GUI also cache-busts, so this mainly protects KV
      const headers = { 'Cache-Control': 'public, max-age=5' };
      if (kind in all) return json(all[kind] ?? empty(kind), headers);
      if (KINDS.includes(kind)) return json(empty(kind), headers);
      return new Response('not found', { status: 404, headers: CORS });
    }

    if (req.method === 'GET' && url.pathname === '/') {
      const blob = await env.DAQQ_DATA.get(KEY);
      const at = blob ? (JSON.parse(blob).ingested_at || '?') : 'no data yet';
      return json({ ok: true, service: 'daqq-data', last_ingest: at });
    }

    return new Response('not found', { status: 404, headers: CORS });
  },
};
