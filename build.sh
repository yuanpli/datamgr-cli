#!/bin/bash

# 创建构建目录
mkdir -p build
rm -rf build/*

echo "====== 构建当前平台版本 ======"
go build -o build/datamgr-cli
if [ $? -eq 0 ]; then
    echo "当前平台版本构建成功: build/datamgr-cli"
else
    echo "当前平台版本构建失败!"
    exit 1
fi

echo "====== 构建 Linux 版本 ======"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/datamgr-cli-linux
if [ $? -eq 0 ]; then
    echo "Linux版本构建成功: build/datamgr-cli-linux"
else
    echo "Linux版本构建失败!"
    exit 1
fi

echo "====== 构建 Windows 版本 ======"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o build/datamgr-cli.exe
if [ $? -eq 0 ]; then
    echo "Windows版本构建成功: build/datamgr-cli.exe"
else
    echo "Windows版本构建失败!"
    exit 1
fi

echo "====== 构建完成 ======"
echo "构建结果:"
ls -lh build/ 