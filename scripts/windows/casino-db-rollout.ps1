param(
    [switch]$PrintOnly,
    [switch]$SkipGit,
    [switch]$SkipTests,
    [switch]$SkipSmoke,
    [switch]$FullSmoke
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$runner = Join-Path $repoRoot 'scripts\windows\run-mudro-no-profile.ps1'

if (-not (Test-Path $runner)) {
    throw "Runner not found: $runner"
}

$steps = @(
    @{ Name = 'Git status'; Enabled = (-not $SkipGit); Command = 'git status --short' },
    @{ Name = 'Main DB casino migrations'; Enabled = $true; Command = 'bash ./scripts/migrate-casino-main.sh' },
    @{ Name = 'Casino microservice migrations'; Enabled = $true; Command = 'bash ./scripts/migrate-casino.sh' },
    @{ Name = 'Casino service tests'; Enabled = (-not $SkipTests); Command = 'go test ./services/casino/...' },
    @{ Name = 'Casino health smoke'; Enabled = (-not $SkipSmoke); Command = 'curl -fsS http://127.0.0.1:8082/healthz' }
)

if ($FullSmoke -and -not $SkipTests) {
    $repoSmoke = @{ Name = 'Repo-wide smoke tests'; Enabled = $true; Command = 'go test ./...' }
    $steps = @($steps[0], $steps[1], $steps[2], $steps[3], $repoSmoke, $steps[4])
}

foreach ($step in $steps) {
    if (-not $step.Enabled) {
        continue
    }

    Write-Host ""
    Write-Host "==> $($step.Name)"
    Write-Host "    $($step.Command)"

    if ($PrintOnly) {
        continue
    }

    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $runner $step.Command
    if ($LASTEXITCODE -ne 0) {
        throw "Step failed: $($step.Name)"
    }
}

Write-Host ""
Write-Host "Casino DB rollout checklist finished."
