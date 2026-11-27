param(
    [Parameter(Mandatory=$true)][string]$DistDir,
    [Parameter(Mandatory=$true)][string]$Version,
    [Parameter(Mandatory=$true)][string]$Root
)

$ErrorActionPreference = "Stop"
$staging = Join-Path $DistDir ".package"
$windowsSource = Join-Path $Root "bin/windows/amd64/cmdr.exe"

if (Test-Path $staging) {
    Remove-Item -Recurse -Force $staging
}

if (-not (Test-Path $windowsSource)) {
    throw "Windows release binary not found: $windowsSource"
}

New-Item -ItemType Directory -Force -Path (Join-Path $staging "windows") | Out-Null
Copy-Item -LiteralPath $windowsSource -Destination (Join-Path $staging "windows/cmdr.exe")
Copy-Item -LiteralPath (Join-Path $Root "README.md") -Destination (Join-Path $staging "windows/README.md")
Copy-Item -LiteralPath (Join-Path $Root "LICENSE") -Destination (Join-Path $staging "windows/LICENSE")

$archive = Join-Path $DistDir "cmdr_${Version}_windows_amd64.zip"
if (Test-Path $archive) {
    Remove-Item -Force $archive
}
$windowsDir = Join-Path $staging "windows"
Compress-Archive -Force -Path (Join-Path $windowsDir "cmdr.exe"), (Join-Path $windowsDir "README.md"), (Join-Path $windowsDir "LICENSE") -DestinationPath $archive
Remove-Item -Recurse -Force $staging
Write-Output "Package created: $archive"
