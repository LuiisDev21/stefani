# Starts (or restarts) STEFANI's web build on http://127.0.0.1:8080.
#
# The desktop binary does not open a port; this uses the `server` build tag and
# serves the frontend from disk, so `pnpm build` is visible on reload without
# restarting the process.
#
#   .\scripts\serve-web.ps1            # restart with the current dist
#   .\scripts\serve-web.ps1 -Rebuild   # also rebuild frontend + Go binary
#   .\scripts\serve-web.ps1 -Lan       # listen on the LAN (no authentication)
param(
    [switch]$Rebuild,
    [switch]$Lan
)

$ErrorActionPreference = "Stop"
$root = Resolve-Path (Join-Path $PSScriptRoot "..")

if ($Rebuild) {
    Push-Location (Join-Path $root "frontend")
    pnpm build
    Pop-Location
    Push-Location $root
    go build -tags server,production -o bin/stefani-web.exe .
    Pop-Location
}

Get-Process stefani-web -ErrorAction SilentlyContinue | Stop-Process -Force

$env:STEFANI_WEB_DIST = Join-Path $root "frontend/dist"
$env:STEFANI_WEB_HOST = if ($Lan) { "0.0.0.0" } else { "127.0.0.1" }

Start-Process -FilePath (Join-Path $root "bin/stefani-web.exe") `
    -WorkingDirectory $root `
    -WindowStyle Hidden `
    -RedirectStandardOutput "$env:TEMP\stefani-web.log" `
    -RedirectStandardError "$env:TEMP\stefani-web.err.log"

Start-Sleep -Seconds 2
$process = Get-Process stefani-web -ErrorAction SilentlyContinue | Select-Object -First 1
if (-not $process) {
    Write-Output "No arranco. Revisa $env:TEMP\stefani-web.err.log"
    exit 1
}
$hostLabel = if ($Lan) { "0.0.0.0 (red local)" } else { "127.0.0.1" }
Write-Output "STEFANI web: http://${hostLabel}:8080  ·  PID $($process.Id)"
