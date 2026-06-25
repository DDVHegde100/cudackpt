#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
make -q 2>/dev/null || make
READY=/tmp/cublas.ready
OUT=/tmp/cublas.out
rm -f "$READY" "$OUT"
export LD_PRELOAD="$ROOT/build/libcudackpt.so"
"$ROOT/build/cublas_gemm" &
pid=$!
for i in $(seq 1 60); do
  [[ -f "$READY" ]] && break
  sleep 0.2
done
[[ -f "$READY" ]] || { kill "$pid" 2>/dev/null; exit 1; }
echo "cublas workload ready pid=$pid"
