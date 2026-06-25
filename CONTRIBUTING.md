# Contributing

Thanks for your interest in cudackpt.

## Before you start

- Read [README.md](README.md) and [docs/OPERATIONS.md](docs/OPERATIONS.md)
- Run tests locally: `go test ./...` and `make test`
- GPU e2e requires Linux, NVIDIA CUDA 12.x, and CRIU: `sudo make e2e-fast`

## Development setup

```bash
git clone https://github.com/DDVHegde100/cudackpt.git
cd cudackpt
./scripts/install-hooks.sh   # optional local hooks
make
go test ./...
```

## Pull requests

1. Fork and branch from `main`
2. Keep changes focused — one logical change per PR
3. Add or update tests when behavior changes
4. Ensure `go test ./...` and CI pass
5. Update docs if CLI flags, env vars, or image layout change

## Commit style

Use short, imperative subjects:

```
fix(control): dial shim RPC via configured RunDir
docs: polish README and changelog for v0.1.0 public release
test: hermetic orchestrator integration with fake CRIU
```

## Reporting issues

Use [GitHub Issues](https://github.com/DDVHegde100/cudackpt/issues) and include:

- OS, driver, CUDA, Go, and CRIU versions (`cudackpt version`, `criu check`)
- Minimal repro steps
- Relevant logs (`restore.events.jsonl`, `CUDACKPT_DEBUG=1`)

## License

By contributing, you agree that your contributions remain subject to the repository license. See [LICENSE](LICENSE).
