#!/usr/bin/env bash
#
# sentry-allowlist.sh — manage which source IPs may reach the public sentry's
# p2p port. daqq participation is by application: the sentry port is NOT open to
# the whole internet. An applicant sends their public source IP; the operator
# adds it here, which inserts a per-IP ufw allow rule for the sentry p2p port.
# Everyone not on the list is dropped by ufw before CometBFT ever sees them.
#
# Run this on the sentry host (root/sudo). It only touches ufw rules scoped to
# the sentry port — it never opens the port to Anywhere.
#
# Usage:
#   sentry-allowlist.sh add    <IP> [note]   # grant an applicant access
#   sentry-allowlist.sh remove <IP>          # revoke access
#   sentry-allowlist.sh list                 # show current allowlist
#
# Env:
#   PORT   sentry p2p port (default 26636)
set -euo pipefail

PORT="${PORT:-26636}"
CMD="${1:-}"
IP="${2:-}"
NOTE="${3:-}"

die() { printf 'error: %s\n' "$*" >&2; exit 1; }

valid_ip() {
	# accept IPv4, IPv4/CIDR, or IPv6 (basic shape check; ufw does final validation)
	case "$1" in
		*[!0-9./]* ) case "$1" in *:* ) return 0 ;; * ) return 1 ;; esac ;;  # has non-v4 chars -> allow if it looks v6
		*.*.*.* ) return 0 ;;
		* ) return 1 ;;
	esac
}

command -v ufw >/dev/null 2>&1 || die "ufw not found (run on the sentry host)"

case "$CMD" in
	add)
		[ -n "$IP" ] || die "usage: sentry-allowlist.sh add <IP> [note]"
		valid_ip "$IP" || die "'$IP' does not look like an IP or CIDR"
		comment="daqq sentry peer${NOTE:+: $NOTE}"
		ufw allow from "$IP" to any port "$PORT" proto tcp comment "$comment"
		echo ">> allowed $IP -> :$PORT/tcp  ($comment)"
		;;
	remove|rm|del)
		[ -n "$IP" ] || die "usage: sentry-allowlist.sh remove <IP>"
		ufw delete allow from "$IP" to any port "$PORT" proto tcp
		echo ">> removed $IP"
		;;
	list|ls|"")
		echo ">> sentry allowlist for :$PORT/tcp"
		# show only the per-IP rules for this port (skip a broad Anywhere allow)
		ufw status | awk -v p="$PORT" '$1 ~ p"/tcp" && $0 !~ /Anywhere/ { print }' \
			|| echo "(none yet — the sentry is closed to all until an applicant is added)"
		;;
	*)
		die "unknown command '$CMD' (use: add | remove | list)"
		;;
esac
