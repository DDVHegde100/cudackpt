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

PARENT="$(commit "test: orchestrator table tests for restore discovery" \
  pkg/control/restore.go pkg/control/orchestrator.go pkg/control/orchestrator_test.go)"

PARENT="$(commit "test: add cublas workload integration test" \
  tests/integration/cublas_gemm.cu CMakeLists.txt scripts/run_cublas_smoke.sh Makefile)"

PARENT="$(commit "test: add failure injection tests for snapshot and restore errors" \
  pkg/control/failure_test.go pkg/image/failure_test.go)"

PARENT="$(commit "ci: add nightly self-hosted GPU matrix workflow" \
  .github/workflows/nightly-gpu.yml)"

PARENT="$(commit "feat(scripts): systemd unit and socket activation for control plane" \
  deploy/cudackpt-run.service deploy/cudackpt.socket deploy/cudackpt@.service \
  scripts/install-systemd.sh Makefile)"

PARENT="$(commit "feat(docker): multi-stage production image with non-root runtime" \
  Dockerfile.prod)"

PARENT="$(commit "docs: operations guide for checkpoint retention restore and rollback" \
  docs/OPERATIONS.md README.md scripts/ops_commits.sh)"

git reset --hard "$PARENT"
git log --oneline -7
