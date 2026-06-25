#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
make -q 2>/dev/null || make
export LD_PRELOAD="$ROOT/build/libcudackpt.so"
"$ROOT/build/vectoradd" &
PID=$!
sleep 1
"$ROOT/build/cudackpt" ping "$PID"
"$ROOT/build/cudackpt" status "$PID"
kill "$PID" 2>/dev/null || true
echo "shim smoke ok pid=$PID"
