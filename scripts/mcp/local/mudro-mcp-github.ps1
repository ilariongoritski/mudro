. "$PSScriptRoot\common.ps1"

Ensure-MudroCommand "npx.cmd"

$secretsDir = Get-MudroSecretsDir
$envFile = Join-Path $secretsDir "mudro-github-mcp.local.env"

New-MudroEnvFileIfMissing $envFile @(
    "# Fine-grained PAT for Mudro GitHub MCP"
    "# Recommended permissions for selected repos only:"
    "# Contents: Read-only"
    "# Pull requests: Read-only"
    "# Issues: Read-only"
    "# Actions: Read-only"
    "MUDRO_GITHUB_PAT="
)

if (-not $env:MUDRO_GITHUB_PAT) {
    Import-MudroEnvFile $envFile
}

if (-not $env:MUDRO_GITHUB_PAT) {
    throw "MUDRO_GITHUB_PAT is not set. Fill $envFile and then enable mudro_github in C:\\Users\\gorit\\.codex\\config.toml."
}

$env:GITHUB_PERSONAL_ACCESS_TOKEN = $env:MUDRO_GITHUB_PAT
& npx.cmd "-y" "@modelcontextprotocol/server-github@2025.4.8"
exit $LASTEXITCODE
