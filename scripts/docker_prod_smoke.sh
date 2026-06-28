#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TAG="${1:-cudackpt:prod-smoke}"
docker build -f Dockerfile.prod -t "$TAG" .

docker run --rm --entrypoint /bin/sh "$TAG" -c '
  set -e
  test -x /usr/bin/cudackpt
  test -f /usr/lib/libcudackpt.so
  id cudackpt >/dev/null
  /usr/bin/cudackpt ps
  echo ps | /usr/bin/cudackpt serve
  command -v criu >/dev/null
  criu --version >/dev/null
'

docker run --rm --entrypoint /usr/bin/cudackpt "$TAG" health >/dev/null || true

echo "docker prod smoke ok tag=$TAG"
