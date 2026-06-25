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
echo "installed systemd units to $PREFIX/etc/systemd/system"
