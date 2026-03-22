param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$SkaroArgs
)

$toolRoot = 'D:\mudr\toolchain'
$localRoot = 'D:\mudr\_mudro-local\skaro'
$claudeOrchRoot = Join-Path 'D:\mudr\_mudro-local' 'claude-orch'
$localEnvFile = Join-Path $localRoot 'claude.env'
$binDir = Join-Path $toolRoot 'uv-bin'
$toolDir = Join-Path $toolRoot 'uv-tools'
$cacheDir = Join-Path $toolRoot 'uv-cache'
$pythonDir = Join-Path $toolRoot 'uv-python'

function Import-LocalEnvFile {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    if (-not (Test-Path -Path $Path)) {
        return
    }

    Get-Content -Path $Path -Encoding UTF8 | ForEach-Object {
        $line = $_.Trim()
        if (-not $line -or $line.StartsWith('#')) {
            return
        }

        $pair = $line -split '=', 2
        if ($pair.Count -ne 2) {
            return
        }

        $name = $pair[0].Trim()
        $value = $pair[1].Trim()

        if (
            ($value.StartsWith('"') -and $value.EndsWith('"')) -or
            ($value.StartsWith("'") -and $value.EndsWith("'"))
        ) {
            $value = $value.Substring(1, $value.Length - 2)
        }

        [Environment]::SetEnvironmentVariable($name, $value, 'Process')
    }
}

New-Item -ItemType Directory -Force -Path $localRoot | Out-Null
New-Item -ItemType Directory -Force -Path $claudeOrchRoot | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $claudeOrchRoot 'ledger') | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $claudeOrchRoot 'runs') | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $claudeOrchRoot 'state') | Out-Null
Import-LocalEnvFile -Path $localEnvFile

$env:UV_TOOL_DIR = $toolDir
$env:UV_TOOL_BIN_DIR = $binDir
$env:UV_CACHE_DIR = $cacheDir
$env:UV_PYTHON_INSTALL_DIR = $pythonDir
$env:UV_PYTHON_BIN_DIR = $binDir
$env:PYTHONUTF8 = '1'
$env:PYTHONIOENCODING = 'utf-8'
$env:TZ = 'Europe/Moscow'
$env:SKARO_LOCAL_ROOT = $localRoot
$env:MUDRO_CLAUDE_ACCOUNTING_ROOT = $claudeOrchRoot
$env:MUDRO_CLAUDE_USAGE_LOG = Join-Path $claudeOrchRoot 'ledger\usage_log.jsonl'
$env:MUDRO_CLAUDE_TOKEN_USAGE = Join-Path $claudeOrchRoot 'ledger\token_usage.yaml'
$env:MUDRO_CLAUDE_ROLE_USAGE = Join-Path $claudeOrchRoot 'ledger\role_usage.yaml'
$env:MUDRO_CLAUDE_STATE_DIR = Join-Path $claudeOrchRoot 'state'
$env:MUDRO_CLAUDE_RUNS_DIR = Join-Path $claudeOrchRoot 'runs'

if (-not $env:ANTHROPIC_BASE_URL) {
    $env:ANTHROPIC_BASE_URL = 'https://claude-api.filips-site.online'
}

if ($env:ANTHROPIC_API_KEY -and -not $env:CLAUDE_API_KEY) {
    $env:CLAUDE_API_KEY = $env:ANTHROPIC_API_KEY
}

$env:PATH = "$binDir;$env:PATH"

[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()
$OutputEncoding = [System.Text.UTF8Encoding]::new()

& skaro @SkaroArgs
exit $LASTEXITCODE