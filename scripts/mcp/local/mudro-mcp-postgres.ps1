. "$PSScriptRoot\common.ps1"

Ensure-MudroCommand "npx.cmd"

$secretsDir = Get-MudroSecretsDir
$envFile = Join-Path $secretsDir "mudro-postgres-mcp.local.env"

New-MudroEnvFileIfMissing $envFile @(
    "# Local Mudro MCP Postgres DSN"
    "MUDRO_MCP_LOCAL_DSN=postgres://postgres:postgres@localhost:5433/gallery?sslmode=disable"
)

if (-not $env:MUDRO_MCP_LOCAL_DSN) {
    Import-MudroEnvFile $envFile
}

if (-not $env:MUDRO_MCP_LOCAL_DSN) {
    throw "MUDRO_MCP_LOCAL_DSN is not set. Update $envFile and restart the MCP client."
}

& npx.cmd "-y" "@modelcontextprotocol/server-postgres@0.6.2" $env:MUDRO_MCP_LOCAL_DSN
exit $LASTEXITCODE
