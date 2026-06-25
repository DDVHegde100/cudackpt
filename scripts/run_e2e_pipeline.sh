#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

run_cycle() {
  local img="$1"
  local expect="$2"
  local label="$3"
  shift 3
  local -a extra_env=("$@")
  local ready="/tmp/vectoradd.ready"
  local out="/tmp/vectoradd.out"
  rm -f "$ready" "$out"
  rm -rf "$img"
  env "${extra_env[@]}" LD_PRELOAD="$ROOT/build/libcudackpt.so" "$ROOT/build/vectoradd" &
  local pid=$!
  for _ in $(seq 1 120); do
    [[ -f "$ready" ]] && break
    kill -0 "$pid" 2>/dev/null || { echo "$label: vectoradd exited early"; return 1; }
    sleep 0.25
  done
  [[ -f "$ready" ]] || { kill "$pid" 2>/dev/null || true; echo "$label: ready timeout"; return 1; }
  sudo -E env "${extra_env[@]}" "$ROOT/build/cudackpt" checkpoint "$pid" "$img"
  "$ROOT/build/cudackpt" validate "$img"
  kill -9 "$pid" 2>/dev/null || true
  sleep 1
  local restored
  restored=$(sudo -E "$ROOT/build/cudackpt" restore "$img" | awk '{print $NF}')
  for _ in $(seq 1 240); do
    [[ -f "$out" ]] && break
    sleep 0.25
  done
  [[ -f "$out" ]] || { echo "$label: output timeout pid=$restored"; return 1; }
  local got
  got="$(tr -d ' \n' < "$out")"
  [[ "$got" == "$expect" ]] || { echo "$label: bad sum got=$got want=$expect"; return 1; }
  echo "$label ok sum=$got restored_pid=$restored"
}

diag() {
  echo "pipeline e2e failed"
}
trap diag ERR

"$ROOT/scripts/check_env.sh"
make
mkdir -p /run/cudackpt

EXPECT="3145728.000000"
PARENT="${CKPT_PARENT:-/tmp/cudackpt-pipeline-parent}"
COMPRESS_IMG="${CKPT_COMPRESS:-/tmp/cudackpt-pipeline-compress}"
DELTA_IMG="${CKPT_DELTA:-/tmp/cudackpt-pipeline-delta}"

run_cycle "$PARENT" "$EXPECT" "baseline"
run_cycle "$COMPRESS_IMG" "$EXPECT" "compress" CUDACKPT_COMPRESS=1 CUDACKPT_SPARSE=1
[[ -f "$COMPRESS_IMG/device.zst" ]] || { echo "compress: missing device.zst"; exit 1; }

rm -f /tmp/vectoradd.ready /tmp/vectoradd.out
env VECTORADD_B_SCALE=3 LD_PRELOAD="$ROOT/build/libcudackpt.so" "$ROOT/build/vectoradd" &
PID=$!
for _ in $(seq 1 120); do
  [[ -f /tmp/vectoradd.ready ]] && break
  sleep 0.25
done
[[ -f /tmp/vectoradd.ready ]] || exit 1
sudo -E env CUDACKPT_COMPRESS=1 CUDACKPT_PARENT_IMAGE="$PARENT" VECTORADD_B_SCALE=3 \
  "$ROOT/build/cudackpt" checkpoint "$PID" "$DELTA_IMG"
"$ROOT/build/cudackpt" validate "$DELTA_IMG"
[[ -f "$DELTA_IMG/delta.bin" ]] || { echo "delta: missing delta.bin"; exit 1; }
kill -9 "$PID" 2>/dev/null || true
sleep 1
RESTORED=$(sudo -E "$ROOT/build/cudackpt" restore "$DELTA_IMG" | awk '{print $NF}')
for _ in $(seq 1 240); do
  [[ -f /tmp/vectoradd.out ]] && break
  sleep 0.25
done
GOT="$(tr -d ' \n' < /tmp/vectoradd.out)"
DELTA_EXPECT="4194304.000000"
[[ "$GOT" == "$DELTA_EXPECT" ]] || { echo "delta: bad sum got=$GOT want=$DELTA_EXPECT"; exit 1; }
echo "delta ok sum=$GOT restored_pid=$RESTORED"

trap - ERR
echo "pipeline e2e ok"
