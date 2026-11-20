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

Do not push a tag or publish a release without approval.

## Stable releases

Stable tags use exactly `vMAJOR.MINOR.PATCH` and are published as normal GitHub Releases (not prereleases):

```bash
# config/metadata.go: Version = "2.1.6"
git tag v2.1.6
git push origin v2.1.6
```

## Pre-releases

Test releases use the preferred `-rc.N` suffix. The workflow publishes them with `prerelease: true`:

```bash
# config/metadata.go: Version = "2.1.6-rc.1"
git tag v2.1.6-rc.1
git push origin v2.1.6-rc.1
```

After testing, publish a separate stable tag such as `v2.1.6`. Never move or overwrite existing tags or releases.

Release artifacts:

- `cmdr_2.1.5_windows_amd64.zip`
- `cmdr_2.1.5_linux_amd64.tar.gz`
- `checksums.txt`
