#!/usr/bin/env bash
#
# Install and initialize Cosmovisor for the daqq quantum-chain node so that a
# governance-approved software upgrade swaps in the new binary at its height
# with no manual downtime ("hot swap").
#
# Idempotent: safe to re-run. Requires Go (for `go install`) and an installed
# `quantumchaind` binary on PATH (build it first: `cd quantum-chain && make install`).
#
# Usage:
#   deploy/scripts/setup-cosmovisor.sh
#   DAEMON_HOME=/data/.quantumchain deploy/scripts/setup-cosmovisor.sh
#
set -euo pipefail

DAEMON_NAME="${DAEMON_NAME:-quantumchaind}"
DAEMON_HOME="${DAEMON_HOME:-$HOME/.quantumchain}"
# cosmovisor pulls an old bytedance/sonic that fails to link under Go 1.24
# (encoding/json.unquoteBytes was removed); build it with an older toolchain.
COSMOVISOR_VERSION="${COSMOVISOR_VERSION:-v1.7.1}"
COSMOVISOR_TOOLCHAIN="${COSMOVISOR_TOOLCHAIN:-go1.23.6}"

export DAEMON_NAME DAEMON_HOME

# 1. Resolve the current daemon binary.
DAEMON_BIN="$(command -v "$DAEMON_NAME" || true)"
if [ -z "$DAEMON_BIN" ]; then
	echo "error: '$DAEMON_NAME' not found on PATH." >&2
	echo "       build it first: (cd quantum-chain && make install)" >&2
	exit 1
fi

# 2. Install cosmovisor if missing.
if ! command -v cosmovisor >/dev/null 2>&1; then
	echo ">> installing cosmovisor (${COSMOVISOR_VERSION} via ${COSMOVISOR_TOOLCHAIN})..."
	GOTOOLCHAIN="$COSMOVISOR_TOOLCHAIN" go install "cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@${COSMOVISOR_VERSION}"
fi
cosmovisor version >/dev/null 2>&1 || true

# 3. Initialize the cosmovisor layout (idempotent). `cosmovisor init` places the
#    current binary at $DAEMON_HOME/cosmovisor/genesis/bin/$DAEMON_NAME.
GENESIS_BIN="$DAEMON_HOME/cosmovisor/genesis/bin/$DAEMON_NAME"
if [ -x "$GENESIS_BIN" ]; then
	echo ">> cosmovisor already initialized at $DAEMON_HOME/cosmovisor (skipping init)"
else
	echo ">> initializing cosmovisor layout under $DAEMON_HOME/cosmovisor"
	cosmovisor init "$DAEMON_BIN"
fi

# 4. Ensure the upgrades directory exists (where staged upgrade binaries live).
mkdir -p "$DAEMON_HOME/cosmovisor/upgrades"

# 5. Report the runtime environment cosmovisor needs (mirrored in the systemd unit).
cat <<EOF

Cosmovisor ready under: $DAEMON_HOME/cosmovisor

Runtime environment (also set in deploy/systemd/quantumchaind.service):

  export DAEMON_NAME=$DAEMON_NAME
  export DAEMON_HOME=$DAEMON_HOME
  export DAEMON_ALLOW_DOWNLOAD_BINARIES=false   # build & place upgrade binaries by hand
  export DAEMON_RESTART_AFTER_UPGRADE=true       # auto-restart the node after the swap

Start the node THROUGH cosmovisor (not quantumchaind directly):

  cosmovisor run start --home "$DAEMON_HOME"

To stage an upgrade named <name>, place its binary at:

  $DAEMON_HOME/cosmovisor/upgrades/<name>/bin/$DAEMON_NAME

When governance approves "software-upgrade <name>" at a height, cosmovisor swaps
to that binary at the height and (with RESTART_AFTER_UPGRADE) resumes the node.
EOF
