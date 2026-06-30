# Run daqq 24/7 on a macOS host (Intel)

An always-on Mac is a fine home for a single daqq node: plenty of RAM/CPU, no
cloud capacity limits, and — because the `/gui` visualizer is published through
a **Cloudflare Tunnel** (outbound only) — **the home IP stays private and no
router port-forwarding is needed**. Only the read-only `/gui` is reachable from
the internet; RPC (26657), gRPC (9090) and the non-GUI REST endpoints stay on
localhost.

macOS has no `systemd`, so the node runs under **launchd** instead: a
`LaunchDaemon` starts it at boot and restarts it on crash.

## Artifacts

| Path | What it is |
|------|------------|
| `deploy/macos/setup-mac.sh` | One-shot, idempotent Mac setup: brew deps → deploy-key clone → `make install` → `init-node.sh` → install + load the launchd daemon |
| `deploy/macos/com.daqq.quantumchaind.plist` | launchd `LaunchDaemon` (templated; `setup-mac.sh` fills `__USER__`/`__HOME__`) that runs `quantumchaind start` 24/7 |
| `deploy/scripts/init-node.sh` | Shared idempotent node init (BSD `sed`-compatible, so it runs on macOS) |
| `deploy/cloudflared/config.example.yml` | Tunnel config that publishes only `/gui` (change the `credentials-file` path to `/Users/<user>/...`) |

## Private repo access (read-only deploy key)

daqq is private, so the Mac authenticates with a read-only SSH deploy key.
One-time, on the Mac:

```sh
# 1. Generate a key pair (no passphrase)
ssh-keygen -t ed25519 -f ~/.ssh/daqq_deploy -N "" -C daqq-mac

# 2. Add the *public* key to GitHub:
#    repo -> Settings -> Deploy keys -> Add deploy key (leave "Allow write" OFF)
cat ~/.ssh/daqq_deploy.pub
```

## Setup

```sh
# Clone with the deploy key, then run the setup from the checkout
GIT_SSH_COMMAND="ssh -i ~/.ssh/daqq_deploy -o IdentitiesOnly=yes" \
  git clone git@github.com:massn/daqq.git ~/daqq
~/daqq/deploy/macos/setup-mac.sh
```

`setup-mac.sh` installs `go`, `jq`, `cloudflared` via Homebrew, builds the node,
initializes `~/.quantumchain` with localhost-only binds, then installs and loads
the launchd daemon (it `plutil -lint`s the plist first). Re-running it `git
pull`s and reloads. Override with `REPO_URL=` / `REPO_DIR=` / `DEPLOY_KEY=` /
`HOME_DIR=` env vars.

It needs **Homebrew** present first; if missing it prints the install command
and stops. Some steps use `sudo` (installing into `/Library/LaunchDaemons`).

### Keep the Mac awake

A sleeping Mac stops the node. Disable sleep (and auto-restart after a power
cut) so it truly runs 24/7:

```sh
sudo systemsetup -setcomputersleep Never
sudo systemsetup -setrestartpowerfailure on   # may require Full Disk Access
```

(Or System Settings → Energy / Battery → "Prevent automatic sleeping when the
display is off".)

## Operate

```sh
curl -s localhost:1317/gui/seeds | jq .                 # is it serving?
tail -f ~/Library/Logs/daqq/quantumchaind.*.log         # logs
sudo launchctl kickstart -k system/com.daqq.quantumchaind  # restart
sudo launchctl print system/com.daqq.quantumchaind | grep -E 'state|pid' # status
sudo launchctl bootout system /Library/LaunchDaemons/com.daqq.quantumchaind.plist # stop+unload
```

## Publish only /gui (Cloudflare Tunnel)

```sh
cloudflared tunnel login
cloudflared tunnel create daqq
cp ~/daqq/deploy/cloudflared/config.example.yml ~/.cloudflared/config.yml
# edit: set <TUNNEL_ID>, <HOSTNAME>, and credentials-file=/Users/<user>/.cloudflared/<TUNNEL_ID>.json
cloudflared tunnel route dns daqq <HOSTNAME>
sudo cloudflared --config ~/.cloudflared/config.yml service install
```

Then `https://<HOSTNAME>/gui/` shows the visualizer; visitors see Cloudflare's
IPs, never the Mac's.

## Optional: hot-swap via Cosmovisor

This setup runs `quantumchaind` directly, which is the simplest robust 24/7
configuration. To get governance-driven binary hot-swaps (Epic B) on the Mac
instead, run `deploy/scripts/setup-cosmovisor.sh`, then change the daemon's
`ProgramArguments` to `cosmovisor run start --home ~/.quantumchain` and add the
`DAEMON_*` env vars from `deploy/systemd/quantumchaind.service`. The hot-swap
flow itself is identical to the Linux runbook in `deploy/README.md`.
