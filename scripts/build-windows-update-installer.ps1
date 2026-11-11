$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$stage = Join-Path $root "build\windows-amd64"
$config = Get-Content -LiteralPath (Join-Path $root "mygo.config.ts") -Raw
$versionMatch = [regex]::Match($config, 'version:\s*"([^"]+)"')
if (-not $versionMatch.Success) { throw "Cannot read the application version." }
$version = $versionMatch.Groups[1].Value
$installer = Join-Path $stage "FuckEnglish Setup $version amd64.exe"
$script = Join-Path $PSScriptRoot "windows-update-installer.nsi"

$makensis = (Get-Command makensis.exe -ErrorAction SilentlyContinue).Source
if (-not $makensis) {
  $cache = Join-Path $env:LOCALAPPDATA "mygo"
  $makensis = Get-ChildItem -Path $cache -Filter makensis.exe -Recurse -ErrorAction SilentlyContinue |
    Where-Object { $_.FullName -match "nsis-3\.13" } |
    Select-Object -First 1 -ExpandProperty FullName
}
if (-not $makensis) {
  throw "NSIS 3.13 makensis.exe was not found."
}

Remove-Item -LiteralPath $installer -Force -ErrorAction SilentlyContinue
& $makensis "-DAPP_STAGE=$stage" "-DAPP_VERSION=$version" "-DAPP_ID=io.github.reappealxy.fuckenglish.preview011" $script
if ($LASTEXITCODE -ne 0 -or -not (Test-Path -LiteralPath $installer)) {
  throw "NSIS failed to create the in-place update installer."
}
Write-Output "Created $installer"
