# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-06-29

### Added

- **Config** — `rpc_secret` key in `/etc/cudackpt.conf` (overridden by `CUDACKPT_RPC_SECRET`)
- **CLI** — `cudackpt serve health` for socket-activated readiness checks
- **Agent** — deep health via `GET /health?deep=1` or `CUDACKPT_AGENT_DEEP_HEALTH=1`
- **Testing** — `DedupDevice` unit tests, restore failure metrics integration tests

### Fixed

- **Metrics** — `cudackpt_restore_failures_total` incremented on preflight, materialize, and CRIU errors
- **Metrics** — `cudackpt metrics` shuts down cleanly on SIGINT/SIGTERM
- **Shim** — unlink per-PID IPC socket on process exit
- **Packaging** — `.deb` declares `Depends: criu` and recommends NVIDIA driver packages

### Changed

- **CI** — nightly GPU workflow runs `go test -race`

## [0.2.0] - 2026-06-28

### Added

- **Shim** — `CUDACKPT_RUN_DIR`, socket mode `0660`, group ownership via `CUDACKPT_SOCKET_GROUP`, per-client IPC worker threads
- **Control plane** — `DialRunDir`, health `ProbeWith`, config file reference (`docs/CONFIG.md`), config parse warnings
- **CLI** — `watch --until-running`, shell completion, `cudackpt serve` for systemd socket activation
- **Agent** — GC error metrics/logging, `GET /health` readiness probe
- **Packaging** — deb postinst/prerm, example `/etc/cudackpt.conf`, systemd unit enablement
- **Testing** — ValidateImage, CheckpointWithRetry, WatchShim, pipeline, materialize, auth, and rollback hermetic tests
- **CI** — golangci-lint, Go race detector
- **Docs** — vectoradd checkpoint example, CONTRIBUTING/SECURITY/LICENSE

### Fixed

- Checkpoint failure metrics no longer inflate on transient retry errors
- Mock RPC server handles OpAuth, OpSnapshot, and status sequences for integration tests
- `cudackpt@.service` socket activation uses `cudackpt serve`

### Known limitations

- Single GPU only; no MIG, NCCL, or CUDA graphs
- GPU restore correctness is workload-dependent
- Full e2e validation requires a self-hosted GPU runner

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

[0.3.0]: https://github.com/DDVHegde100/cudackpt/releases/tag/v0.3.0
[0.2.0]: https://github.com/DDVHegde100/cudackpt/releases/tag/v0.2.0
[0.1.0]: https://github.com/DDVHegde100/cudackpt/releases/tag/v0.1.0
