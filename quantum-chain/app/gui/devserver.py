#!/usr/bin/env python3
"""Mock dev server for iterating on the quantum-chain GUI without a live chain.

Serves the real index.html (re-read from disk on every request, so edits show
up on a plain browser reload) plus stubbed /gui/net_info and /gui/seeds
endpoints with rich mock data. The block height advances over time so the
round-progress card animates just like it does against a real node.

The mock reports several peers, so it exercises the multi-node network graph;
to see the single-node ("Running solo") view, run a real node instead
(`task quickstart`) and open http://localhost:1317/gui/.

Usage:
    python3 quantum-chain/app/gui/devserver.py        # serve on :8787
    task gui:preview                                  # same, via Taskfile
"""
import json
import os
import time
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path

PORT = int(os.environ.get("GUI_PREVIEW_PORT", "8787"))
INDEX = Path(__file__).resolve().parent / "index.html"
START = time.time()
BASE_HEIGHT = 1234


def current_height():
    # ~1 block every 2 seconds, so round_id (height // 50) ticks and the
    # commit/reveal/final phases visibly progress.
    return BASE_HEIGHT + int((time.time() - START) // 2)


def net_info():
    h = current_height()
    return {
        "self": {
            "id": "a1b2c3d4e5f600000000000000000000deadbeef",
            "moniker": "alice",
            "height": str(h),
        },
        "peers": [
            {"id": "f00ba711111111111111111111111111cafef00d", "moniker": "bob",
             "ip": "10.0.0.12", "outbound": True},
            {"id": "beef2222222222222222222222222222feedface", "moniker": "carol",
             "ip": "10.0.0.13", "outbound": True},
            {"id": "aaaa3333333333333333333333333333bbbbcccc", "moniker": "dave",
             "ip": "192.168.1.40", "outbound": False},
        ],
        "n_peers": 3,
    }


def seeds():
    latest_round = current_height() // 50
    out = []
    for i in range(6):
        r = latest_round - i
        if r < 0:
            break
        out.append({"round_id": r, "seed": f"{(r * 2654435761) & 0xffffffffffffffff:016x}"})
    return {"seeds": out}


def problems():
    return {
        "problems": [
            {"id": 1, "name": "random_circuit", "enabled": True,
             "description": "Theoretical output distribution of a random quantum "
                            "circuit seeded by the per-round beacon."},
            {"id": 2, "name": "random_circuit_sampling", "enabled": False,
             "description": "Sampling histograms (case B). Reserved for the future."},
        ]
    }


class Handler(BaseHTTPRequestHandler):
    def _json(self, payload):
        body = json.dumps(payload).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        path = self.path.split("?")[0]
        if path.endswith("/net_info"):
            return self._json(net_info())
        if path.endswith("/seeds"):
            return self._json(seeds())
        if path.endswith("/problems"):
            return self._json(problems())
        # everything else -> the live index.html from disk
        try:
            html = INDEX.read_bytes()
        except OSError as e:
            self.send_error(500, str(e))
            return
        self.send_response(200)
        self.send_header("Content-Type", "text/html; charset=utf-8")
        self.send_header("Content-Length", str(len(html)))
        self.end_headers()
        self.wfile.write(html)

    def log_message(self, *args):
        pass  # quiet


if __name__ == "__main__":
    print(f"GUI mock dev server on http://localhost:{PORT}/  (index: {INDEX})")
    ThreadingHTTPServer(("127.0.0.1", PORT), Handler).serve_forever()
