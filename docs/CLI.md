# cudackpt CLI Reference

The `cudackpt` binary is the control plane for checkpoint, restore, image management, and observability. Unless noted, commands talk to the shim over a Unix socket at `/run/cudackpt/<pid>.sock`.

Configuration is loaded from `/etc/cudackpt.conf` and environment variables (see [OPERATIONS.md](OPERATIONS.md)).

## Checkpoint and restore

### `cudackpt checkpoint <pid> [dir]`

Full checkpoint of a running process: freeze → GPU snapshot → optional image processing → CRIU dump → atomic finalize.

- **pid** — target process with `LD_PRELOAD=libcudackpt.so`
- **dir** — output image directory (default: `$CUDACKPT_IMAGE_ROOT/ckpt-<pid>`)

Environment toggles during snapshot: `CUDACKPT_COMPRESS`, `CUDACKPT_SPARSE`, `CUDACKPT_DEDUP`, `CUDACKPT_PARENT_IMAGE`.

### `cudackpt restore <image>`

Restore from a finalized image: preflight → materialize device → CRIU restore → GPU restore via shim → resume.

Prints `restore ok pid=<pid>` on success.

### `cudackpt rollback <image> [--stop <pid>]`

Validate and restore a prior known-good image. Optionally stop a running process first (`SIGTERM`, then `SIGKILL`).

### `cudackpt snapshot <pid> <dir>`

GPU snapshot only (freeze + write manifest/device sidecars). Does not invoke CRIU.

### `cudackpt gpu-restore <pid> <dir>`

GPU restore RPC only for an already-running process.

## Shim RPC

### `cudackpt freeze <pid>`

Quiesce CUDA work and enter frozen state.

### `cudackpt resume <pid>`

Resume application after checkpoint or restore.

### `cudackpt ping <pid>`

Verify shim socket connectivity.

### `cudackpt status <pid>`

Print shim state name and numeric code.

### `cudackpt stats <pid>`

Print tracker counters (allocs, bytes, streams, modules, unsupported code).

### `cudackpt watch <pid> [--until-running] [--timeout <duration>]`

Poll shim status until interrupted (Ctrl+C), or exit when `--until-running` sees `running`/`restored`, or fail on `--timeout`.

## Image inspection

### `cudackpt list [root]`

List image directories under root (default: `$CUDACKPT_IMAGE_ROOT`).

### `cudackpt inspect <image>`

Human-readable dump of meta, manifest summary, and error sidecars.

### `cudackpt validate <image>`

Verify manifest CRCs, required artifacts, and `COMPLETE` marker.

### `cudackpt report <image>`

Structured text report of manifest and metadata.

### `cudackpt diff <image-a> <image-b>`

Compare manifests and metadata; report drift.

## Retention and promotion

### `cudackpt gc [--root dir] [--older-than 14d] [--pin file] [--dry-run]`

Delete stale images: always removes `*.staging`, removes finalized images older than the retention window unless pinned.

### `cudackpt promote <src> <dest> [--pin file]`

Validate, copy an image to a stable path, and append to the pin file for GC exclusion.

Wrapper: `./scripts/promote_image.sh`.

## Observability

### `cudackpt health [-d]`

Host health probe. `-d` runs deep checks (driver, CRIU mem_track, capabilities).

Exit code 1 when degraded.

### `cudackpt bench <pid> [count]`

Measure RPC ping and status latency (default 100 iterations).

### `cudackpt metrics [--listen addr]`

Serve Prometheus metrics on `/metrics` (default `:9090`). Blocks until interrupted.

### `cudackpt agent [--listen addr]`

Long-running daemon: metrics HTTP server (`/metrics`), readiness probe (`GET /health`), periodic gauge refresh, optional scheduled GC via `CUDACKPT_AGENT_GC_INTERVAL`. GC failures increment `cudackpt_gc_errors_total` and emit JSON log events.

Systemd unit: `cudackpt-agent.service`.

### `cudackpt serve`

One-shot socket-activation helper for systemd. Reads a command from stdin (default `ps`) and writes shim listing to stdout. Used by `cudackpt@.service`.

### `cudackpt completion bash|zsh`

Emit shell completion script to stdout. Install:

```bash
# bash
source <(cudackpt completion bash)

# zsh
source <(cudackpt completion zsh)
```

## Process discovery

### `cudackpt ps [-v]`

List PIDs with active shim sockets. `-v` includes state names.

## Environment reference

| Variable | Purpose |
|----------|---------|
| `CUDACKPT_IMAGE_ROOT` | Checkpoint storage root |
| `CUDACKPT_RUN_DIR` | Shim socket directory (must match in target process and CLI) |
| `CUDACKPT_CONFIG` | Config file path |
| `CUDACKPT_LOG` | JSON log output file |
| `CUDACKPT_RPC_SECRET` | Shared secret for shim RPC auth |
| `CUDACKPT_METRICS_ADDR` | Agent/metrics listen address |
| `CUDACKPT_PIN_FILE` | GC/promote pin list |
| `CUDACKPT_AGENT_GC_INTERVAL` | Agent background GC interval |
| `LD_PRELOAD` | Path to `libcudackpt.so` (stored in meta at checkpoint) |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Command failed (validation, RPC, CRIU, I/O) |
| 2 | Usage error |

## See also

- [OPERATIONS.md](OPERATIONS.md) — retention, restore runbooks, systemd
- [README.md](../README.md) — architecture and quick start
