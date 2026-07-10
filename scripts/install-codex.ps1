param(
    [switch]$Live
)

$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
$binaryDir = Join-Path $projectRoot "bin"
$binary = Join-Path $binaryDir "keyword-market-mcp.exe"
$launcher = Join-Path $PSScriptRoot "start-mcp.ps1"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required but was not found on PATH."
}
if (-not (Get-Command codex -ErrorAction SilentlyContinue)) {
    throw "Codex CLI is required but was not found on PATH."
}

New-Item -ItemType Directory -Force -Path $binaryDir | Out-Null
Push-Location $projectRoot
try {
    & go build -o $binary .\cmd\keyword-market-mcp
    if ($LASTEXITCODE -ne 0) {
        throw "Go build failed with exit code $LASTEXITCODE."
    }
}
finally {
    Pop-Location
}

$previousErrorActionPreference = $ErrorActionPreference
$ErrorActionPreference = "SilentlyContinue"
& codex mcp get keyword-market *> $null
$serverExists = $LASTEXITCODE -eq 0
$ErrorActionPreference = $previousErrorActionPreference
if ($serverExists) {
    & codex mcp remove keyword-market
    if ($LASTEXITCODE -ne 0) {
        throw "Could not remove the existing keyword-market MCP configuration."
    }
}

$command = @(
    "mcp", "add", "keyword-market", "--",
    "powershell.exe", "-NoLogo", "-NoProfile", "-NonInteractive",
    "-ExecutionPolicy", "Bypass", "-File", $launcher
)
if ($Live) {
    $command += "-Live"
}
else {
    $command += "-Mock"
}

& codex @command
if ($LASTEXITCODE -ne 0) {
    throw "Could not register keyword-market with Codex."
}

# This server performs research reads and only persists results to its local
# SQLite database. Pre-approve just this server so non-interactive Codex calls
# do not cancel while waiting for a confirmation prompt.
$codexHome = if ($env:CODEX_HOME) { $env:CODEX_HOME } else { Join-Path $HOME ".codex" }
$configPath = Join-Path $codexHome "config.toml"
$configText = [IO.File]::ReadAllText($configPath)
$tablePattern = '(?ms)(^\[mcp_servers\.keyword-market\]\r?\n(?:(?!^\[).*(?:\r?\n|$))*)'
$tableMatch = [regex]::Match($configText, $tablePattern)
if (-not $tableMatch.Success) {
    throw "keyword-market was added but its config table was not found in $configPath."
}
if ($tableMatch.Value -notmatch '(?m)^default_tools_approval_mode\s*=') {
    $updatedTable = $tableMatch.Value.TrimEnd("`r", "`n") + "`r`ndefault_tools_approval_mode = `"approve`"`r`n"
    $configText = $configText.Remove($tableMatch.Index, $tableMatch.Length).Insert($tableMatch.Index, $updatedTable)
    [IO.File]::WriteAllText($configPath, $configText, [Text.UTF8Encoding]::new($false))
}

$mode = if ($Live) { "live Google Ads" } else { "deterministic mock" }
Write-Host "keyword-market MCP installed in $mode mode."
Write-Host "Restart ChatGPT Desktop/Codex, then use /mcp or ask it to validate a product idea."
