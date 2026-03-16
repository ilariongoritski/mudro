. "$PSScriptRoot\common.ps1"

Ensure-MudroCommand "npx.cmd"

$secretsDir = Get-MudroSecretsDir
$envFile = Join-Path $secretsDir "magic-21st.local.env"

New-MudroEnvFileIfMissing $envFile @(
    "# 21st.dev Magic API key"
    "# Generate it in https://21st.dev/magic/console"
    "TWENTY_FIRST_API_KEY="
)

Import-MudroEnvFile $envFile

if ($env:TWENTY_FIRST_API_KEY) {
    $env:API_KEY = $env:TWENTY_FIRST_API_KEY
}

& npx.cmd "-y" "@21st-dev/magic@0.1.0"
exit $LASTEXITCODE
