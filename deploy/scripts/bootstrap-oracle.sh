#!/usr/bin/env bash
#
# One-shot, idempotent bootstrap for a daqq node on an Oracle Cloud Always Free
# VM (Ubuntu 22.04, ARM64 / Ampere A1). Installs deps + Go, builds the node,
# initializes it (localhost-only binds), puts it under Cosmovisor, and starts it
# as a systemd service. Cloudflare Tunnel needs interactive auth, so its final
# steps are printed rather than run.
#
# Safe to re-run: every step checks before acting.
#
# daqq is a PRIVATE repo, so the VM authenticates to GitHub with a read-only
# deploy key (default: ~/.ssh/daqq_deploy). Create that key and add its public
# half to the repo's Deploy keys (read-only) BEFORE running this. There is no
# public `curl | bash` one-liner — clone the checkout over SSH first, then run
# this script from it. See deploy/README.md ("Private repo access").
#
# Usage (on the VM, after the deploy key is in place):
#   GIT_SSH_COMMAND="ssh -i ~/.ssh/daqq_deploy -o IdentitiesOnly=yes" \
#     git clone git@github.com:massn/daqq.git ~/daqq
#   ~/daqq/deploy/scripts/bootstrap-oracle.sh
#
# Env overrides: REPO_URL=, REPO_DIR=, DEPLOY_KEY=, GO_VER=
#
set -euo pipefail

REPO_URL="${REPO_URL:-git@github.com:massn/daqq.git}"
REPO_DIR="${REPO_DIR:-$HOME/daqq}"
DEPLOY_KEY="${DEPLOY_KEY:-$HOME/.ssh/daqq_deploy}"
GO_VER="${GO_VER:-1.24.1}"
APP=quantumchaind

# Detect CPU architecture so prebuilt downloads (Go, cloudflared) match the host.
# Works on both x86 VPS (e.g. ConoHa) and ARM (e.g. Oracle Ampere).
case "$(uname -m)" in
	x86_64 | amd64) ARCH=amd64 ;;
	aarch64 | arm64) ARCH=arm64 ;;
	*)
		echo "unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
esac

say() { printf '\n>> %s\n' "$*"; }

# ---- 1. OS dependencies ---------------------------------------------------
if command -v apt-get >/dev/null 2>&1; then
	say "installing apt dependencies"
	sudo apt-get update -y
	sudo apt-get install -y git make jq build-essential curl
else
	echo "warning: apt-get not found; ensure git/make/jq/build-essential/curl are installed" >&2
fi

# ---- 2. Go toolchain ------------------------------------------------------
NEED_GO=1
if command -v go >/dev/null 2>&1 && go version | grep -q "go${GO_VER}"; then NEED_GO=0; fi
if [ "$NEED_GO" = 1 ]; then
	say "installing Go ${GO_VER} (${ARCH})"
	curl -fsSL "https://go.dev/dl/go${GO_VER}.linux-${ARCH}.tar.gz" -o /tmp/go.tgz
	sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
	rm -f /tmp/go.tgz
fi
export PATH="$PATH:/usr/local/go/bin:$HOME/go/bin"
grep -q '/usr/local/go/bin' "$HOME/.bashrc" 2>/dev/null \
	|| echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >>"$HOME/.bashrc"
go version

# ---- 3. clone / update the repo (private repo, SSH deploy key) ------------
# For an SSH remote, make sure git uses the deploy key and trusts github.com,
# then verify auth up front so failures are obvious (not a vague clone error).
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
		echo "    ssh-keygen -t ed25519 -f $DEPLOY_KEY -N '' -C daqq-oracle-vm" >&2
		echo "    # then add ${DEPLOY_KEY}.pub to GitHub repo -> Deploy keys (read-only)" >&2
		echo "  See deploy/README.md ('Private repo access')." >&2
		exit 1
	fi
fi

if [ -d "$REPO_DIR/.git" ]; then
	say "updating existing checkout at $REPO_DIR"
	git -C "$REPO_DIR" pull --ff-only
else
	say "cloning $REPO_URL -> $REPO_DIR"
	git clone "$REPO_URL" "$REPO_DIR"
fi

# ---- 4. build & install the node -----------------------------------------
say "building $APP"
( cd "$REPO_DIR/quantum-chain" && make install )
"$APP" version

# ---- 5. initialize the node (idempotent, localhost-only binds) -----------
say "initializing node"
"$REPO_DIR/deploy/scripts/init-node.sh"

# ---- 6. put the node under Cosmovisor ------------------------------------
say "setting up cosmovisor"
"$REPO_DIR/deploy/scripts/setup-cosmovisor.sh"

# ---- 7. install + start the systemd service ------------------------------
say "installing systemd unit (templated for user $USER)"
UNIT_SRC="$REPO_DIR/deploy/systemd/quantumchaind.service"
UNIT_DST=/etc/systemd/system/quantumchaind.service
sudo sed -e "s#^User=.*#User=$USER#" \
	-e "s#/home/ubuntu#$HOME#g" \
	-e "s#^Environment=PATH=.*#Environment=PATH=/usr/local/go/bin:$HOME/go/bin:/usr/local/bin:/usr/bin:/bin#" \
	"$UNIT_SRC" | sudo tee "$UNIT_DST" >/dev/null
sudo systemctl daemon-reload
sudo systemctl enable --now quantumchaind
sleep 3
sudo systemctl --no-pager --full status quantumchaind | head -12 || true

# ---- 8. cloudflared (binary now; tunnel auth is interactive) --------------
if ! command -v cloudflared >/dev/null 2>&1; then
	say "installing cloudflared (${ARCH})"
	curl -fsSL "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${ARCH}" -o /tmp/cloudflared
	sudo install -m 0755 /tmp/cloudflared /usr/local/bin/cloudflared && rm -f /tmp/cloudflared
fi
cloudflared --version || true

cat <<EOF

================ NODE UP ================
Local GUI:   curl -s localhost:1317/gui/seeds | jq .
Logs:        journalctl -u quantumchaind -f

Finish publishing ONLY /gui via Cloudflare Tunnel (interactive):
  cloudflared tunnel login
  cloudflared tunnel create daqq
  cp $REPO_DIR/deploy/cloudflared/config.example.yml ~/.cloudflared/config.yml
  # edit <TUNNEL_ID> and <HOSTNAME> in that file
  cloudflared tunnel route dns daqq <HOSTNAME>
  sudo cloudflared --config ~/.cloudflared/config.yml service install
=========================================
EOF
