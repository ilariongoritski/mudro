Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$packages = @(
    @{ Id = 'GoLang.Go'; Name = 'Go' },
    @{ Id = 'BurntSushi.ripgrep.MSVC'; Name = 'ripgrep' },
    @{ Id = 'ezwinports.make'; Name = 'make' }
)

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
$binDir = Join-Path $env:USERPROFILE 'bin'
New-Item -ItemType Directory -Force -Path $binDir | Out-Null

$mudroCmd = @"
@echo off
setlocal
set "PATH=C:\Program Files\Go\bin;%PATH%"
set "PATH=%LOCALAPPDATA%\Microsoft\WinGet\Packages\ezwinports.make_Microsoft.Winget.Source_8wekyb3d8bbwe\bin;%PATH%"
set "BASH=D:\Git\bin\bash.exe"
if not exist "%BASH%" set "BASH=C:\Program Files\Git\bin\bash.exe"
if not exist "%BASH%" (
  echo [mudro] Git Bash not found. Install Git for Windows.
  exit /b 1
)
"%BASH%" -lc "cd '$repoBash' && %*"
"@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mudro.cmd') -Value $mudroCmd

Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mmake.cmd') -Value "@echo off`r`nmudro make %*`r`n"
$mhCmd = @'
@echo off
setlocal
call mmake up || exit /b 1
call mps || exit /b 1
call mmake dbcheck || exit /b 1
call mmake migrate || exit /b 1
call mmake tables || exit /b 1
call mt || exit /b 1
call me2e || exit /b 1
call mmake count-posts || exit /b 1
'@
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'mh.cmd') -Value $mhCmd
$mtCmd = @'
@echo off
setlocal
set "PATH=C:\Program Files\Go\bin;%PATH%"
set "BASH=D:\Git\bin\bash.exe"
if not exist "%BASH%" set "BASH=C:\Program Files\Git\bin\bash.exe"
if not exist "%BASH%" exit /b 1
"%BASH%" -lc "cd '__REPO_BASH__' && PKGS=$(go list ./... | grep -v '/e2e$'); go test $PKGS"
'@
$mtCmd = $mtCmd.Replace('__REPO_BASH__', $repoBash)
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
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'go.cmd') -Value "@echo off`r`n""C:\Program Files\Go\bin\go.exe"" %*`r`n"
Set-Content -Encoding ASCII -Path (Join-Path $binDir 'make.cmd') -Value "@echo off`r`n""%LOCALAPPDATA%\Microsoft\WinGet\Packages\ezwinports.make_Microsoft.Winget.Source_8wekyb3d8bbwe\bin\make.exe"" %*`r`n"
$rgCmd = @'
@echo off
setlocal
for /f "delims=" %%F in ('dir /b /s "%LOCALAPPDATA%\Microsoft\WinGet\Packages\BurntSushi.ripgrep.MSVC_Microsoft.Winget.Source_8wekyb3d8bbwe\rg.exe" 2^>nul') do (
  "%%F" %*
  exit /b %ERRORLEVEL%
)
echo [rg] ripgrep not found in winget package directory.
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
