param(
    [switch]$Mock,
    [switch]$Live
)

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$binary = Join-Path $projectRoot "bin\keyword-market-mcp.exe"

if (-not (Test-Path -LiteralPath $binary)) {
    throw "MCP binary not found at $binary. Run scripts\install-codex.ps1 first."
}

Set-Location -LiteralPath $projectRoot
if ($Mock -and $Live) {
    throw "Choose either -Mock or -Live, not both."
}
if ($Mock) {
    $env:GOOGLE_ADS_MOCK = "true"
}
elseif ($Live) {
    $env:GOOGLE_ADS_MOCK = "false"
}

& $binary --stdio
exit $LASTEXITCODE
