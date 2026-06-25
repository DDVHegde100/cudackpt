#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
NAME="${1:-cudackpt}"
if ! gh auth status >/dev/null 2>&1; then
  echo "run: gh auth login -h github.com"
  exit 1
fi
if git remote get-url origin >/dev/null 2>&1; then
  git push -u origin main
  exit 0
fi
gh repo create "$NAME" --public --source=. --remote=origin --push
