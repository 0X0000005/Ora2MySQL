#!/bin/bash
# Oracle 转 MySQL DDL 工具 - Linux/macOS 编译脚本
# 支持编译 Linux、Windows 和 macOS 版本，并使用 UPX 压缩

set -e

echo "========================================"
echo "Oracle 转 MySQL DDL 工具 - 编译脚本"
echo "========================================"
echo ""

# 设置 Go 环境变量
export CGO_ENABLED=0

echo "[1/5] 编译 Linux 版本..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o o2m-linux ./cmd/o2m
echo "✓ Linux 版本编译成功: o2m-linux"

echo ""
echo "[2/5] 编译 Windows 版本..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o o2m.exe ./cmd/o2m
echo "✓ Windows 版本编译成功: o2m.exe"

echo ""
echo "[3/5] 编译 macOS 版本..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o o2m-macos ./cmd/o2m
echo "✓ macOS 版本编译成功: o2m-macos"

echo ""
echo "[4/5] 使用 UPX 压缩二进制文件..."
if ! command -v upx &> /dev/null; then
    echo "⚠ 警告：未找到 UPX，跳过压缩步骤"
    echo "提示：可使用以下命令安装 UPX："
    echo "  Ubuntu/Debian: sudo apt-get install upx"
    echo "  macOS:         brew install upx"
else
    echo "压缩 Linux 版本..."
    upx -9 o2m-linux
    echo "✓ Linux 版本压缩完成"
    
    echo "压缩 Windows 版本..."
    upx -9 o2m.exe
    echo "✓ Windows 版本压缩完成"
    
    echo "压缩 macOS 版本..."
    upx -9 o2m-macos
    echo "✓ macOS 版本压缩完成"
fi

echo ""
echo "[5/5] 设置可执行权限..."
chmod +x o2m-linux o2m-macos

echo ""
echo "编译完成！"
echo ""
echo "生成的文件："
if [ -f o2m-linux ]; then
    ls -lh o2m-linux | awk '{print "  - o2m-linux     ("$5")"}'
fi
if [ -f o2m.exe ]; then
    ls -lh o2m.exe | awk '{print "  - o2m.exe       ("$5")"}'
fi
if [ -f o2m-macos ]; then
    ls -lh o2m-macos | awk '{print "  - o2m-macos     ("$5")"}'
fi

echo ""
echo "使用方法："
echo "  命令行模式： ./o2m-linux -i input.sql -o output.sql"
echo "  Web 模式：   ./o2m-linux -web -port 8080"
echo "  帮助信息：   ./o2m-linux -h"
echo ""
echo "========================================"
