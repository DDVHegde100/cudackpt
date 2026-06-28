# Configuration reference

cudackpt reads settings from `/etc/cudackpt.conf` and environment variables. The CLI, agent, and orchestrator all use `config.Load()`.

## Precedence

1. Built-in defaults
2. Config file (`/etc/cudackpt.conf` or `CUDACKPT_CONFIG`)
3. Environment variables (override file values)

The target CUDA process must also receive matching environment variables for shim-specific settings (for example `CUDACKPT_RUN_DIR`, `CUDACKPT_RPC_SECRET`).

## Config file format

`key=value` lines; `#` starts a comment.

Example (`deploy/cudackpt.conf.example`):

```ini
image_root=/var/lib/cudackpt
run_dir=/run/cudackpt
restore_timeout=60s
shim_poll=200ms
max_retries=3
retry_backoff=500ms
```

## Keys

| Key | Default | Purpose |
|-----|---------|---------|
| `image_root` | `/var/lib/cudackpt` | Checkpoint image storage root |
| `run_dir` | `/run/cudackpt` | Per-PID shim Unix socket directory |
| `restore_timeout` | `60s` | Max wait for shim after CRIU restore |
| `shim_poll` | `200ms` | Poll interval for `watch` and restore backoff base |
| `max_retries` | `3` | Checkpoint retry attempts |
| `retry_backoff` | `500ms` | Delay between checkpoint retries |
| `rpc_secret` | empty | Shared secret for authenticated shim RPC (env override: `CUDACKPT_RPC_SECRET`) |

## Environment variables

| Variable | Overrides |
|----------|-----------|
| `CUDACKPT_CONFIG` | Config file path |
| `CUDACKPT_IMAGE_ROOT` | `image_root` |
| `CUDACKPT_RUN_DIR` | `run_dir` (CLI **and** shim via `LD_PRELOAD`) |
| `CUDACKPT_RESTORE_TIMEOUT` | `restore_timeout` |
| `CUDACKPT_SHIM_POLL` | `shim_poll` |
| `CUDACKPT_MAX_RETRIES` | `max_retries` |
| `CUDACKPT_RETRY_BACKOFF` | `retry_backoff` |
| `CUDACKPT_RPC_SECRET` | `rpc_secret` (shim target process must also receive this env var) |

Unknown keys and invalid values emit warnings to stderr when the config file is loaded.

## Shim-only environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `CUDACKPT_SOCKET_GROUP` | `cudackpt` | Group owning shim IPC sockets (`0660`) |

## Agent-specific environment

These are read by `cudackpt agent` in addition to the shared config:

| Variable | Default | Purpose |
|----------|---------|---------|
| `CUDACKPT_METRICS_ADDR` | `127.0.0.1:9090` | Prometheus listen address |
| `CUDACKPT_AGENT_GC_INTERVAL` | disabled | Periodic image GC interval |
| `CUDACKPT_AGENT_GC_MAX_AGE` | `336h` (14d) | Delete images older than this |
| `CUDACKPT_PIN_FILE` | none | Paths exempt from GC |

The systemd unit `cudackpt-agent.service` sets core paths via `Environment=`; use `/etc/cudackpt.conf` for shared settings loaded by `config.Load()`.

## Debian package

Installing the `.deb` creates `/etc/cudackpt.conf` from the example file and ensures `/var/lib/cudackpt` and `/run/cudackpt` exist with `cudackpt` group ownership.
