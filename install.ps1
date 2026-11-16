param(
    [string]$Version = $env:AIC_VERSION,
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\aic"
)

$ErrorActionPreference = "Stop"

$Repo = "eliel-elie/ai-commits"
if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = "latest"
}

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "ARM64" { "arm64" }
    default { "amd64" }
}

$Asset = "aic-windows-$Arch.exe"
if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$Dest = Join-Path $InstallDir "aic.exe"

Write-Host "Downloading $Asset ($Version)..."
Invoke-WebRequest -Uri $Url -OutFile $Dest -UseBasicParsing

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$PathEntries = @()
if ($UserPath) {
    $PathEntries = $UserPath -split ";" | Where-Object { $_ -ne "" }
}

if ($PathEntries -notcontains $InstallDir) {
    $NewPath = ($PathEntries + $InstallDir) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "Added $InstallDir to your User PATH (reopen new terminals if needed)."
}

if (($env:PATH -split ";") -notcontains $InstallDir) {
    $env:PATH = "$env:PATH;$InstallDir"
}

Write-Host "Installed to $Dest"
& $Dest --version
