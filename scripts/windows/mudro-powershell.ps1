Set-StrictMode -Version Latest

function Get-MudroRepoRoot {
    if ($PSScriptRoot) {
        $candidate = Resolve-Path (Join-Path $PSScriptRoot "..\..")
        return $candidate.Path
    }
    return (Get-Location).Path
}

function Convert-MudroPathToBash {
    param([Parameter(Mandatory = $true)][string]$WindowsPath)

    $resolved = (Resolve-Path $WindowsPath).Path
    $drive = $resolved.Substring(0, 1).ToLowerInvariant()
    $rest = $resolved.Substring(2).Replace('\', '/')
    return "/$drive$rest"
}

function Convert-MudroPathToWsl {
    param(
        [Parameter(Mandatory = $true)][string]$WindowsPath,
        [string]$Distro = $env:MUDRO_WSL_DISTRO
    )

    if ($Distro) {
        return (& wsl -d $Distro wslpath -a $WindowsPath).Trim()
    }
    return (& wsl wslpath -a $WindowsPath).Trim()
}

function Get-MudroGitBashPath {
    $candidates = @(
        "C:\Program Files\Git\bin\bash.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    return $null
}

function Invoke-MudroShell {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)][string]$Command,
        [string]$RepoRoot = (Get-MudroRepoRoot),
        [ValidateSet('gitbash', 'wsl')][string]$Engine = $(if ($env:MUDRO_SHELL_ENGINE) { $env:MUDRO_SHELL_ENGINE } else { 'gitbash' }),
        [string]$Distro = $env:MUDRO_WSL_DISTRO
    )

    if ($Engine -eq 'gitbash') {
        $bashPath = Get-MudroGitBashPath
        if (-not $bashPath) {
            throw "Git Bash is not installed. Install Git for Windows or set MUDRO_SHELL_ENGINE=wsl."
        }

        $repoBash = Convert-MudroPathToBash -WindowsPath $RepoRoot
        $bashCmd = "cd '$repoBash' && $Command"
        & $bashPath -lc $bashCmd
        return
    }

    if (-not (Get-Command wsl -ErrorAction SilentlyContinue)) {
        throw "WSL is not installed or not available in PATH."
    }

    $repoWsl = Convert-MudroPathToWsl -WindowsPath $RepoRoot -Distro $Distro
    $bashCmd = "cd '$repoWsl' && $Command"

    if ($Distro) {
        & wsl -d $Distro -e bash -lc $bashCmd
    }
    else {
        & wsl -e bash -lc $bashCmd
    }
}

function mudro-make {
    [CmdletBinding()]
    param(
        [Parameter(ValueFromRemainingArguments = $true)][string[]]$Args
    )

    $target = if ($Args.Count -gt 0) { $Args -join ' ' } else { 'help' }
    Invoke-MudroShell -Command "make $target"
}

function mudro-health { Invoke-MudroShell -Command "make health" }
function mudro-up { Invoke-MudroShell -Command "make up" }
function mudro-down { Invoke-MudroShell -Command "make down" }
function mudro-ps { Invoke-MudroShell -Command "make ps" }
function mudro-logs { Invoke-MudroShell -Command "make logs" }
function mudro-dbcheck { Invoke-MudroShell -Command "make dbcheck" }
function mudro-migrate { Invoke-MudroShell -Command "make migrate" }
function mudro-tables { Invoke-MudroShell -Command "make tables" }
function mudro-test { Invoke-MudroShell -Command "go test ./..." }
function mudro-e2e { Invoke-MudroShell -Command "go test ./e2e -run TestCmd -count=1" }
function mudro-shell { Invoke-MudroShell -Command "exec bash" }
function mudro-casino-rollout {
    [CmdletBinding()]
    param(
        [Parameter(ValueFromRemainingArguments = $true)][string[]]$Args
    )

    $repoRoot = Get-MudroRepoRoot
    $scriptPath = Join-Path $repoRoot "scripts\windows\casino-db-rollout.ps1"
    & powershell.exe -NoProfile -ExecutionPolicy Bypass -File $scriptPath @Args
}

Set-Alias mmake mudro-make -Scope Global
Set-Alias mh mudro-health -Scope Global
Set-Alias mt mudro-test -Scope Global
Set-Alias me2e mudro-e2e -Scope Global
Set-Alias mcasino mudro-casino-rollout -Scope Global

function Enable-MudroPowerShellProfile {
    [CmdletBinding()]
    param(
        [string]$ProfilePath = $PROFILE.CurrentUserCurrentHost,
        [string]$ScriptPath = (Join-Path (Get-MudroRepoRoot) "scripts\windows\mudro-powershell.ps1")
    )

    if (-not (Test-Path $ProfilePath)) {
        New-Item -ItemType File -Force -Path $ProfilePath | Out-Null
    }

    $line = ". '$ScriptPath'"
    $current = Get-Content -Raw $ProfilePath -ErrorAction SilentlyContinue
    if ($null -eq $current -or $current -notmatch [regex]::Escape($line)) {
        Add-Content -Path $ProfilePath -Value "`n# mudro helpers`n$line`n"
    }
}
