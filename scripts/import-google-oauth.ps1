param(
    [Parameter(Mandatory = $true)]
    [string]$ClientSecretFile,

    [string]$ApplicationDefaultCredentialsFile = (Join-Path $env:APPDATA "gcloud\application_default_credentials.json"),

    [string]$EnvFile
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($EnvFile)) {
    $EnvFile = Join-Path (Split-Path -Parent $PSScriptRoot) ".env"
}

$clientConfig = Get-Content -Raw -LiteralPath $ClientSecretFile | ConvertFrom-Json
if ($null -eq $clientConfig.installed) {
    throw "Expected an OAuth Desktop app credentials file with an 'installed' section."
}

$adc = Get-Content -Raw -LiteralPath $ApplicationDefaultCredentialsFile | ConvertFrom-Json
if ([string]::IsNullOrWhiteSpace($adc.refresh_token)) {
    throw "The application-default credentials file does not contain a refresh token."
}

$updates = [ordered]@{
    GOOGLE_ADS_CLIENT_ID     = [string]$clientConfig.installed.client_id
    GOOGLE_ADS_CLIENT_SECRET = [string]$clientConfig.installed.client_secret
    GOOGLE_ADS_REFRESH_TOKEN = [string]$adc.refresh_token
}

$lines = [Collections.Generic.List[string]]::new()
$lines.AddRange([string[]](Get-Content -LiteralPath $EnvFile))
foreach ($entry in $updates.GetEnumerator()) {
    $found = $false
    for ($i = 0; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -match ('^\s*' + [regex]::Escape($entry.Key) + '\s*=')) {
            $lines[$i] = $entry.Key + "=" + $entry.Value
            $found = $true
            break
        }
    }
    if (-not $found) {
        $lines.Add($entry.Key + "=" + $entry.Value)
    }
}

[IO.File]::WriteAllLines($EnvFile, $lines, [Text.UTF8Encoding]::new($false))
Write-Host "OAuth client credentials and refresh token imported into .env."
