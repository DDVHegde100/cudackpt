#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
IMG="${CKPT_IMAGE:-/tmp/cudackpt-e2e-cublas}"
READY="/tmp/cublas.ready"
OUT="/tmp/cublas.out"
EXPECT="4194304.000000"
diag() {
  echo "cublas e2e failed"
  "$ROOT/scripts/diag.sh" "$IMG" || true
}
trap diag ERR
"$ROOT/scripts/check_env.sh"
make
mkdir -p /run/cudackpt
rm -f "$READY" "$OUT"
rm -rf "$IMG"
export LD_PRELOAD="$ROOT/build/libcudackpt.so"
"$ROOT/build/cublas_gemm" &
PID=$!
for _ in $(seq 1 120); do
  if [[ -f "$READY" ]]; then
    break
  fi
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "cublas_gemm exited early"
    exit 1
  fi
  sleep 0.25
done
if [[ ! -f "$READY" ]]; then
  kill "$PID" 2>/dev/null || true
  echo "timeout waiting for cublas ready"
  exit 1
fi
sudo -E "$ROOT/build/cudackpt" checkpoint "$PID" "$IMG"
"$ROOT/build/cudackpt" validate "$IMG"
kill -9 "$PID" 2>/dev/null || true
sleep 1
RESTORED=$(sudo -E "$ROOT/build/cudackpt" restore "$IMG" | awk '{print $NF}')
for _ in $(seq 1 240); do
  if [[ -f "$OUT" ]]; then
    break
  fi
  sleep 0.25
done
if [[ ! -f "$OUT" ]]; then
  echo "timeout waiting for cublas output restored_pid=$RESTORED"
  exit 1
fi
GOT="$(tr -d ' \n' < "$OUT")"
if [[ "$GOT" != "$EXPECT" ]]; then
  echo "bad cublas result got=$GOT want=$EXPECT restored_pid=$RESTORED"
  exit 1
fi
trap - ERR
echo "cublas e2e ok sum=$GOT restored_pid=$RESTORED"
