# ClearC Release Build Script
# 用于本地构建发行版

param(
    [Parameter(Mandatory=$false)]
    [string]$Version = "1.0.0",
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("windows", "macos", "linux", "all")]
    [string]$Platform = "windows",
    
    [Parameter(Mandatory=$false)]
    [switch]$NSIS,
    
    [Parameter(Mandatory=$false)]
    [switch]$Clean
)

$ErrorActionPreference = "Stop"
$AppName = "ClearC"
$OutputDir = "dist"

Write-Host @"

  ╔═══════════════════════════════════════╗
  ║     ClearC Release Build Script       ║
  ║         Version: $Version                ║
  ╚═══════════════════════════════════════╝

"@ -ForegroundColor Cyan

# 检查 Wails
$WailsPath = Get-Command wails -ErrorAction SilentlyContinue
if (-not $WailsPath) {
    $WailsPath = "$env:USERPROFILE\go\bin\wails.exe"
    if (-not (Test-Path $WailsPath)) {
        Write-Host "错误: 未找到 Wails CLI，请先安装: go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Red
        exit 1
    }
} else {
    $WailsPath = $WailsPath.Source
}

# 清理
if ($Clean) {
    Write-Host "清理旧构建..." -ForegroundColor Yellow
    if (Test-Path $OutputDir) { Remove-Item -Recurse -Force $OutputDir }
    if (Test-Path "build\bin") { Remove-Item -Recurse -Force "build\bin\*" -ErrorAction SilentlyContinue }
}

# 创建输出目录
if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir | Out-Null
}

function Build-Platform {
    param(
        [string]$OS,
        [string]$Arch,
        [string]$Extra = ""
    )
    
    Write-Host "`n构建 $OS/$Arch..." -ForegroundColor Yellow
    
    $buildArgs = @("build", "-platform", "$OS/$Arch")
    if ($Extra) { $buildArgs += $Extra }
    
    & $WailsPath @buildArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "构建失败: $OS/$Arch" -ForegroundColor Red
        return $false
    }
    
    Write-Host "构建成功: $OS/$Arch" -ForegroundColor Green
    return $true
}

function Package-Windows {
    Write-Host "`n打包 Windows..." -ForegroundColor Yellow
    
    $nsisArg = if ($NSIS) { "-nsis" } else { "" }
    if (-not (Build-Platform "windows" "amd64" $nsisArg)) { return }
    
    $zipName = "$OutputDir\$AppName-$Version-windows-amd64.zip"
    Compress-Archive -Path "build\bin\*.exe" -DestinationPath $zipName -Force
    Write-Host "已创建: $zipName" -ForegroundColor Green
    
    # 如果有 NSIS 安装包
    $installer = Get-ChildItem "build\bin\*-amd64-installer.exe" -ErrorAction SilentlyContinue
    if ($installer) {
        $installerDest = "$OutputDir\$AppName-$Version-windows-amd64-installer.exe"
        Copy-Item $installer.FullName $installerDest
        Write-Host "已创建: $installerDest" -ForegroundColor Green
    }
}

function Package-MacOS {
    Write-Host "`n打包 macOS..." -ForegroundColor Yellow
    Write-Host "注意: 从 Windows 交叉编译 macOS 需要额外配置" -ForegroundColor DarkYellow
    
    foreach ($arch in @("amd64", "arm64")) {
        if (-not (Build-Platform "darwin" $arch)) { continue }
        
        $zipName = "$OutputDir\$AppName-$Version-macos-$arch.zip"
        Compress-Archive -Path "build\bin\*.app" -DestinationPath $zipName -Force
        Write-Host "已创建: $zipName" -ForegroundColor Green
    }
}

function Package-Linux {
    Write-Host "`n打包 Linux..." -ForegroundColor Yellow
    Write-Host "注意: 从 Windows 交叉编译 Linux 需要额外配置" -ForegroundColor DarkYellow
    
    if (-not (Build-Platform "linux" "amd64")) { return }
    
    # PowerShell 没有原生 tar，使用 7z 或 tar（如果可用）
    $tarPath = Get-Command tar -ErrorAction SilentlyContinue
    if ($tarPath) {
        Push-Location "build\bin"
        tar -czvf "..\..\$OutputDir\$AppName-$Version-linux-amd64.tar.gz" *
        Pop-Location
        Write-Host "已创建: $OutputDir\$AppName-$Version-linux-amd64.tar.gz" -ForegroundColor Green
    } else {
        $zipName = "$OutputDir\$AppName-$Version-linux-amd64.zip"
        Compress-Archive -Path "build\bin\*" -DestinationPath $zipName -Force
        Write-Host "已创建: $zipName (注: Linux 通常使用 .tar.gz)" -ForegroundColor Green
    }
}

# 执行构建
switch ($Platform.ToLower()) {
    "windows" { Package-Windows }
    "macos"   { Package-MacOS }
    "linux"   { Package-Linux }
    "all" {
        Package-Windows
        Write-Host "`n" -NoNewline
        Write-Host "═" * 50 -ForegroundColor DarkGray
        Write-Host @"

提示: 跨平台构建建议使用 GitHub Actions
      推送 tag 即可自动构建所有平台:
      
      git tag v$Version
      git push origin v$Version

"@ -ForegroundColor Cyan
    }
}

Write-Host "`n构建完成！输出目录: $OutputDir\" -ForegroundColor Cyan
Get-ChildItem $OutputDir | Format-Table Name, Length, LastWriteTime
