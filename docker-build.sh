#!/bin/bash

# 颜色代码
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}${BOLD}[DOCKER-BUILD]${NC} $1"
}

print_error() {
    echo -e "${RED}${BOLD}[ERROR]${NC} $1"
}

print_header() {
    echo -e "\n${BLUE}${BOLD}======== $1 ========${NC}\n"
}

# 设置版本信息
VERSION="1.0.0"
BUILD_TIME=$(date "+%Y-%m-%d %H:%M:%S")

print_header "DATAMGR-CLI Docker构建脚本"
echo "版本: $VERSION"
echo "构建时间: $BUILD_TIME"

# 创建输出目录
mkdir -p build

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    print_error "未找到Docker命令，请安装Docker"
    exit 1
fi

# 构建模式
if [ "$1" = "buildx" ]; then
    # 使用buildx进行多平台构建
    print_header "使用Docker Buildx进行多平台构建"
    
    # 检查buildx是否可用
    if ! docker buildx version &> /dev/null; then
        print_error "Docker Buildx不可用，请安装或启用Docker Buildx"
        exit 1
    fi
    
    # 准备buildx构建器
    print_message "准备多平台构建器..."
    docker buildx create --name datamgr-builder --use || docker buildx use datamgr-builder
    docker buildx inspect --bootstrap
    
    # 构建多平台镜像
    print_message "构建多平台Docker镜像..."
    docker buildx build --platform linux/amd64,linux/arm64,linux/386 \
        --build-arg VERSION=$VERSION \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        -t datamgr-cli:$VERSION \
        --load .
    
    if [ $? -ne 0 ]; then
        print_error "Docker镜像构建失败"
        exit 1
    fi
    
    print_message "多平台Docker镜像构建成功: datamgr-cli:$VERSION"
else
    # 常规构建模式
    print_header "使用Docker构建当前平台版本"
    
    print_message "构建Docker镜像..."
    docker build \
        --build-arg VERSION=$VERSION \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        -t datamgr-cli:$VERSION .
    
    if [ $? -ne 0 ]; then
        print_error "Docker镜像构建失败"
        exit 1
    fi
    
    print_message "Docker镜像构建成功: datamgr-cli:$VERSION"
fi

# 从Docker容器中提取可执行文件
print_header "从Docker容器中提取可执行文件"

print_message "创建临时容器..."
CONTAINER_ID=$(docker create datamgr-cli:$VERSION)

print_message "提取可执行文件..."
docker cp $CONTAINER_ID:/usr/local/bin/datamgr-cli ./build/datamgr-cli

print_message "删除临时容器..."
docker rm $CONTAINER_ID

print_message "生成SHA256校验和..."
cd build
if [[ "$OSTYPE" == "darwin"* ]]; then
    shasum -a 256 datamgr-cli > datamgr-cli.sha256
else
    sha256sum datamgr-cli > datamgr-cli.sha256
fi
cd ..

print_header "构建完成"
print_message "可执行文件已保存到: build/datamgr-cli"
print_message "Docker镜像: datamgr-cli:$VERSION"

# 显示使用说明
print_header "使用说明"
echo "1. 直接运行可执行文件:"
echo "   ./build/datamgr-cli"
echo ""
echo "2. 使用Docker运行:"
echo "   docker run --rm -it datamgr-cli:$VERSION"
echo ""
echo "3. 连接到数据库:"
echo "   docker run --rm -it datamgr-cli:$VERSION connect -H host -P port -u user -p password -D dbname"
echo "" 