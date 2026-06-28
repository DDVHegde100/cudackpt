# vectoradd checkpoint walkthrough

End-to-end checkpoint and restore of the bundled `vectoradd` test binary.

## Prerequisites

- Linux host with NVIDIA GPU, CUDA 12.x, CRIU, and cudackpt built (`make`)
- `sudo` for CRIU restore

## Run

```bash
./run.sh
```

The script:

1. Builds `vectoradd` if needed
2. Starts vectoradd under `LD_PRELOAD=libcudackpt.so`
3. Checkpoints the process
4. Kills and restores from the saved image
5. Verifies the restored process reaches `running` state

## Environment

| Variable | Purpose |
|----------|---------|
| `CUDACKPT_RUN_DIR` | Must match between app and CLI (default `/run/cudackpt`) |
| `CUDACKPT_IMAGE_ROOT` | Where checkpoint images are stored |
| `CUDACKPT_RPC_SECRET` | Optional shared secret for shim RPC |

See [docs/CONFIG.md](../../docs/CONFIG.md) for full configuration.
