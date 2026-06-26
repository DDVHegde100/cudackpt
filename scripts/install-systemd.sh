#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PREFIX="${DESTDIR:-/}"
install -d "$PREFIX/etc/systemd/system"
install -m 644 "$ROOT/deploy/cudackpt-run.service" "$PREFIX/etc/systemd/system/"
install -m 644 "$ROOT/deploy/cudackpt.socket" "$PREFIX/etc/systemd/system/"
install -m 644 "$ROOT/deploy/cudackpt@.service" "$PREFIX/etc/systemd/system/"
install -m 644 "$ROOT/deploy/cudackpt-agent.service" "$PREFIX/etc/systemd/system/"
getent group cudackpt >/dev/null || groupadd -r cudackpt 2>/dev/null || true
if ! getent passwd cudackpt >/dev/null; then
  useradd -r -g cudackpt -d /var/lib/cudackpt -s /usr/sbin/nologin cudackpt 2>/dev/null || true
fi
echo "installed systemd units to $PREFIX/etc/systemd/system"
