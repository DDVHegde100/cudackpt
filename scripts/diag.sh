#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
IMG="${1:-/tmp/cudackpt-e2e}"
echo "=== inspect $IMG ==="
"$ROOT/build/cudackpt" inspect "$IMG" 2>&1 || true
if [[ -f "$IMG/restore.log" ]]; then
  echo "=== restore.log ==="
  cat "$IMG/restore.log"
fi
if [[ -f "$IMG/restored.pid" ]]; then
  echo "=== restored.pid ==="
  cat "$IMG/restored.pid"
fi
if [[ -f "$IMG/snapshot.err" ]]; then
  echo "=== snapshot.err ==="
  cat "$IMG/snapshot.err"
fi
if [[ -f "$IMG/restore.err" ]]; then
  echo "=== restore.err ==="
  cat "$IMG/restore.err"
fi
echo "=== shims ==="
"$ROOT/build/cudackpt" ps 2>&1 || true
