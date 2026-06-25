# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-06-25

First public release.

### Added

- **Shim** — CUDA Driver API interposition via `LD_PRELOAD` with Unix-socket RPC control plane
- **Checkpoint / restore** — orchestrated GPU snapshot + CRIU dump/restore with versioned image format
- **Image pipeline** — compression, sparse pages, deduplication, and delta snapshots
- **CLI** — checkpoint, restore, rollback, inspect, validate, report, and granular RPC commands
- **Operations** — restore phase JSONL events, exponential backoff polling, pidfile handoff, image retention GC, promote-image helper
- **Observability** — Prometheus metrics endpoint via `cudackpt agent` and systemd units
- **Security** — optional RPC shared-secret auth (`CUDACKPT_RPC_SECRET`)
- **Packaging** — `.deb` build script, `VERSION` file, `cudackpt version` command
- **CI** — unit tests, Docker prod smoke, nightly GPU e2e matrix, tag-triggered release workflow
- **Docs** — README, [CLI reference](docs/CLI.md), [operations guide](docs/OPERATIONS.md)

### Known limitations

- Single GPU only; no MIG, NCCL, or CUDA graphs
- GPU restore correctness is workload-dependent
- Full e2e validation requires a self-hosted GPU runner

[0.1.0]: https://github.com/DDVHegde100/cudackpt/releases/tag/v0.1.0
