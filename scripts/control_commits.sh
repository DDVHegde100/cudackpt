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

PARENT="$(commit "feat(control): checkpoint job queue with timeout and retry policy" \
  pkg/control/queue.go pkg/config/config.go)"

PARENT="$(commit "feat(control): restore preflight checks before CRIU" \
  pkg/control/preflight.go pkg/control/orchestrator.go)"

PARENT="$(commit "feat(control): compare two images and report drift" \
  pkg/control/drift.go pkg/control/drift_test.go)"

PARENT="$(commit "feat(cli): watch command for shim state transitions" \
  pkg/control/watch.go cmd/cudackpt/main.go)"

PARENT="$(commit "feat(cli): bench subcommand with throughput and latency tables" \
  pkg/bench/bench.go pkg/bench/bench_test.go cmd/cudackpt/main.go)"

PARENT="$(commit "feat(health): deep probe for driver version CRIU features and caps" \
  pkg/health/deep.go pkg/health/deep_test.go pkg/health/caps_linux.go pkg/health/caps_stub.go cmd/cudackpt/main.go)"

PARENT="$(commit "feat(config): file-based config with env override precedence" \
  pkg/config/config.go pkg/config/config_test.go pkg/config/load_test.go cmd/cudackpt/main.go)"

PARENT="$(commit "test: rpc integration tests with mock shim socket" \
  pkg/rpc/proto.go pkg/rpc/integration_test.go scripts/control_commits.sh)"

git reset --hard "$PARENT"
git log --oneline -8
