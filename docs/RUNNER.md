# Self-hosted GPU runner

Full end-to-end validation (vectoradd checkpoint/restore) runs on a **self-hosted GitHub Actions runner** with an NVIDIA GPU. This is optional for development but recommended before production use.

## Requirements

- Linux x86_64 host with NVIDIA GPU
- CUDA 12.x toolkit and driver
- CRIU 3.x (`criu check` passes)
- Outbound network access to GitHub

## Register the runner

1. In GitHub: **Settings → Actions → Runners → New self-hosted runner**
2. Choose **Linux x64** and follow the install commands on the GPU host
3. Label the runner (e.g. `self-hosted`, `linux`, `gpu`)

## Verify on the host

```bash
git clone https://github.com/DDVHegde100/cudackpt.git
cd cudackpt
./scripts/check_env.sh
make
sudo make e2e-fast
```

## Workflows that use the runner

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `e2e-selfhosted.yml` | manual (`workflow_dispatch`) | GPU e2e on demand |
| `nightly-gpu.yml` | nightly cron | cuBLAS + pipeline matrix |
| `release.yml` → `gpu-gate` | tag `v*` | Gate releases (disabled until runner exists) |

## Enable release GPU gate

After the runner is registered and e2e passes locally, edit `.github/workflows/release.yml`:

```yaml
gpu-gate:
  if: true   # was: if: false
```

Push a new tag (e.g. `v0.2.1`) to run the gated release pipeline.
