# Phanes DNA Installer Script for Windows (PowerShell)
# Usage: irm https://raw.githubusercontent.com/arleyS3/phanes-dna/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$InstallDir = "$env:LOCALAPPDATA\phanes-dna\bin"
$BinaryPath = "$InstallDir\phanes-dna.exe"

Write-Host "🧬 Installing Phanes DNA on Windows..." -ForegroundColor Cyan

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

if (Get-Command "go" -ErrorAction SilentlyContinue) {
    Write-Host "🔨 Building phanes-dna using Go..." -ForegroundColor Yellow
    go install github.com/arleyS3/phanes-dna/cmd/phanes-dna@latest
    $GoBin = "$env:GOPATH\bin\phanes-dna.exe"
    if (-not (Test-Path $GoBin)) { $GoBin = "$env:USERPROFILE\go\bin\phanes-dna.exe" }
    if (Test-Path $GoBin) {
        Copy-Item -Path $GoBin -Destination $BinaryPath -Force
    }
} else {
    Write-Host "⚠️ Go not found. Downloading binary release..." -ForegroundColor Yellow
    $DownloadUrl = "https://github.com/arleyS3/phanes-dna/releases/latest/download/phanes-dna-windows-amd64.zip"
    $ZipPath = "$env:TEMP\phanes-dna.zip"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath
    Expand-Archive -Path $ZipPath -DestinationPath $InstallDir -Force
    Remove-Item $ZipPath -Force
}

# Add to User PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "📝 Adding $InstallDir to User PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
}

Write-Host "✅ Phanes DNA installed successfully to $BinaryPath!" -ForegroundColor Green
Write-Host "🚀 Run 'phanes-dna' to launch the interactive terminal UI, or 'phanes-dna doctor' for health checks." -ForegroundColor Green
