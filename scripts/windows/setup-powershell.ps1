Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$packages = @(
    @{ Id = 'GoLang.Go'; Name = 'Go' },
    @{ Id = 'BurntSushi.ripgrep.MSVC'; Name = 'ripgrep' },
    @{ Id = 'ezwinports.make'; Name = 'make' }
)

$winget = Get-Command winget -ErrorAction SilentlyContinue
if ($winget) {
    foreach ($p in $packages) {
        $installed = winget list --id $p.Id --exact 2>$null
        if ($LASTEXITCODE -ne 0 -or -not $installed) {
            Write-Host "Installing $($p.Name) ($($p.Id))..."
            winget install -e --id $p.Id --accept-package-agreements --accept-source-agreements
        }
        else {
            Write-Host "$($p.Name) already installed."
        }
    }
}
else {
    Write-Host "winget not found. Skipping package installation step."
}

$env:Path = [System.Environment]::GetEnvironmentVariable('Path', 'Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path', 'User')

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

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$repoBash = Convert-WindowsPathToBash $repoRoot
$repoWsl = Convert-WindowsPathToWsl $repoRoot
$runHelper = (Join-Path $repoRoot 'scripts\windows\run-mudro-no-profile.ps1')
$binDir = Join-Path $env:USERPROFILE 'bin'
New-Item -ItemType Directory -Force -Path $binDir | Out-Null

$mudroCmd = @"
@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$runHelper" %*
"@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mudro.cmd') -Value $mudroCmd

Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mmake.cmd') -Value "@echo off`r`nmudro make %*`r`n"
$mhCmd = @'
@echo off
setlocal
call mmake core-up || exit /b 1
call mmake core-ps || exit /b 1
call mmake dbcheck-core || exit /b 1
call mmake migrate-runtime || exit /b 1
call mmake tables-core || exit /b 1
call mt || exit /b 1
call me2e || exit /b 1
call mmake count-posts-core || exit /b 1
'@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mh.cmd') -Value $mhCmd
$mtCmd = @"
@echo off
setlocal
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "$runHelper" make test-active
"@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mt.cmd') -Value $mtCmd
$me2eCmd = @"
@echo off
setlocal
if not "%MUDRO_WSL_DISTRO%"=="" (
  wsl -d %MUDRO_WSL_DISTRO% -e bash -lc "cd '$repoWsl' && go test ./e2e -run TestCmd -count=1"
) else (
  wsl -e bash -lc "cd '$repoWsl' && go test ./e2e -run TestCmd -count=1"
)
"@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'me2e.cmd') -Value $me2eCmd
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mps.cmd') -Value "@echo off`r`nmudro make ps`r`n"
$goCmd = @'
@echo off
setlocal
for %%F in (go.exe) do set "GO_BIN=%%~$PATH:F"
if not defined GO_BIN if exist "C:\Program Files\Go\bin\go.exe" set "GO_BIN=C:\Program Files\Go\bin\go.exe"
if not defined GO_BIN (
  echo [go] go.exe not found in PATH.
  exit /b 1
)
"%GO_BIN%" %*
'@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'go.cmd') -Value $goCmd
$makeCmd = @'
@echo off
setlocal
for %%F in (make.exe) do set "MAKE_BIN=%%~$PATH:F"
if not defined MAKE_BIN if exist "%LOCALAPPDATA%\Microsoft\WinGet\Packages\ezwinports.make_Microsoft.Winget.Source_8wekyb3d8bbwe\bin\make.exe" set "MAKE_BIN=%LOCALAPPDATA%\Microsoft\WinGet\Packages\ezwinports.make_Microsoft.Winget.Source_8wekyb3d8bbwe\bin\make.exe"
if not defined MAKE_BIN (
  echo [make] make.exe not found in PATH.
  exit /b 1
)
"%MAKE_BIN%" %*
'@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'make.cmd') -Value $makeCmd
$rgCmd = @'
@echo off
setlocal
for %%F in (rg.exe) do set "RG_BIN=%%~$PATH:F"
if defined RG_BIN (
  "%RG_BIN%" %*
  exit /b %ERRORLEVEL%
)
for /f "delims=" %%F in ('dir /b /s "%LOCALAPPDATA%\Microsoft\WinGet\Packages\BurntSushi.ripgrep.MSVC_Microsoft.Winget.Source_8wekyb3d8bbwe\rg.exe" 2^>nul') do (
  "%%F" %*
  exit /b %ERRORLEVEL%
)
echo [rg] rg.exe not found in PATH or winget package directory.
exit /b 1
'@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'rg.cmd') -Value $rgCmd

Write-Host ""
Write-Host "Tools status:"
Write-Host "go    => $((Get-Command go -ErrorAction SilentlyContinue).Source)"
Write-Host "rg    => $((Get-Command rg -ErrorAction SilentlyContinue).Source)"
Write-Host "make  => $((Get-Command make -ErrorAction SilentlyContinue).Source)"
Write-Host ""
Write-Host "Windows wrappers created in $binDir"
Write-Host "Commands: mudro, mmake, mh, mt, me2e, mps"
Write-Host "Example: mmake up"
