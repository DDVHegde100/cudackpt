#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$ROOT/.git/hooks"
install -m 755 "$ROOT/scripts/hooks/prepare-commit-msg" "$ROOT/.git/hooks/prepare-commit-msg"
install -m 755 "$ROOT/scripts/hooks/commit-msg" "$ROOT/.git/hooks/commit-msg"
echo "installed prepare-commit-msg and commit-msg hooks"
