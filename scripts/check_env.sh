#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
ok=1
if [[ "$(uname -s)" != "Linux" ]]; then
  echo "need Linux"
  ok=0
fi
if ! command -v nvcc >/dev/null 2>&1; then
  echo "need nvcc"
  ok=0
fi
if ! command -v criu >/dev/null 2>&1; then
  echo "need criu"
  ok=0
fi
if ! command -v cmake >/dev/null 2>&1; then
  echo "need cmake"
  ok=0
fi
if ! command -v go >/dev/null 2>&1; then
  echo "need go"
  ok=0
fi
if [[ ! -e /dev/nvidia0 ]] && [[ ! -e /dev/nvidiactl ]]; then
  echo "need nvidia device nodes"
  ok=0
fi
if ! criu check >/dev/null 2>&1; then
  echo "criu check failed (try sudo criu check)"
  ok=0
fi
if [[ $ok -eq 1 ]]; then
  echo "env ok"
  exit 0
fi
exit 1
