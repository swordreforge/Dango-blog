#!/bin/bash
# Windows 编译脚本 - 移除符号表 + UPX 最佳压缩

set -e

PROJECT_NAME="myblog-gogogo"
OUTPUT_DIR="build"
VERSION=$(date +%Y%m%d)

echo "========================================="
echo "Windows 编译配置"
echo "========================================="
echo "目标平台: Windows (amd64)"
echo "移除符号表: 是"
echo "UPX 压缩: 最佳压缩"
echo "========================================="

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 检查 UPX 是否安装
if ! command -v upx &> /dev/null; then
    echo "错误: UPX 未安装"
    echo "请先安装 UPX:"
    echo "  Arch Linux: sudo pacman -S upx"
    echo "  Ubuntu/Debian: sudo apt-get install upx"
    echo "  macOS: brew install upx"
    exit 1
fi

# 编译参数说明:
# -ldflags="-s -w"  -s 移除符号表, -w 移除 DWARF 调试信息
# --trimpath       移除文件系统路径
# -tags=netgo      使用纯 Go 网络栈,去除 CGO 依赖

echo ""
echo "正在编译 Windows 版本..."

GOOS=windows GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=$VERSION" \
    --trimpath \
    -tags=netgo \
    -o "$OUTPUT_DIR/${PROJECT_NAME}.exe"

echo "✓ 编译完成: $OUTPUT_DIR/${PROJECT_NAME}.exe"

# 获取原始文件大小
ORIGINAL_SIZE=$(du -h "$OUTPUT_DIR/${PROJECT_NAME}.exe" | cut -f1)
echo "原始大小: $ORIGINAL_SIZE"

# UPX 最佳压缩
echo ""
echo "正在使用 UPX 最佳压缩..."
upx --best --lzma "$OUTPUT_DIR/${PROJECT_NAME}.exe"

# 获取压缩后文件大小
COMPRESSED_SIZE=$(du -h "$OUTPUT_DIR/${PROJECT_NAME}.exe" | cut -f1)
echo "压缩后大小: $COMPRESSED_SIZE"

echo ""
echo "========================================="
echo "编译完成!"
echo "输出文件: $OUTPUT_DIR/${PROJECT_NAME}.exe"
echo "原始大小: $ORIGINAL_SIZE"
echo "压缩后: $COMPRESSED_SIZE"
echo "========================================="