#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export GIT_AUTHOR_NAME="Dhruv Hegde"
export GIT_AUTHOR_EMAIL="ddvhegde100@gmail.com"
export GIT_COMMITTER_NAME="Dhruv Hegde"
export GIT_COMMITTER_EMAIL="ddvhegde100@gmail.com"
OLD="$(git rev-parse HEAD)"
git checkout --orphan reorg-main
git rm -rf --cached . >/dev/null 2>&1 || true
commit() {
  local msg="$1"
  shift
  git add -f "$@"
  if git diff --cached --quiet; then
    echo "skip empty: $msg"
    return
  fi
  TREE=$(git write-tree)
  if [ -z "${PARENT:-}" ]; then
    PARENT=$(git commit-tree "$TREE" -m "$msg")
  else
    PARENT=$(git commit-tree "$TREE" -p "$PARENT" -m "$msg")
  fi
  echo "$PARENT"
}
PARENT=""
PARENT=$(commit "chore: add gitignore and commit hooks" \
  .gitignore scripts/hooks scripts/install-hooks.sh)
PARENT=$(commit "docs: add README with architecture and installation" README.md)
PARENT=$(commit "build: add CMake CUDA build configuration" CMakeLists.txt)
PARENT=$(commit "build: add Makefile install and test targets" Makefile go.mod)
PARENT=$(commit "feat(shim): add sharded allocation tracker with sequence ordering" \
  shim/tracker.hpp shim/tracker.cpp tests/unit/tracker_test.cpp)
PARENT=$(commit "feat(shim): add CUDA driver interposition layer" \
  shim/interpose.c shim/cuda_check.h shim/app_hooks.c)
PARENT=$(commit "feat(shim): add Unix socket IPC and process state machine" \
  shim/ipc.c shim/ckpt_ops.h shim/log.h)
PARENT=$(commit "feat(shim): add parallel snapshot engine with CRC32C manifest" shim/snapshot.cu)
PARENT=$(commit "feat(shim): add parallel restore with VA remap fallback" shim/restore.cu)
PARENT=$(commit "feat(image): add versioned manifest format and chunk verification" \
  pkg/image/format.go pkg/image/format_test.go)
PARENT=$(commit "feat(image): add checkpoint metadata capture" \
  pkg/image/meta.go pkg/image/meta_test.go)
PARENT=$(commit "feat(storage): add tiered host image storage" \
  pkg/storage/tier.go pkg/storage/tier_test.go)
PARENT=$(commit "pkg: add typed errors and runtime configuration" \
  internal/ckpterr/err.go pkg/config/config.go pkg/config/config_test.go)
PARENT=$(commit "feat(rpc): add binary length-prefixed shim control protocol" \
  pkg/rpc/proto.go pkg/rpc/proto_test.go)
PARENT=$(commit "feat(criu): add process checkpoint wrapper with GPU mounts" \
  third_party/criu/wrapper.go)
PARENT=$(commit "feat(control): add checkpoint and restore orchestration" \
  pkg/control/orchestrator.go pkg/control/shimctl.go pkg/control/state.go)
PARENT=$(commit "feat(control): add inspect validate and ASCII image report" \
  pkg/control/inspect.go pkg/control/validate.go \
  pkg/report/report.go pkg/report/report_test.go)
PARENT=$(commit "feat(health): add host environment health probe" \
  pkg/health/health.go pkg/health/health_test.go)
PARENT=$(commit "feat(cli): add cudackpt command with report and health" cmd/cudackpt/main.go)
PARENT=$(commit "test: add vectoradd CUDA integration workload" tests/integration/vectoradd.cu)
PARENT=$(commit "ci: add Linux CUDA build and unit test workflow" .github/workflows/ci.yml)
PARENT=$(commit "ci: add self-hosted GPU end-to-end workflow" .github/workflows/e2e-selfhosted.yml)
PARENT=$(commit "feat(docker): add GPU container image for reproducible e2e" \
  Dockerfile .dockerignore)
PARENT=$(commit "feat(scripts): add environment validation and shim smoke test" \
  scripts/check_env.sh scripts/run_shim_smoke.sh)
PARENT=$(commit "feat(scripts): add checkpoint and restore helpers" \
  scripts/run_checkpoint.sh scripts/run_restore_only.sh)
PARENT=$(commit "feat(scripts): add e2e runners and failure diagnostics" \
  scripts/run_e2e.sh scripts/run_e2e_fast.sh scripts/diag.sh)
PARENT=$(commit "feat(scripts): add docker e2e and publish helpers" \
  scripts/run_docker_e2e.sh scripts/publish.sh)
PARENT=$(commit "feat(scripts): add self-hosted runner setup and full test pipeline" \
  scripts/setup_selfhosted_runner.sh scripts/run_all.sh)
PARENT=$(commit "feat(scripts): add benchmark and extended test runner" scripts/bench.sh)
git reset --hard "$PARENT"
git branch -M main
echo "reorganized to 30 commits on main (was $OLD)"
