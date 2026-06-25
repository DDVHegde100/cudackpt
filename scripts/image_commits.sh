#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export GIT_AUTHOR_NAME="Dhruv Hegde"
export GIT_AUTHOR_EMAIL="ddvhegde100@gmail.com"
export GIT_COMMITTER_NAME="Dhruv Hegde"
export GIT_COMMITTER_EMAIL="ddvhegde100@gmail.com"

PARENT="$(git rev-parse HEAD)"

commit() {
  local msg="$1"
  shift
  git add -f "$@"
  TREE="$(git write-tree)"
  PARENT="$(git commit-tree "$TREE" -p "$PARENT" -m "$msg")"
  echo "$PARENT"
}

PARENT="$(commit "feat(image): manifest v2 with compression flags" \
  pkg/image/flags.go pkg/image/format.go pkg/image/format_test.go \
  pkg/image/flags_test.go shim/snapshot.cu)"

PARENT="$(commit "feat(image): zstd chunk compression for device.bin" \
  pkg/image/compress.go pkg/image/compress_test.go go.mod go.sum)"

PARENT="$(commit "feat(storage): content-addressed chunk store with dedup" \
  pkg/storage/cas.go pkg/storage/cas_test.go pkg/image/dedup.go pkg/image/sparse_test.go)"

PARENT="$(commit "feat(storage): sparse device image support for zero pages" \
  pkg/image/sparse.go)"

PARENT="$(commit "feat(image): atomic checkpoint directory finalize" \
  pkg/image/finalize.go pkg/image/finalize_test.go pkg/image/pipeline.go)"

PARENT="$(commit "feat(image): incremental snapshot delta format" \
  pkg/image/delta.go)"

PARENT="$(commit "feat(control): structured JSON logging across orchestrator" \
  pkg/log/jsonlog.go pkg/log/jsonlog_test.go \
  pkg/control/orchestrator.go pkg/control/shimctl.go pkg/control/validate.go \
  scripts/image_commits.sh)"

git reset --hard "$PARENT"
git log --oneline -7
