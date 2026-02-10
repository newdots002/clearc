# ClearC Build Script for Windows
# This script builds the application for multiple platforms

param(
    [string]$Platform = "all"
)

$WailsPath = "$env:USERPROFILE\go\bin\wails.exe"

# Ensure GOOS and GOARCH are set correctly for Windows builds
$env:GOOS = "windows"
$env:GOARCH = "amd64"

Write-Host "ClearC Build Script" -ForegroundColor Cyan
Write-Host "===================" -ForegroundColor Cyan

function Build-Windows {
    Write-Host "`nBuilding for Windows (amd64)..." -ForegroundColor Yellow
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    & $WailsPath build -platform windows/amd64
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Windows build successful!" -ForegroundColor Green
    } else {
        Write-Host "Windows build failed!" -ForegroundColor Red
    }
}

function Build-MacOS {
    Write-Host "`nBuilding for macOS (amd64)..." -ForegroundColor Yellow
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    & $WailsPath build -platform darwin/amd64
    if ($LASTEXITCODE -eq 0) {
        Write-Host "macOS (amd64) build successful!" -ForegroundColor Green
    } else {
        Write-Host "macOS (amd64) build failed!" -ForegroundColor Red
    }

    Write-Host "`nBuilding for macOS (arm64)..." -ForegroundColor Yellow
    $env:GOOS = "darwin"
    $env:GOARCH = "arm64"
    & $WailsPath build -platform darwin/arm64
    if ($LASTEXITCODE -eq 0) {
        Write-Host "macOS (arm64) build successful!" -ForegroundColor Green
    } else {
        Write-Host "macOS (arm64) build failed!" -ForegroundColor Red
    }
}

function Build-Linux {
    Write-Host "`nBuilding for Linux (amd64)..." -ForegroundColor Yellow
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    & $WailsPath build -platform linux/amd64
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Linux build successful!" -ForegroundColor Green
    } else {
        Write-Host "Linux build failed!" -ForegroundColor Red
    }
}

switch ($Platform.ToLower()) {
    "windows" { Build-Windows }
    "macos" { Build-MacOS }
    "linux" { Build-Linux }
    "all" {
        Build-Windows
        # Note: Cross-compilation for macOS and Linux from Windows
        # may require additional setup (CGO, cross-compilers)
        Write-Host "`nNote: Cross-compilation for macOS/Linux from Windows" -ForegroundColor Yellow
        Write-Host "requires additional toolchains (CGO, cross-compilers)." -ForegroundColor Yellow
        Write-Host "For production builds, use native build environments." -ForegroundColor Yellow
    }
    default {
        Write-Host "Usage: .\build.ps1 [-Platform <windows|macos|linux|all>]" -ForegroundColor White
    }
}

Write-Host "`nBuild output location: build\bin\" -ForegroundColor Cyan
