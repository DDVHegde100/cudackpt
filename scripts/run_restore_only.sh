#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
IMG="${1:-${CKPT_IMAGE:-/tmp/cudackpt-ckpt}}"
OUT="/tmp/vectoradd.out"
"$ROOT/scripts/check_env.sh"
make -q 2>/dev/null || make
test -d "$IMG/criu"
RESTORED=$(sudo -E "$ROOT/build/cudackpt" restore "$IMG" | awk '{print $NF}')
for _ in $(seq 1 240); do
  if [[ -f "$OUT" ]]; then
    break
  fi
  sleep 0.25
done
if [[ ! -f "$OUT" ]]; then
  "$ROOT/scripts/diag.sh" "$IMG" || true
  echo "timeout restored_pid=$RESTORED"
  exit 1
fi
echo "restore ok pid=$RESTORED sum=$(tr -d ' \n' < "$OUT")"
