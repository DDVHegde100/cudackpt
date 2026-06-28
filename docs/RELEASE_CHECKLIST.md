# Release checklist

Use this before tagging a new version.

## Pre-release

- [ ] `VERSION` matches the intended tag (without `v` prefix)
- [ ] `CHANGELOG.md` has a section for the release
- [ ] `.github/release-notes/vX.Y.Z.md` exists for the tag
- [ ] `go test -race ./...` and `make lint` pass locally
- [ ] `make deb` succeeds (or CI deb job on `main`)
- [ ] `make docker-prod-smoke` succeeds when Docker is available
- [ ] `SECURITY.md` lists the release in supported versions
- [ ] Restore and GC metrics documented in `docs/OPERATIONS.md`

## Tag and publish

```bash
git tag -a vX.Y.Z -m "cudackpt X.Y.Z"
git push origin main
git push origin vX.Y.Z
```

The release workflow builds the `.deb`, runs verify (tests + lint), then publishes to GitHub Releases.

## Post-release

- [ ] Confirm CI on `main` is green
- [ ] Confirm release workflow completed and `.deb` is attached
- [ ] Enable `gpu-gate` in `release.yml` once a self-hosted GPU runner is registered (see [RUNNER.md](RUNNER.md))

## Hotfix

For patch releases, branch from the tag if needed, cherry-pick fixes, bump `VERSION`, update changelog, and tag `vX.Y.Z+1`.
