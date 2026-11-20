param(
    [switch]$Linux
)

$ErrorActionPreference = "Stop"
& (Join-Path $PSScriptRoot "build-windows.ps1") -Arch all
if ($LASTEXITCODE -ne 0) { throw "Windows build failed." }

if ($Linux) {
    if (-not (Get-Command bash -ErrorAction SilentlyContinue)) {
        throw "bash is required to run the Linux build script."
    }
    & bash (Join-Path $PSScriptRoot "build-linux.sh") all
    if ($LASTEXITCODE -ne 0) { throw "Linux build failed." }
}
