param(
    [Parameter(Mandatory=$true)][string]$DistDir
)

$ErrorActionPreference = "Stop"
$resolvedDist = (Resolve-Path -LiteralPath $DistDir).Path
$files = Get-ChildItem -LiteralPath $resolvedDist -File |
    Where-Object { $_.Name -like '*.zip' -or $_.Name -like '*.tar.gz' } |
    Sort-Object Name

if ($files.Count -eq 0) {
    throw "No release archives found in $resolvedDist."
}

$lines = foreach ($file in $files) {
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $file.FullName).Hash.ToLower()
    '{0}  {1}' -f $hash, $file.Name
}

$output = Join-Path $resolvedDist "checksums.txt"
[System.IO.File]::WriteAllLines($output, $lines, [System.Text.UTF8Encoding]::new($false))
Write-Output "Checksums: $output"
