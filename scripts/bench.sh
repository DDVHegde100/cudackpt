#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
make -q 2>/dev/null || make
echo "go tests"
go test ./...
echo "tracker test"
make test
echo "bench ok"
