# Releasing cmdr

## Local builds

Windows amd64:

```powershell
powershell -ExecutionPolicy Bypass -File ./scripts/build-windows.ps1 -Arch amd64
```

Linux amd64:

```bash
bash ./scripts/build-linux.sh amd64
```

Both local scripts use the version from `config/metadata.go` and inject it, Git commit, and build date with Go `-ldflags`.

## GitHub Actions

Pushes and pull requests to `main` run `.github/workflows/test.yml`. It builds Windows amd64 and Linux amd64, runs formatting, vet, and tests, and uploads test artifacts for pushes to `main`.

The release workflow is `.github/workflows/release.yml`.

## Publishing

Update `config/metadata.go` to the release version, commit it, and create a matching tag:

```bash
git add config/metadata.go
git commit -m "Release v2.1.5"
git tag v2.1.5
git push origin main
git push origin v2.1.5
```

Do not push a tag or publish a release without approval. The workflow only accepts `vMAJOR.MINOR.PATCH`, rejects `v.1.0.1`, runs tests, builds Windows amd64 and Linux amd64, creates archives, generates `checksums.txt`, and publishes the GitHub Release.

Release artifacts:

- `cmdr_2.1.5_windows_amd64.zip`
- `cmdr_2.1.5_linux_amd64.tar.gz`
- `checksums.txt`
