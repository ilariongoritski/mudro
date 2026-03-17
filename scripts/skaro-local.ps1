param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$SkaroArgs
)

$toolRoot = 'D:\mudr\toolchain'
$binDir = Join-Path $toolRoot 'uv-bin'
$toolDir = Join-Path $toolRoot 'uv-tools'
$cacheDir = Join-Path $toolRoot 'uv-cache'
$pythonDir = Join-Path $toolRoot 'uv-python'

$env:UV_TOOL_DIR = $toolDir
$env:UV_TOOL_BIN_DIR = $binDir
$env:UV_CACHE_DIR = $cacheDir
$env:UV_PYTHON_INSTALL_DIR = $pythonDir
$env:UV_PYTHON_BIN_DIR = $binDir
$env:PYTHONUTF8 = '1'
$env:PYTHONIOENCODING = 'utf-8'
$env:PATH = "$binDir;$env:PATH"

[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new()
$OutputEncoding = [System.Text.UTF8Encoding]::new()

& skaro @SkaroArgs
exit $LASTEXITCODE
