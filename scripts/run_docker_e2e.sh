#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
IMG="${CKPT_IMAGE:-/tmp/cudackpt-smoke}"
TAG="${CUDACKPT_IMAGE:-cudackpt:dev}"
docker build -t "$TAG" "$ROOT"
docker run --rm --gpus all --privileged \
  -v /run/cudackpt:/run/cudackpt \
  -e CKPT_IMAGE="$IMG" \
  "$TAG" ./scripts/run_e2e.sh
