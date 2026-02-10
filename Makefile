# ClearC Makefile
# 用于 macOS/Linux 本地构建

APP_NAME := ClearC
VERSION ?= 1.0.0
OUTPUT_DIR := dist

.PHONY: all clean windows macos linux help

help:
	@echo ""
	@echo "ClearC Build System"
	@echo "==================="
	@echo ""
	@echo "用法: make [target] [VERSION=x.x.x]"
	@echo ""
	@echo "目标:"
	@echo "  windows    构建 Windows 版本"
	@echo "  macos      构建 macOS 版本 (amd64 + arm64)"
	@echo "  linux      构建 Linux 版本"
	@echo "  all        构建所有平台"
	@echo "  clean      清理构建产物"
	@echo ""
	@echo "示例:"
	@echo "  make macos VERSION=1.2.0"
	@echo "  make all"
	@echo ""

all: windows macos linux

clean:
	@echo "清理构建产物..."
	rm -rf $(OUTPUT_DIR)
	rm -rf build/bin/*

$(OUTPUT_DIR):
	mkdir -p $(OUTPUT_DIR)

# Windows
windows: $(OUTPUT_DIR)
	@echo "构建 Windows (amd64)..."
	wails build -platform windows/amd64
	cd build/bin && zip -r "../../$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip" *.exe
	@echo "✓ $(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip"

# macOS
macos: macos-amd64 macos-arm64

macos-amd64: $(OUTPUT_DIR)
	@echo "构建 macOS (amd64)..."
	wails build -platform darwin/amd64
	cd build/bin && zip -r "../../$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-amd64.zip" *.app
	@echo "✓ $(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-amd64.zip"

macos-arm64: $(OUTPUT_DIR)
	@echo "构建 macOS (arm64)..."
	wails build -platform darwin/arm64
	cd build/bin && zip -r "../../$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-arm64.zip" *.app
	@echo "✓ $(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-arm64.zip"

macos-universal: $(OUTPUT_DIR)
	@echo "构建 macOS Universal..."
	wails build -platform darwin/universal
	cd build/bin && zip -r "../../$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-universal.zip" *.app
	@echo "✓ $(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-macos-universal.zip"

# Linux
linux: $(OUTPUT_DIR)
	@echo "构建 Linux (amd64)..."
	wails build -platform linux/amd64
	cd build/bin && tar -czvf "../../$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz" *
	@echo "✓ $(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz"

# 开发
dev:
	wails dev

# 检查依赖
check:
	@echo "检查构建依赖..."
	@which go > /dev/null || (echo "❌ Go 未安装" && exit 1)
	@echo "✓ Go: $$(go version)"
	@which wails > /dev/null || (echo "❌ Wails 未安装" && exit 1)
	@echo "✓ Wails: $$(wails version)"
	@which node > /dev/null || (echo "❌ Node.js 未安装" && exit 1)
	@echo "✓ Node: $$(node --version)"
	@echo ""
	@echo "所有依赖已就绪！"
