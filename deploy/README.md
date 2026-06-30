# daqq deployment & protocol upgrades

Reproducible artifacts for running a daqq node 24/7 and for rolling out a new
**protocol version** (the shared data structures — see the root README's
[Protocol](../README.md#protocol) section) with no downtime via Cosmovisor.

> Private, environment-specific notes (real hostnames, the Cloudflare Tunnel
> walkthrough) live in `deploy/oracle-cloudflare.md`, which is git-ignored on
> purpose. Everything in this README and the directories below **is** tracked.

## Artifacts

| Path | What it is |
|------|------------|
| `deploy/scripts/bootstrap-oracle.sh` | **One-shot, idempotent** VM setup: deps → Go → build → init → cosmovisor → systemd → cloudflared |
| `deploy/scripts/init-node.sh` | Initialize a node home idempotently with localhost-only binds |
| `deploy/scripts/setup-cosmovisor.sh` | Install + initialize Cosmovisor so upgrades hot-swap the binary |
| `deploy/scripts/healthcheck.sh` | Liveness probe + auto-restart if the chain stalls (timer/cron) |
| `deploy/scripts/upgrade-drill.sh` | Local end-to-end rehearsal of a governance upgrade + Cosmovisor swap |
| `deploy/systemd/quantumchaind.service` | systemd unit that runs the node **through Cosmovisor** |
| `deploy/cloudflared/` | Cloudflare Tunnel config templates (publish only `/gui`) |

## Private repo access (read-only deploy key)

daqq is a **private** repository, so a fresh VM can't `git clone` it (and there
is no public `curl | bash` one-liner — `raw.githubusercontent.com` is private
too). The VM authenticates with a **read-only SSH deploy key**. One-time setup:

```sh
# 1. On the VM, generate a key pair (no passphrase)
ssh-keygen -t ed25519 -f ~/.ssh/daqq_deploy -N "" -C daqq-oracle-vm

# 2. Add the *public* key to GitHub:
#    repo -> Settings -> Deploy keys -> Add deploy key (leave "Allow write" OFF)
cat ~/.ssh/daqq_deploy.pub
```

A read-only deploy key is scoped to this one repo and can't push — minimal blast
radius if the VM is compromised. (Alternatives, not used here: a fine-grained
read-only PAT over HTTPS, or `rsync` from a Mac checkout.)

## Hosting (Oracle Cloud Always Free)

**On a fresh Ubuntu 22.04 ARM64 VM, after the deploy key is in place:**

```sh
# Clone over SSH with the deploy key, then run the bootstrap from the checkout
GIT_SSH_COMMAND="ssh -i ~/.ssh/daqq_deploy -o IdentitiesOnly=yes" \
  git clone git@github.com:massn/daqq.git ~/daqq
~/daqq/deploy/scripts/bootstrap-oracle.sh
```

`bootstrap-oracle.sh` defaults `REPO_URL` to `git@github.com:massn/daqq.git`,
auto-detects `~/.ssh/daqq_deploy`, trusts `github.com`'s host key, and verifies
SSH auth before cloning (clear error if the deploy key is missing). Re-running it
`git pull`s with the same key. Override with `REPO_URL=` / `REPO_DIR=` /
`DEPLOY_KEY=` env vars.

It then runs, idempotently: apt deps → Go → clone/build → `init-node.sh`
(localhost-only binds) → `setup-cosmovisor.sh` → install & start the systemd
service → install cloudflared. It then prints the interactive Cloudflare Tunnel
steps (auth can't be scripted). Optionally wire `healthcheck.sh` to a 5-minute
timer (snippet at the bottom of that script).

The individual steps, if you prefer to run them by hand, are: `make install` →
`init-node.sh` → `setup-cosmovisor.sh` → install `quantumchaind.service` →
publish only `/gui` via the `cloudflared/` template. The private
`oracle-cloudflare.md` has the long-form Cloudflare walkthrough.

Design rule that must not be broken: **only the read-only `/gui` visualizer is
public.** RPC (26657), gRPC (9090) and the non-GUI REST endpoints stay private.

## Protocol upgrade runbook (cut a new version, hot-swap with no downtime)

In daqq, "upgrading the protocol" means changing the shared on-chain data
structures (e.g. proto `…v1` → `…v2`, or adding a new problem module) and
migrating state. Cosmos governance schedules it; Cosmovisor swaps the binary at
the agreed height so the network never stops.

**Always rehearse locally first:** `deploy/scripts/upgrade-drill.sh` runs the
whole flow against a throwaway node and asserts the chain survives the swap.

To cut and ship version `<name>` (e.g. `v2`):

1. **Change the protocol.** Add the new proto version / module, bump the owning
   module's consensus version, and write its migration.
2. **Declare the upgrade.** Add `app/upgrades/<name>/constants.go` with
   `UpgradeName = "<name>"` and a `StoreUpgrades` listing any added/renamed/
   deleted module stores (see `app/upgrades/v1_1` for the no-op template).
3. **Wire it.** In `app/upgrades.go`:
   - register the handler in `setupUpgradeHandlers()` (run migrations), and
   - add a `case <name>.UpgradeName` in `setupStoreLoaders()` if stores changed.
4. **Build & stage the new binary on every node** (downloads are disabled):

   ```sh
   cd quantum-chain && make install
   mkdir -p "$DAEMON_HOME/cosmovisor/upgrades/<name>/bin"
   cp "$(command -v quantumchaind)" "$DAEMON_HOME/cosmovisor/upgrades/<name>/bin/quantumchaind"
   ```

5. **Propose & vote** the on-chain upgrade for a future height `H`:

   ```sh
   GOV=$(quantumchaind q auth module-account gov -o json | jq -r '.account.value.address')
   # proposal.json: a MsgSoftwareUpgrade with authority=$GOV and plan {name:"<name>", height:"H"}
   quantumchaind tx gov submit-proposal proposal.json --from <key> -y ...
   quantumchaind tx gov vote <id> yes --from <key> -y ...
   ```

6. **At height `H`** the running binary halts (`UPGRADE "<name>" NEEDED`),
   Cosmovisor swaps to `upgrades/<name>/bin` and restarts
   (`DAEMON_RESTART_AFTER_UPGRADE=true`); migrations run and the chain resumes.
   Confirm height keeps advancing and `journalctl -u quantumchaind` shows the swap.

The local drill exercises exactly this path: genesis binary (no handler) →
`UPGRADE NEEDED` → swap to `upgrades/<name>/bin` → chain continues.
