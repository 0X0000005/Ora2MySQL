@echo off
:: 设置控制台为 UTF-8 编码，防止中文输出乱码
chcp 65001 >nul
REM Oracle 转 MySQL DDL 工具 - Windows 编译脚本
REM 支持编译 Windows 和 Linux 版本，并使用 UPX 压缩

echo ========================================
echo Oracle 转 MySQL DDL 工具 - 编译脚本
echo ========================================
echo.

REM 设置 Go 环境变量
set CGO_ENABLED=0

echo [1/4] 编译 Windows 版本...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o o2m.exe ./cmd/o2m
if %errorlevel% neq 0 (
    echo 错误：Windows 版本编译失败
    exit /b 1
)
echo ✓ Windows 版本编译成功: o2m.exe

echo.
echo [2/4] 编译 Linux 版本...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o o2m-linux ./cmd/o2m
if %errorlevel% neq 0 (
    echo 错误：Linux 版本编译失败
    exit /b 1
)
echo ✓ Linux 版本编译成功: o2m-linux

echo.
echo [3/4] 使用 UPX 压缩二进制文件...
where upx >nul 2>nul
if %errorlevel% neq 0 (
    echo ⚠ 警告：未找到 UPX，跳过压缩步骤
    echo 提示：可从 https://upx.github.io/ 下载 UPX
    goto :skip_upx
)

echo 压缩 Windows 版本...
upx -9 o2m.exe
echo ✓ Windows 版本压缩完成

echo 压缩 Linux 版本...
upx -9 o2m-linux
echo ✓ Linux 版本压缩完成

:skip_upx

echo.
echo [4/4] 编译完成！
echo.
echo 生成的文件：
if exist o2m.exe (
    for %%F in (o2m.exe) do echo   - o2m.exe         (%%~zF bytes^)
)
if exist o2m-linux (
    for %%F in (o2m-linux) do echo   - o2m-linux      (%%~zF bytes^)
)

echo.
echo 使用方法：
echo   [命令模式]: o2m.exe -i input.sql -o output.sql
echo   [Web 模式]:  o2m.exe -web -port 8080
echo   [帮助信息]:  o2m.exe -h
echo.
echo ========================================
