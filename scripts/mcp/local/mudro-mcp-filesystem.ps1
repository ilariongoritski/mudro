. "$PSScriptRoot\common.ps1"

$repoRoot = Get-MudroRepoRoot
Ensure-MudroCommand "npx.cmd"

& npx.cmd "-y" "@modelcontextprotocol/server-filesystem@2026.1.14" $repoRoot
exit $LASTEXITCODE
