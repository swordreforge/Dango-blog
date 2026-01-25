# Makefile for myblog-gogogo - Go 交叉编译

.PHONY: help build-all build-windows build-linux build-mac build-freebsd build-openbsd clean test

# 项目名称
PROJECT_NAME := myblog-gogogo
VERSION := $(shell date +%Y%m%d)
OUTPUT_DIR := build

# 通用编译参数
LDFLAGS := -s -w -X main.Version=$(VERSION)
BUILD_FLAGS := --trimpath -tags=netgo

help:
	@echo "Go 交叉编译命令:"
	@echo ""
	@echo "  make build-all        - 编译所有平台"
	@echo "  make build-windows    - 编译 Windows (amd64 + 386 + arm64)"
	@echo "  make build-linux      - 编译 Linux (amd64 + arm64 + 386 + riscv64)"
	@echo "  make build-mac        - 编译 macOS (amd64 + arm64)"
	@echo "  make build-freebsd    - 编译 FreeBSD (amd64 + arm64)"
	@echo "  make build-openbsd    - 编译 OpenBSD (amd64 + arm64)"
	@echo "  make clean            - 清理编译输出"
	@echo "  make test             - 运行测试"
	@echo ""
	@echo "编译特性:"
	@echo "  ✓ 交叉编译 (无需目标平台环境)"
	@echo "  ✓ 移除符号表 (-s -w)"
	@echo "  ✓ UPX 最佳压缩 (Windows)"
	@echo "  ✓ 纯 Go 网络栈 (netgo)"

build-all: build-windows build-linux build-mac build-freebsd build-openbsd
	@echo ""
	@echo "========================================="
	@echo "所有平台编译完成!"
	@echo "输出目录: $(OUTPUT_DIR)/"
	@echo "========================================="

# Windows 平台
build-windows: build-windows-amd64 build-windows-386 build-windows-arm64
	@echo ""
	@echo "✓ Windows 所有版本编译完成"

build-windows-amd64:
	@echo "编译 Windows amd64..."
	@mkdir -p $(OUTPUT_DIR)/windows
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-amd64.exe
	@which upx > /dev/null && upx --best --lzma $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-amd64.exe || echo "  (跳过 UPX 压缩)"
	@echo "✓ $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-amd64.exe"

build-windows-386:
	@echo "编译 Windows 386..."
	@mkdir -p $(OUTPUT_DIR)/windows
	@GOOS=windows GOARCH=386 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-386.exe
	@which upx > /dev/null && upx --best --lzma $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-386.exe || echo "  (跳过 UPX 压缩)"
	@echo "✓ $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-386.exe"

build-windows-arm64:
	@echo "编译 Windows arm64..."
	@mkdir -p $(OUTPUT_DIR)/windows
	@GOOS=windows GOARCH=arm64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-arm64.exe
	@which upx > /dev/null && upx --best --lzma $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-arm64.exe || echo "  (跳过 UPX 压缩)"
	@echo "✓ $(OUTPUT_DIR)/windows/$(PROJECT_NAME)-arm64.exe"

# Linux 平台
build-linux: build-linux-amd64 build-linux-arm64 build-linux-386 build-linux-riscv64
	@echo ""
	@echo "✓ Linux 所有版本编译完成"

build-linux-amd64:
	@echo "编译 Linux amd64..."
	@mkdir -p $(OUTPUT_DIR)/linux
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-amd64
	@echo "✓ $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-amd64"

build-linux-arm64:
	@echo "编译 Linux arm64..."
	@mkdir -p $(OUTPUT_DIR)/linux
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-arm64
	@echo "✓ $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-arm64"

build-linux-386:
	@echo "编译 Linux 386..."
	@mkdir -p $(OUTPUT_DIR)/linux
	@GOOS=linux GOARCH=386 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-386
	@echo "✓ $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-386"

build-linux-riscv64:
	@echo "编译 Linux riscv64..."
	@mkdir -p $(OUTPUT_DIR)/linux
	@GOOS=linux GOARCH=riscv64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-riscv64
	@echo "✓ $(OUTPUT_DIR)/linux/$(PROJECT_NAME)-riscv64"

# macOS 平台
build-mac: build-mac-amd64 build-mac-arm64
	@echo ""
	@echo "✓ macOS 所有版本编译完成"

build-mac-amd64:
	@echo "编译 macOS amd64..."
	@mkdir -p $(OUTPUT_DIR)/darwin
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/darwin/$(PROJECT_NAME)-amd64
	@echo "✓ $(OUTPUT_DIR)/darwin/$(PROJECT_NAME)-amd64"

build-mac-arm64:
	@echo "编译 macOS arm64 (Apple Silicon)..."
	@mkdir -p $(OUTPUT_DIR)/darwin
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/darwin/$(PROJECT_NAME)-arm64
	@echo "✓ $(OUTPUT_DIR)/darwin/$(PROJECT_NAME)-arm64"

# FreeBSD 平台
build-freebsd: build-freebsd-amd64 build-freebsd-arm64
	@echo ""
	@echo "✓ FreeBSD 所有版本编译完成"

build-freebsd-amd64:
	@echo "编译 FreeBSD amd64..."
	@mkdir -p $(OUTPUT_DIR)/freebsd
	@GOOS=freebsd GOARCH=amd64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/freebsd/$(PROJECT_NAME)-amd64
	@echo "✓ $(OUTPUT_DIR)/freebsd/$(PROJECT_NAME)-amd64"

build-freebsd-arm64:
	@echo "编译 FreeBSD arm64..."
	@mkdir -p $(OUTPUT_DIR)/freebsd
	@GOOS=freebsd GOARCH=arm64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/freebsd/$(PROJECT_NAME)-arm64
	@echo "✓ $(OUTPUT_DIR)/freebsd/$(PROJECT_NAME)-arm64"

# OpenBSD 平台
build-openbsd: build-openbsd-amd64 build-openbsd-arm64
	@echo ""
	@echo "✓ OpenBSD 所有版本编译完成"

build-openbsd-amd64:
	@echo "编译 OpenBSD amd64..."
	@mkdir -p $(OUTPUT_DIR)/openbsd
	@GOOS=openbsd GOARCH=amd64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/openbsd/$(PROJECT_NAME)-amd64
	@echo "✓ $(OUTPUT_DIR)/openbsd/$(PROJECT_NAME)-amd64"

build-openbsd-arm64:
	@echo "编译 OpenBSD arm64..."
	@mkdir -p $(OUTPUT_DIR)/openbsd
	@GOOS=openbsd GOARCH=arm64 go build -ldflags="$(LDFLAGS)" $(BUILD_FLAGS) -o $(OUTPUT_DIR)/openbsd/$(PROJECT_NAME)-arm64
	@echo "✓ $(OUTPUT_DIR)/openbsd/$(PROJECT_NAME)-arm64"

clean:
	@echo "清理编译输出..."
	@rm -rf $(OUTPUT_DIR)
	@echo "✓ 清理完成"

test:
	@echo "运行测试..."
	@go test -v ./...