@echo off
setlocal enabledelayedexpansion

REM 设置版本信息
set VERSION=1.0.0
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "BUILD_TIME=%dt:~0,4%-%dt:~4,2%-%dt:~6,2% %dt:~8,2%:%dt:~10,2%:%dt:~12,2%"

echo ======== DATAMGR-CLI 构建脚本 ========
echo 版本: %VERSION%
echo 构建时间: %BUILD_TIME%

REM 创建输出目录
if not exist build mkdir build

REM 清理旧构建
echo [BUILD] 清理旧构建文件...
del /q build\* 2>nul

REM 设置 LDFLAGS，包含版本信息
set LDFLAGS=-X github.com/yuanpli/datamgr-cli/cmd.Version=%VERSION% -X "github.com/yuanpli/datamgr-cli/cmd.BuildTime=%BUILD_TIME%"

echo.
echo ======== 构建 Windows 版本 ========
echo.

REM 构建 Windows(amd64) 版本
echo [BUILD] 正在构建 Windows(amd64) 版本...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -ldflags "%LDFLAGS%" -o build\datamgr-cli-windows-amd64.exe

if %ERRORLEVEL% neq 0 (
    echo [ERROR] Windows(amd64) 版本构建失败!
    exit /b 1
) else (
    echo [BUILD] Windows(amd64) 版本构建成功: build\datamgr-cli-windows-amd64.exe
    copy build\datamgr-cli-windows-amd64.exe build\datamgr-cli.exe >nul
)

REM 构建 Windows(386) 版本
echo [BUILD] 正在构建 Windows(386) 版本...
set GOOS=windows
set GOARCH=386
set CGO_ENABLED=0
go build -ldflags "%LDFLAGS%" -o build\datamgr-cli-windows-386.exe

if %ERRORLEVEL% neq 0 (
    echo [ERROR] Windows(386) 版本构建失败!
    exit /b 1
) else (
    echo [BUILD] Windows(386) 版本构建成功: build\datamgr-cli-windows-386.exe
)

echo.
echo ======== 生成SHA256校验和 ========
echo.

REM 生成SHA256校验和
echo [BUILD] 正在生成SHA256校验和...
cd build
for %%f in (*) do (
    if not "%%~xf"==".sha256" (
        certutil -hashfile "%%f" SHA256 | findstr /v "SHA256" | findstr /v "CertUtil" > "%%f.sha256"
        echo [BUILD] 生成校验和: %%f.sha256
    )
)
cd ..

echo.
echo ======== 构建完成 ========
echo.
echo [BUILD] Windows平台构建已完成。可执行文件位于 build\ 目录中
echo [BUILD] 构建产物:
dir /b build\

REM 询问是否压缩
echo.
set /p compress=是否要压缩二进制文件以减小体积? (y/n): 
if /i "%compress%"=="y" (
    echo.
    echo ======== 压缩二进制文件 ========
    echo.
    
    where upx >nul 2>&1
    if %ERRORLEVEL% neq 0 (
        echo [BUILD] 未找到UPX工具，跳过压缩步骤
        echo [BUILD] 如需压缩二进制文件，请安装UPX工具: https://github.com/upx/upx/releases
    ) else (
        echo [BUILD] 正在压缩二进制文件...
        for %%f in (build\*.exe) do (
            echo [BUILD] 压缩: %%f
            upx -9 "%%f"
        )
        echo [BUILD] 压缩完成
    )
)

echo.
echo ======== 全部完成 ========
echo.
echo [BUILD] datamgr-cli v%VERSION% 构建工作已完成

endlocal 