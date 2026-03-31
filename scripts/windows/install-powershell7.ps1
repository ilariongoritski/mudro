Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (Get-Command pwsh.exe -ErrorAction SilentlyContinue) {
    Write-Host "PowerShell 7 is already installed."
    exit 0
}

if (-not (Get-Command winget.exe -ErrorAction SilentlyContinue)) {
    Write-Error "winget.exe was not found. Install App Installer from Microsoft Store first."
    exit 1
}

Write-Host "Installing PowerShell 7 via winget..."
winget install --id Microsoft.PowerShell --source winget --accept-package-agreements --accept-source-agreements

if (Get-Command pwsh.exe -ErrorAction SilentlyContinue) {
    Write-Host "PowerShell 7 installed successfully."
    Write-Host "Restart VS Code or open a new terminal session."
    exit 0
}

Write-Error "Installation finished, but pwsh.exe is still not visible in PATH. Restart the session and check again."
exit 1
