@echo off
REM Windows 编译脚本 - 移除符号表 + UPX 最佳压缩
REM 需要先安装 UPX: https://upx.github.io/

setlocal enabledelayedexpansion

set PROJECT_NAME=myblog-gogogo
set OUTPUT_DIR=build
set VERSION=%date:~0,4%%date:~5,2%%date:~8,2%

echo =========================================
echo Windows 编译配置
echo =========================================
echo 目标平台: Windows ^(amd64^)
echo 移除符号表: 是
echo UPX 压缩: 最佳压缩
echo =========================================

REM 创建输出目录
if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM 检查 UPX 是否安装
where upx >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: UPX 未安装
    echo 请先安装 UPX: https://upx.github.io/
    exit /b 1
)

REM 编译参数说明:
REM -ldflags="-s -w"  -s 移除符号表, -w 移除 DWARF 调试信息
REM --trimpath       移除文件系统路径
REM -tags=netgo      使用纯 Go 网络栈,去除 CGO 依赖

echo.
echo 正在编译 Windows 版本...

set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

go build ^
    -ldflags="-s -w -X main.Version=%VERSION%" ^
    --trimpath ^
    -tags=netgo ^
    -o "%OUTPUT_DIR%\%PROJECT_NAME%.exe"

if %errorlevel% neq 0 (
    echo 编译失败!
    exit /b 1
)

echo ✓ 编译完成: %OUTPUT_DIR%\%PROJECT_NAME%.exe

REM UPX 最佳压缩
echo.
echo 正在使用 UPX 最佳压缩...
upx --best --lzma "%OUTPUT_DIR%\%PROJECT_NAME%.exe"

echo.
echo =========================================
echo 编译完成!
echo 输出文件: %OUTPUT_DIR%\%PROJECT_NAME%.exe
echo =========================================

endlocal