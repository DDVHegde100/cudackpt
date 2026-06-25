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

PARENT="$(commit "feat(shim): capture and restore CUDA module state" \
  shim/module_io.hpp shim/module_io.cpp \
  shim/tracker.hpp shim/tracker.cpp \
  shim/interpose.c)"

PARENT="$(commit "feat(shim): persist primary context and device flags in manifest" \
  shim/snapshot.cu shim/restore.cu)"

PARENT="$(commit "feat(shim): snapshot stream capture state beyond sync" \
  shim/module_io.cpp shim/snapshot.cu)"

PARENT="$(commit "feat(shim): add event and callback quiesce before freeze" \
  shim/quiesce.hpp shim/quiesce.cpp shim/ipc.c shim/interpose.c)"

PARENT="$(commit "feat(shim): reject unsupported APIs with structured error codes" \
  shim/errcode.h shim/errcode.c shim/interpose.c)"

PARENT="$(commit "feat(shim): add allocation coalescing for manifest compaction" \
  shim/coalesce.hpp shim/coalesce.cpp shim/snapshot.cu)"

PARENT="$(commit "perf(shim): pinned host staging buffers for snapshot and restore" \
  shim/pinned_pool.hpp shim/pinned_pool.cpp shim/snapshot.cu shim/restore.cu CMakeLists.txt)"

PARENT="$(commit "feat(shim): expose tracker stats over RPC" \
  shim/ipc.c pkg/rpc/proto.go pkg/control/shimctl.go cmd/cudackpt/main.go scripts/shim_commits.sh)"

git reset --hard "$PARENT"
git log --oneline -8
