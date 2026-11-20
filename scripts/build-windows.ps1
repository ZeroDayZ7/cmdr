param(
    [ValidateSet("amd64")]
    [string]$Arch = "amd64"
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $root

function Test-CommandExists([string]$Name) {
    return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Get-ProjectVersion {
    $metadataPath = Join-Path $root "config/metadata.go"
    $metadata = Get-Content -Raw -Path $metadataPath
    if ($metadata -notmatch 'Version\s*=\s*"([0-9]+(?:\.[0-9]+){1,3})"') {
        throw "Could not read a valid Version from config/metadata.go"
    }
    return $Matches[1]
}

function Convert-VersionParts([string]$Version) {
    $parts = $Version.Split('.') | ForEach-Object { [int]$_ }
    while ($parts.Count -lt 4) { $parts += 0 }
    if ($parts.Count -ne 4) { throw "Unsupported version format: $Version" }
    return $parts
}

function Write-VersionConfig([string]$Version) {
    $parts = Convert-VersionParts $Version
    $configPath = Join-Path $root "build/windows/versioninfo.json"
    $config = @{
        FixedFileInfo = @{
            FileVersion = @{ Major = $parts[0]; Minor = $parts[1]; Patch = $parts[2]; Build = $parts[3] }
            ProductVersion = @{ Major = $parts[0]; Minor = $parts[1]; Patch = $parts[2]; Build = $parts[3] }
            FileFlagsMask = "3f"
            FileFlags = "00"
            FileOS = "040004"
            FileType = "01"
            FileSubType = "00"
        }
        StringFileInfo = @{
            Comments = ""
            CompanyName = ""
            FileDescription = "cmdr command-line developer tools"
            FileVersion = $Version
            InternalName = "cmdr"
            LegalCopyright = ""
            LegalTrademarks = ""
            OriginalFilename = "cmdr.exe"
            PrivateBuild = ""
            ProductName = "cmdr"
            ProductVersion = $Version
            SpecialBuild = ""
        }
        VarFileInfo = @{
            Translation = @{ LangID = "0409"; CharsetID = "04B0" }
        }
    }
    $json = $config | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($configPath, $json + [Environment]::NewLine, [System.Text.UTF8Encoding]::new($false))
}

function Write-Manifest([string]$Version) {
    $manifestPath = Join-Path $root "build/windows/manifest.xml"
    $manifest = [System.IO.File]::ReadAllText($manifestPath, [System.Text.UTF8Encoding]::new($false))
    $manifestVersion = ($Version.Split('.') | ForEach-Object { [int]$_ })
    while ($manifestVersion.Count -lt 4) { $manifestVersion += 0 }
    $assemblyVersion = (($manifestVersion[0..3]) -join '.')
    $manifest = [regex]::Replace($manifest, '(<assemblyIdentity[^>]*\bversion=")[^"]*(")', '${1}' + $assemblyVersion + '${2}')
    [System.IO.File]::WriteAllText($manifestPath, $manifest, [System.Text.UTF8Encoding]::new($false))
}

if (-not (Test-CommandExists "go")) { throw "Go is not available in PATH." }

if (-not (Test-Path "build/windows/icon.ico")) { throw "Missing build/windows/icon.ico" }
if (-not (Test-Path "build/windows/manifest.xml")) { throw "Missing build/windows/manifest.xml" }

$goversioninfo = Get-Command goversioninfo -ErrorAction SilentlyContinue
if ($null -eq $goversioninfo) {
    Write-Host "goversioninfo is missing; installing pinned v1.7.0..."
    & go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.7.0
    if ($LASTEXITCODE -ne 0) { throw "Failed to install goversioninfo." }
    $goversioninfo = Get-Command goversioninfo -ErrorAction SilentlyContinue
}
if ($null -eq $goversioninfo) {
    $candidate = Join-Path ([Environment]::GetEnvironmentVariable("GOPATH", "Process") -replace "^$", (go env GOPATH)) "bin/goversioninfo.exe"
    if (-not (Test-Path $candidate)) { throw "goversioninfo was installed but is not available in PATH: $candidate" }
    $goversioninfo = Get-Item $candidate
}

if (-not (Test-CommandExists "gcc")) {
    throw "CGO is required by github.com/chai2010/webp; GCC is not available in PATH."
}

$version = Get-ProjectVersion
$parts = Convert-VersionParts $version
Write-VersionConfig $version
Write-Manifest $version

$architectures = @("amd64")

foreach ($targetArch in $architectures) {
    $compiler = Get-Command gcc -ErrorAction SilentlyContinue

    Get-ChildItem -Path $root -Filter "resource_windows_*.syso" -File -ErrorAction SilentlyContinue |
        Remove-Item -Force

    $syso = Join-Path $root "resource_windows_$targetArch.syso"
    $extraArgs = @("-64")

    Write-Host "Generating Windows resource for $targetArch..."
    & $goversioninfo.Source @extraArgs "-icon" (Join-Path $root "build/windows/icon.ico") "-manifest" (Join-Path $root "build/windows/manifest.xml") "-o" $syso "-propagate-ver-strings" (Join-Path $root "build/windows/versioninfo.json")
    if ($LASTEXITCODE -ne 0) { throw "goversioninfo failed for $targetArch." }
    if (-not (Test-Path $syso)) { throw "Expected resource was not created: $syso" }

    $outDir = Join-Path $root "bin/windows/$targetArch"
    New-Item -ItemType Directory -Force -Path $outDir | Out-Null
    $output = Join-Path $outDir "cmdr.exe"

    $commit = (& git rev-parse HEAD 2>$null)
    if (-not $commit) { $commit = "unknown" }
    $buildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-dd'T'HH:mm:ss'Z'")
    $ldflags = "-s -w -X github.com/zerodayz7/cmdr/config.Version=$version -X github.com/zerodayz7/cmdr/config.GitCommit=$commit -X github.com/zerodayz7/cmdr/config.BuildDate=$buildDate"

    Write-Host "Building cmdr $version for windows/$targetArch..."
    & go build -trimpath -ldflags $ldflags -o $output .
    if ($LASTEXITCODE -ne 0) { throw "Build failed for windows/$targetArch." }
    if (-not (Test-Path $output)) { throw "Build did not create $output" }

    $length = (Get-Item $output).Length
    if ($length -le 0) { throw "Built executable is empty: $output" }
    Write-Host "SUCCESS: $output ($length bytes)"

    Remove-Item -Force $syso -ErrorAction SilentlyContinue
}

$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

Remove-Variable env:GOOS -ErrorAction SilentlyContinue
Remove-Variable env:GOARCH -ErrorAction SilentlyContinue
Remove-Variable env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Variable env:CC -ErrorAction SilentlyContinue
Get-ChildItem -Path $root -Filter "resource_windows_*.syso" -ErrorAction SilentlyContinue | Remove-Item -Force
