# Updating cmdr

## Check for updates

```bash
cmdr update --check
```

This uses the public GitHub Releases API and makes no file changes.

## Install updates

```bash
cmdr update
```

The updater downloads the artifact for Windows amd64 or Linux amd64, downloads `checksums.txt`, verifies SHA-256, validates the archive, extracts only the expected executable, and performs a platform-specific install.

## Supported flags

```bash
cmdr update --force
cmdr update --version 2.1.5
cmdr update --prerelease
```

Stable releases are used by default and prereleases are never selected accidentally.

To test prereleases:

```bash
cmdr update --check --prerelease
cmdr update --prerelease
```

Prerelease mode recognizes SemVer prerelease tags and is intended for the preferred `-rc.N` test format. It can progress from a prerelease to a newer prerelease or to a stable release. Prereleases are never exposed by ordinary `cmdr update`. Downgrades and reinstalls still require `--force`.

## Security

GitHub SHA-256 is integrity verification, not an independent cryptographic signature of the publisher. Sigstore/Cosign support can be added later.

Archives reject absolute paths, traversal, symlinks, duplicate executables, and unexpected files.

## Windows

Windows does not overwrite a running executable directly. A hidden helper waits for the current process and performs the replacement with rollback.

If the installation directory is not writable, cmdr reports an error instead of requesting UAC or administrator access.

## Linux

Linux uses a staged executable beside the current executable and an atomic rename. Existing permissions and rollback are handled locally. cmdr does not invoke `sudo`.
