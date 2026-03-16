. "$PSScriptRoot\common.ps1"

$repoRoot = Get-MudroRepoRoot
$image = "mudro-mcp-git:2026.1.14"
$dockerfile = Join-Path $repoRoot "scripts\mcp\docker\git-server\Dockerfile"

Ensure-MudroCommand "docker.exe"

$null = & docker.exe "image" "inspect" $image 2>$null
if ($LASTEXITCODE -ne 0) {
    & docker.exe "build" "--tag" $image "--file" $dockerfile $repoRoot
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

$mount = "type=bind,src=$repoRoot,dst=/workspace,readonly"
& docker.exe "run" "--rm" "-i" "--network" "none" "--mount" $mount $image "--repository" "/workspace"
exit $LASTEXITCODE
