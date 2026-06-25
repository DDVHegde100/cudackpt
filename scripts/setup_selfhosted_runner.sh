#!/usr/bin/env bash
set -euo pipefail
if ! command -v gh >/dev/null 2>&1; then
  echo "install gh: https://cli.github.com"
  exit 1
fi
REPO="${1:-$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo DDVHegde100/cudackpt)}"
LABEL="${2:-gpu}"
DIR="${3:-$HOME/actions-runner}"
mkdir -p "$DIR"
cd "$DIR"
if [[ ! -f config.sh ]]; then
  curl -fsSL -o actions-runner.tar.gz \
    https://github.com/actions/runner/releases/download/v2.321.0/actions-runner-linux-x64-2.321.0.tar.gz
  tar xzf actions-runner.tar.gz
fi
TOKEN=$(gh api -X POST "repos/$REPO/actions/runners/registration-token" -q .token)
./config.sh --url "https://github.com/$REPO" --token "$TOKEN" --labels "$LABEL" --unattended
sudo ./svc.sh install
sudo ./svc.sh start
echo "runner registered for $REPO label=$LABEL"
