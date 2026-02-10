# ClearC - 跨平台系统清理工具

ClearC 是一款专为开发者设计的跨平台系统清理工具，可以帮助您快速清理开发垃圾文件，释放磁盘空间。

## 功能特性

- **开发垃圾清理**: 自动识别并清理 node_modules、Go 模块缓存、Python 缓存、Rust target 等开发相关垃圾
- **系统临时文件清理**: 清理系统临时文件和浏览器缓存
- **未使用文件检测**: 扫描长期未访问的文件，提供清理建议
- **系统托盘常驻**: 最小化到系统托盘，随时快速访问
- **跨平台支持**: 支持 Windows、macOS 和 Linux

## 技术栈

- **后端**: Go + Wails v2
- **前端**: React + TypeScript + Tailwind CSS
- **系统托盘**: fyne.io/systray
- **系统信息**: gopsutil

## 开发环境要求

- Go 1.22+
- Node.js 18+
- Wails CLI v2

## 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## 开发

```bash
# 安装前端依赖
cd frontend && npm install && cd ..

# 启动开发模式
wails dev
```

## 构建

### Windows

```powershell
# 设置环境变量
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# 构建
wails build
```

### macOS

```bash
# Intel Mac
wails build -platform darwin/amd64

# Apple Silicon
wails build -platform darwin/arm64
```

### Linux

```bash
wails build -platform linux/amd64
```

## 项目结构

```
clearc/
├── main.go                 # 应用入口
├── app.go                  # 应用逻辑和前端绑定
├── internal/
│   ├── platform/           # 跨平台抽象层
│   │   ├── platform.go     # 通用接口
│   │   ├── windows.go      # Windows 实现
│   │   ├── darwin.go       # macOS 实现
│   │   └── linux.go        # Linux 实现
│   ├── scanner/            # 磁盘扫描器
│   │   ├── scanner.go      # 扫描逻辑
│   │   ├── atime_windows.go
│   │   └── atime_unix.go
│   ├── cleaner/            # 文件清理器
│   │   └── cleaner.go
│   ├── config/             # 配置管理
│   │   └── config.go
│   └── tray/               # 系统托盘
│       ├── tray.go
│       └── icon.go
├── frontend/               # React 前端
│   ├── src/
│   │   ├── components/     # UI 组件
│   │   ├── types/          # TypeScript 类型
│   │   └── App.tsx         # 主应用
│   └── package.json
├── build/                  # 构建输出
└── wails.json              # Wails 配置
```

## 支持的清理类型

### 开发缓存
- Node.js: `node_modules`, `.npm`, `.pnpm-store`
- Go: `go/pkg/mod/cache`
- Python: `__pycache__`, `.pytest_cache`, `.venv`
- Rust: `target` (debug builds)
- IDE: `.idea`, `.vscode` 缓存

### 系统文件
- Windows 临时文件
- 浏览器缓存 (Chrome, Edge, Firefox)
- 系统日志

## 配置文件

配置文件位置:
- Windows: `%APPDATA%\ClearC\config.json`
- macOS: `~/Library/Application Support/ClearC/config.json`
- Linux: `~/.config/ClearC/config.json`

## 许可证

MIT License
