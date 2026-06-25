#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
"$ROOT/scripts/check_env.sh"
make -q 2>/dev/null || make
"$ROOT/scripts/run_shim_smoke.sh"
sudo -E "$ROOT/scripts/run_e2e_fast.sh"
echo "all ok"
