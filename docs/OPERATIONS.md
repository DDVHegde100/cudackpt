# Operations Guide

This document covers checkpoint retention, restore procedures, and rollback for cudackpt deployments.

## Checkpoint retention

Images are stored under `CUDACKPT_IMAGE_ROOT` (default `/var/lib/cudackpt`) as directories named `ckpt-<pid>` or custom paths passed to `cudackpt checkpoint`.

Each finalized image contains a `COMPLETE` marker. Do not use directories missing this file.

Recommended retention policy:

1. Keep the last known-good image per workload indefinitely.
2. Keep hourly snapshots for 24 hours, daily for 14 days.
3. Delete staging directories (`*.staging`) immediately; they are incomplete.

Example cleanup of images older than 14 days:

```bash
find /var/lib/cudackpt -maxdepth 1 -type d -name 'ckpt-*' -mtime +14 -exec rm -rf {} +
```

Before deletion, validate critical images:

```bash
cudackpt validate /var/lib/cudackpt/ckpt-12345
cudackpt report /var/lib/cudackpt/ckpt-12345
```

Compressed and delta images require materialization before GPU restore. The orchestrator handles this automatically during `cudackpt restore`.

## Restore procedure

1. Confirm host health:

```bash
cudackpt health -d
```

2. Validate the target image:

```bash
cudackpt validate /var/lib/cudackpt/ckpt-<pid>
cudackpt diff /var/lib/cudackpt/ckpt-<pid> /var/lib/cudackpt/ckpt-<other>
```

3. Restore:

```bash
sudo cudackpt restore /var/lib/cudackpt/ckpt-<pid>
```

The restore pipeline:

1. Preflight checks (`COMPLETE`, manifest, meta, criu, device artifacts)
2. Materialize `device.bin` from compression or delta parent
3. CRIU process restore with environment from `meta.bin`
4. GPU restore via shim RPC on discovered PIDs
5. Resume signal to application

Monitor restore:

```bash
cudackpt watch <new-pid> --until-running --timeout 60s
cudackpt stats <new-pid>
```

Inspect failures:

```bash
cat /var/lib/cudackpt/ckpt-<pid>/restore.log
cat /var/lib/cudackpt/ckpt-<pid>/restore.err
./scripts/diag.sh /var/lib/cudackpt/ckpt-<pid>
```

## Rollback

Rollback means restoring a prior known-good image after a failed upgrade or bad checkpoint.

1. Stop the current process if it is still running.
2. Select the previous validated image:

```bash
cudackpt list /var/lib/cudackpt
cudackpt validate /var/lib/cudackpt/ckpt-<old-pid>
```

3. Restore the old image:

```bash
sudo cudackpt restore /var/lib/cudackpt/ckpt-<old-pid>
```

4. Verify application output and GPU state:

```bash
cudackpt stats <pid>
```

If GPU restore fails but CRIU restore succeeds, use granular recovery:

```bash
cudackpt gpu-restore <pid> /var/lib/cudackpt/ckpt-<old-pid>
cudackpt resume <pid>
```

## Systemd deployment

Install runtime directory and socket units:

```bash
sudo ./scripts/install-systemd.sh
sudo systemctl daemon-reload
sudo systemctl enable --now cudackpt-run.service
sudo systemctl enable cudackpt.socket
```

Production container (non-root runtime):

```bash
docker build -f Dockerfile.prod -t cudackpt:prod .
```

## Environment reference

| Variable | Purpose |
|----------|---------|
| `CUDACKPT_IMAGE_ROOT` | Checkpoint storage root |
| `CUDACKPT_RUN_DIR` | Shim socket directory (must match in target process and CLI) |
| `CUDACKPT_CONFIG` | Config file path |
| `CUDACKPT_COMPRESS` | Enable zstd compression post-snapshot |
| `CUDACKPT_SPARSE` | Enable sparse zero-page encoding |
| `CUDACKPT_DEDUP` | Enable CAS deduplication |
| `CUDACKPT_PARENT_IMAGE` | Parent path for delta snapshots |
| `CUDACKPT_LOG` | JSON log output file |

Shim IPC sockets are created mode `0660` under `run_dir`. The `cudackpt` system group (see `cudackpt-run.service`) can dial sockets without root when the target process runs as a member of that group.

## Failure triage

| Symptom | Check |
|---------|-------|
| Checkpoint aborts early | `snapshot.err`, `cudackpt stats <pid>` unsupported code |
| Restore hangs | `restore.log`, shim sockets in `/run/cudackpt` |
| CRC validation fails | `cudackpt validate`, re-checkpoint |
| CRIU errors | `criu check`, `CAP_SYS_ADMIN`, `health -d` |
