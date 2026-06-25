#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ $# -lt 2 ]]; then
  echo "usage: promote_image.sh <src-image> <dest-image> [--pin pinfile]" >&2
  exit 2
fi

SRC="$1"
DEST="$2"
shift 2
PIN="${CUDACKPT_PIN_FILE:-}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --pin)
      PIN="$2"
      shift 2
      ;;
    *)
      echo "unknown flag: $1" >&2
      exit 2
      ;;
  esac
done

make -q 2>/dev/null || make go

args=(promote "$SRC" "$DEST")
if [[ -n "$PIN" ]]; then
  args+=(--pin "$PIN")
fi

exec "$ROOT/build/cudackpt" "${args[@]}"
