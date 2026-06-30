#!/usr/bin/env bash
#
# Set up a daqq quantum-chain node to run 24/7 on a macOS host (tested target:
# an always-on Intel Mac). Installs deps via Homebrew, clones the private repo
# with a read-only SSH deploy key, builds the node, initializes it
# (localhost-only binds), and installs a launchd LaunchDaemon so it starts at
# boot and restarts on crash. Cloudflare Tunnel auth is interactive, so its
# final steps are printed rather than run.
#
# Only the embedded /gui is meant to be exposed (via the tunnel, outbound);
# RPC/gRPC and the non-GUI REST endpoints stay on localhost. The tunnel keeps
# the home IP private and needs no router port-forwarding.
#
# daqq is a PRIVATE repo: create a read-only deploy key first and add its
# public half to the repo's Deploy keys. See deploy/macos/README.md.
#
# Usage (on the Mac, after the deploy key is in place):
#   GIT_SSH_COMMAND="ssh -i ~/.ssh/daqq_deploy -o IdentitiesOnly=yes" \
#     git clone git@github.com:massn/daqq.git ~/daqq
#   ~/daqq/deploy/macos/setup-mac.sh
#
# Env overrides: REPO_URL=, REPO_DIR=, DEPLOY_KEY=, HOME_DIR=
#
set -euo pipefail

REPO_URL="${REPO_URL:-git@github.com:massn/daqq.git}"
REPO_DIR="${REPO_DIR:-$HOME/daqq}"
DEPLOY_KEY="${DEPLOY_KEY:-$HOME/.ssh/daqq_deploy}"
HOME_DIR="${HOME_DIR:-$HOME/.quantumchain}"
APP=quantumchaind
LABEL=com.daqq.quantumchaind
PLIST_DST="/Library/LaunchDaemons/${LABEL}.plist"
LOG_DIR="$HOME/Library/Logs/daqq"

say() { printf '\n>> %s\n' "$*"; }

# ---- 0. sanity --------------------------------------------------------------
[ "$(uname -s)" = "Darwin" ] || { echo "error: this script is for macOS (Darwin)." >&2; exit 1; }
case "$(uname -m)" in
	x86_64) ;; # Intel — the tested target
	arm64)  echo "note: Apple Silicon detected; Homebrew is /opt/homebrew. This script assumes Intel (/usr/local). Adjust PATHs if needed." >&2 ;;
esac

# ---- 1. Homebrew + dependencies --------------------------------------------
if ! command -v brew >/dev/null 2>&1; then
	echo "error: Homebrew not found. Install it first (https://brew.sh), then re-run:" >&2
	echo '  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"' >&2
	exit 1
fi
say "installing dependencies via Homebrew (go, jq, cloudflared)"
brew install go jq cloudflared

# ---- 2. SSH access to the (private) repo -----------------------------------
if printf '%s' "$REPO_URL" | grep -q '^git@'; then
	mkdir -p "$HOME/.ssh" && chmod 700 "$HOME/.ssh"
	if ! ssh-keygen -F github.com >/dev/null 2>&1; then
		say "trusting github.com host key"
		ssh-keyscan -t rsa,ecdsa,ed25519 github.com >>"$HOME/.ssh/known_hosts" 2>/dev/null || true
	fi
	SSH_CMD="ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new"
	if [ -f "$DEPLOY_KEY" ]; then
		SSH_CMD="$SSH_CMD -i $DEPLOY_KEY -o IdentitiesOnly=yes"
		export GIT_SSH_COMMAND="$SSH_CMD"
	fi
	# `ssh -T git@github.com` exits 1 even on success, so capture and grep.
	auth_out="$($SSH_CMD -T git@github.com 2>&1 || true)"
	if ! printf '%s' "$auth_out" | grep -qi 'successfully authenticated'; then
		echo "error: cannot authenticate to github.com over SSH." >&2
		echo "  daqq is private; add a read-only deploy key first:" >&2
		echo "    ssh-keygen -t ed25519 -f $DEPLOY_KEY -N '' -C daqq-mac" >&2
		echo "    # then add ${DEPLOY_KEY}.pub to GitHub repo -> Deploy keys (read-only)" >&2
		echo "  See deploy/macos/README.md." >&2
		exit 1
	fi
fi

# ---- 3. clone / update the repo --------------------------------------------
if [ -d "$REPO_DIR/.git" ]; then
	say "updating existing checkout at $REPO_DIR"
	git -C "$REPO_DIR" pull --ff-only
else
	say "cloning $REPO_URL -> $REPO_DIR"
	git clone "$REPO_URL" "$REPO_DIR"
fi

# ---- 4. build & install the node -------------------------------------------
say "building $APP"
export PATH="$PATH:$HOME/go/bin"
( cd "$REPO_DIR/quantum-chain" && make install )
"$APP" version

# ---- 5. initialize the node (idempotent, localhost-only binds) -------------
say "initializing node"
HOME_DIR="$HOME_DIR" "$REPO_DIR/deploy/scripts/init-node.sh"

# ---- 6. install the launchd LaunchDaemon -----------------------------------
say "installing launchd daemon $LABEL (24/7, restarts on crash)"
mkdir -p "$LOG_DIR"
TMP_PLIST="$(mktemp)"
sed -e "s#__USER__#$USER#g" -e "s#__HOME__#$HOME#g" \
	"$REPO_DIR/deploy/macos/${LABEL}.plist" >"$TMP_PLIST"
plutil -lint "$TMP_PLIST"
sudo cp "$TMP_PLIST" "$PLIST_DST"
sudo chown root:wheel "$PLIST_DST"
sudo chmod 644 "$PLIST_DST"
rm -f "$TMP_PLIST"
# Reload idempotently (ignore "not loaded" on first run), then start.
sudo launchctl bootout system "$PLIST_DST" 2>/dev/null || true
sudo launchctl bootstrap system "$PLIST_DST"
sudo launchctl enable "system/${LABEL}"
sleep 3
say "daemon status (pid shown if running):"
sudo launchctl print "system/${LABEL}" 2>/dev/null | grep -E 'state|pid =' | head -3 || true

cat <<EOF

================ NODE UP ================
Local GUI:   curl -s localhost:1317/gui/seeds | jq .
Logs:        tail -f "$LOG_DIR"/quantumchaind.*.log
Restart:     sudo launchctl kickstart -k system/${LABEL}

Publish ONLY /gui via Cloudflare Tunnel (interactive auth):
  cloudflared tunnel login
  cloudflared tunnel create daqq
  cp $REPO_DIR/deploy/cloudflared/config.example.yml ~/.cloudflared/config.yml
  # edit ~/.cloudflared/config.yml: set <TUNNEL_ID>, <HOSTNAME>, and change the
  # credentials-file path to /Users/$USER/.cloudflared/<TUNNEL_ID>.json
  cloudflared tunnel route dns daqq <HOSTNAME>
  sudo cloudflared --config ~/.cloudflared/config.yml service install
=========================================
EOF
