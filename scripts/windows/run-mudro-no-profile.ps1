param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$CommandParts = @()
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

if ($CommandParts.Count -eq 0) {
    Write-Error "Command is required. Example: run-mudro-no-profile.ps1 make health-runtime"
    exit 1
}

function Convert-WindowsPathToBash([string]$PathValue) {
    $resolved = (Resolve-Path $PathValue).Path
    $drive = $resolved.Substring(0, 1).ToLowerInvariant()
    $rest = $resolved.Substring(2).Replace('\\', '/').Replace('\', '/')
    return "/$drive$rest"
}

function Convert-WindowsPathToWsl([string]$PathValue) {
    $resolved = (Resolve-Path $PathValue).Path
    $drive = $resolved.Substring(0, 1).ToLowerInvariant()
    $rest = $resolved.Substring(2).Replace('\\', '/').Replace('\', '/')
    return "/mnt/$drive$rest"
}

function Get-GitBashPath {
    $candidates = @(
        $env:MUDRO_GIT_BASH,
        "C:\Program Files\Git\bin\bash.exe"
    ) | Where-Object { $_ -and $_.Trim().Length -gt 0 }

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    return $null
}

function Invoke-ThroughWsl([string]$RepoRoot, [string]$Cmd) {
    if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
        return $false
    }

    $repoWsl = Convert-WindowsPathToWsl $RepoRoot
    $bootstrap = "export DOCKER_CONFIG=`$HOME/.docker-mudro; mkdir -p `$DOCKER_CONFIG; if [ ! -f `$DOCKER_CONFIG/config.json ]; then echo '{}' > `$DOCKER_CONFIG/config.json; fi"
    $bashCmd = "$bootstrap; cd '$repoWsl' && $Cmd"
    $distro = $env:MUDRO_WSL_DISTRO
    if ($distro -and $distro.Trim().Length -gt 0) {
        & wsl -d $distro -e bash -lc $bashCmd | Out-Host
    }
    else {
        & wsl -e bash -lc $bashCmd | Out-Host
    }
    return $LASTEXITCODE -eq 0
}

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$repoBash = Convert-WindowsPathToBash $repoRoot
$bashPath = Get-GitBashPath

$command = $CommandParts -join ' '
$engine = $env:MUDRO_HELPER_ENGINE
if (-not $engine -or $engine.Trim().Length -eq 0) {
    $engine = "wsl"
}

if ($engine -eq "wsl") {
    if (Invoke-ThroughWsl $repoRoot $command) {
        exit 0
    }
    Write-Error "WSL execution failed."
    exit $LASTEXITCODE
}

if (-not $bashPath) {
    Write-Error "Git Bash not found. Set MUDRO_GIT_BASH or use MUDRO_HELPER_ENGINE=wsl."
    exit 1
}

& $bashPath --noprofile --norc -lc "cd '$repoBash' && $command"
if ($LASTEXITCODE -eq 0) {
    exit 0
}

Write-Warning "Git Bash execution failed, trying WSL fallback."
if (Invoke-ThroughWsl $repoRoot $command) {
    exit 0
}

Write-Error "Git Bash and WSL execution failed."
exit 1
