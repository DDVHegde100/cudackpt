#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

IMAGE_ROOT="${CUDACKPT_IMAGE_ROOT:-/tmp/cudackpt-example}"
RUN_DIR="${CUDACKPT_RUN_DIR:-/run/cudackpt}"
IMAGE="${IMAGE_ROOT}/vectoradd-demo"
SHIM="${ROOT}/build/libcudackpt.so"
BIN="${ROOT}/build/vectoradd"
CLI="${ROOT}/build/cudackpt"

if [[ ! -f "$SHIM" || ! -f "$CLI" ]]; then
  make
fi

mkdir -p "$IMAGE_ROOT"
export LD_PRELOAD="$SHIM"
export CUDACKPT_RUN_DIR="$RUN_DIR"

echo "starting vectoradd..."
"$BIN" &
PID=$!
sleep 1

echo "checkpoint pid=$PID -> $IMAGE"
sudo -E "$CLI" checkpoint "$PID" "$IMAGE"
sudo kill "$PID" || true
sleep 1

echo "restore $IMAGE"
RESTORED=$(sudo -E "$CLI" restore "$IMAGE" | awk '{print $2}' | cut -d= -f2)
echo "restored pid=$RESTORED"
sudo -E "$CLI" watch "$RESTORED" --until-running --timeout 30s
echo "example ok"
