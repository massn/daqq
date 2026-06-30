#!/usr/bin/env python3
"""daqq Qiskit reference client (validator "carol").

A heterogeneous, independent re-implementation of the per-round random-circuit
problem in Python/Qiskit — to test whether a different language/stack agrees
byte-for-byte with the Go nodes (cross-validation across implementations).

It re-derives each round's circuit from the shared seed (FNV-1a seed fold +
XorShift32 gate selection + the H/S/T + brickwork-CNOT layers — identical to
circuit.MakeRandomQC in Go), computes the exact output distribution with
Qiskit's Statevector, encodes it the same way daqq does, and submits it.

Modes:
  verify <seed_hex> <onchain_hash>   compare our distribution hash to on-chain
  run                                 daemon: poll rounds, compute, submit (carol)
"""
import hashlib
import json
import os
import subprocess
import sys
import time
import urllib.request

WIDTH, DEPTH = 5, 10

# ---- environment (run mode) ----
API = os.environ.get("QC_API", "http://localhost:1317")
QC_BIN = os.environ.get("QC_BIN", "/root/go/bin/quantumchaind")
QC_RPC = os.environ.get("QC_RPC", "tcp://localhost:26677")
QC_FROM = os.environ.get("QC_FROM", "carol")
QC_HOME = os.environ.get("QC_HOME", "/root/.quantumchain3")
QC_KEYRING = os.environ.get("QC_KEYRING_BACKEND", "test")
QC_CHAIN = os.environ.get("QC_CHAIN_ID", "quantum-chain")


def seed_state(seed_bytes):
    """FNV-1a over the full seed -> 32-bit XorShift32 state (== circuit.SeedState)."""
    h = 2166136261
    for b in seed_bytes:
        h = ((h ^ b) * 16777619) & 0xFFFFFFFF
    return h


def make_layers(state):
    gs = ["H", "S", "T"]

    def nxt():
        nonlocal state
        state ^= (state << 13) & 0xFFFFFFFF
        state &= 0xFFFFFFFF
        state ^= state >> 17
        state &= 0xFFFFFFFF
        state ^= (state << 5) & 0xFFFFFFFF
        state &= 0xFFFFFFFF
        return gs[state % 3]

    layers = []
    for i in range(DEPTH):
        m = i % 4
        if m in (0, 2):
            layers.append(("single", [nxt() for _ in range(WIDTH)]))
        elif m == 1:
            layers.append(("cnot", [(2 * j, 2 * j + 1) for j in range(WIDTH // 2)]))
        else:
            cn, j = [], 0
            while 2 * j + 2 < WIDTH:
                cn.append((2 * j + 1, 2 * j + 2))
                j += 1
            layers.append(("cnot", cn))
    return layers


def build_circuit(seed_hex):
    from qiskit import QuantumCircuit
    st = seed_state(bytes.fromhex(seed_hex))
    qc = QuantumCircuit(WIDTH)
    for kind, data in make_layers(st):
        if kind == "single":
            for q, g in enumerate(data):
                getattr(qc, g.lower())(q)
        else:
            for c, t in data:
                qc.cx(c, t)
    return qc


def go_format(p):
    """Quantize to 9 decimals — matches Go circuit.Distribution (FormatFloat 'f',9)."""
    return "%.9f" % float(p)


def compute_distribution(seed_hex, reverse=False):
    """{state_string: prob_string} matching daqq's quantized encoding."""
    from qiskit.quantum_info import Statevector
    sv = Statevector(build_circuit(seed_hex))
    probs = sv.probabilities_dict()  # qiskit little-endian keys
    out = {}
    for k, p in probs.items():
        s = go_format(p)
        if s == "0.000000000":       # dropped like Go (rounds to zero)
            continue
        state = k[::-1] if reverse else k
        out[state] = s
    return out


def dist_hash(dist):
    s = "".join("%s=%s;" % (st, pr) for st, pr in sorted(dist.items()))
    return hashlib.sha256(s.encode()).hexdigest()


# ---------------- verify mode ----------------
def verify(seed_hex, onchain_hash):
    for rev in (False, True):
        d = compute_distribution(seed_hex, reverse=rev)
        h = dist_hash(d)
        print("ordering=%-8s states=%2d hash=%s match=%s"
              % ("reversed" if rev else "qiskit", len(d), h[:16], h == onchain_hash))
    print("on-chain hash:", onchain_hash[:16])


# ---------------- run mode ----------------
def get_json(url):
    with urllib.request.urlopen(url, timeout=10) as r:
        return json.load(r)


def latest_round():
    seeds = get_json(API + "/gui/seeds").get("seeds", [])
    return (seeds[0]["round_id"], seeds[0]["seed"]) if seeds else (None, None)


def submit(round_id, dist, reverse):
    entries = [{"state": s, "probability": dist[s]} for s in sorted(dist)]
    payload = json.dumps({"entries": entries})
    args = [QC_BIN, "tx", "random_circuit", "submit-result", str(round_id),
            "--distribution", payload, "--from", QC_FROM, "--chain-id", QC_CHAIN,
            "--keyring-backend", QC_KEYRING, "--node", QC_RPC, "--home", QC_HOME,
            "--gas", "400000", "--fees", "0stake", "-y", "-o", "json"]
    out = subprocess.run(args, capture_output=True, text=True)
    print("submit round", round_id, "rc", out.returncode, (out.stdout or out.stderr)[:160], flush=True)


def run():
    # Qiskit is little-endian; daqq/itsubaki labels states in the opposite bit
    # order, so reverse by default to match the Go nodes' canonical hashes.
    reverse = os.environ.get("QC_REVERSE", "1") == "1"
    done = set()
    print("carol (Qiskit) listening; reverse=%s api=%s" % (reverse, API), flush=True)
    while True:
        try:
            rid, seed = latest_round()
            if rid is not None and rid not in done and seed:
                dist = compute_distribution(seed, reverse=reverse)
                submit(rid, dist, reverse)
                done.add(rid)
        except Exception as e:
            print("loop error:", e, flush=True)
        time.sleep(15)


if __name__ == "__main__":
    if len(sys.argv) >= 4 and sys.argv[1] == "verify":
        verify(sys.argv[2], sys.argv[3])
    elif len(sys.argv) >= 2 and sys.argv[1] == "run":
        run()
    else:
        print(__doc__)
        sys.exit(1)
