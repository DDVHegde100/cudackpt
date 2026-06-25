#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DAYS="${1:-14}"
DRY_RUN="${DRY_RUN:-0}"
ROOT_DIR="${CUDACKPT_IMAGE_ROOT:-/var/lib/cudackpt}"
PIN="${CUDACKPT_PIN_FILE:-}"

make -q 2>/dev/null || make go

args=(gc "--root" "$ROOT_DIR" "--older-than" "${DAYS}d")
if [[ -n "$PIN" ]]; then
  args+=(--pin "$PIN")
fi
if [[ "$DRY_RUN" == "1" ]]; then
  args+=(--dry-run)
fi

exec "$ROOT/build/cudackpt" "${args[@]}"
