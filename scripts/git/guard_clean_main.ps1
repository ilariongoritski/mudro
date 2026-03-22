Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$branch = (& git -C $repoRoot rev-parse --abbrev-ref HEAD 2>$null).Trim()

if (-not $branch -or $branch -eq 'HEAD') {
    exit 0
}

if ($branch -ne 'main' -and $branch -ne 'master') {
    exit 0
}

$status = & git -C $repoRoot status --porcelain --untracked-files=all
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

if (@($status).Count -gt 0) {
    Write-Host "Refusing to continue: worktree is dirty on protected branch '$branch'."
    Write-Host 'Commit, stash, or discard changes before retrying.'
    & git -C $repoRoot status --short
    exit 1
}

exit 0
