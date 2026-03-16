Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Get-MudroRepoRoot {
    $root = Resolve-Path (Join-Path $PSScriptRoot "..\..\..\")
    return $root.Path
}

function Get-MudroCodexHome {
    if ($env:CODEX_HOME) {
        return $env:CODEX_HOME
    }

    return (Join-Path $env:USERPROFILE ".codex")
}

function Get-MudroSecretsDir {
    $dir = Join-Path (Get-MudroCodexHome) "secrets"
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir | Out-Null
    }

    return $dir
}

function Import-MudroEnvFile([string]$Path) {
    if (-not (Test-Path $Path)) {
        return
    }

    foreach ($line in Get-Content -Path $Path) {
        if ([string]::IsNullOrWhiteSpace($line)) {
            continue
        }

        $trimmed = $line.Trim()
        if ($trimmed.StartsWith("#")) {
            continue
        }

        $parts = $line -split "=", 2
        if ($parts.Count -ne 2) {
            continue
        }

        $name = $parts[0].Trim()
        $value = $parts[1].Trim()

        if ($value.Length -ge 2) {
            $singleQuoted = $value.StartsWith("'") -and $value.EndsWith("'")
            $doubleQuoted = $value.StartsWith('"') -and $value.EndsWith('"')
            if ($singleQuoted -or $doubleQuoted) {
                $value = $value.Substring(1, $value.Length - 2)
            }
        }

        [Environment]::SetEnvironmentVariable($name, $value, "Process")
    }
}

function Ensure-MudroCommand([string]$Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command not found in PATH: $Name"
    }
}

function New-MudroEnvFileIfMissing([string]$Path, [string[]]$Lines) {
    if (Test-Path $Path) {
        return
    }

    $parent = Split-Path -Parent $Path
    if (-not (Test-Path $parent)) {
        New-Item -ItemType Directory -Path $parent | Out-Null
    }

    Set-Content -Path $Path -Value $Lines -Encoding ascii
}
