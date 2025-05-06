FROM golang:1.23-alpine as builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git make bash upx

# 复制源代码
COPY . .

# 设置版本信息
ARG VERSION=1.0.0
ARG BUILD_TIME

# 构建
RUN if [ -z "$BUILD_TIME" ]; then \
        BUILD_TIME=$(date "+%Y-%m-%d %H:%M:%S"); \
    fi && \
    LDFLAGS="-X github.com/yuanpli/datamgr-cli/cmd.Version=$VERSION -X github.com/yuanpli/datamgr-cli/cmd.BuildTime=\"$BUILD_TIME\"" && \
    echo "Building version $VERSION ($BUILD_TIME)" && \
    go build -ldflags "$LDFLAGS" -o /app/datamgr-cli && \
    upx -9 /app/datamgr-cli

# 多阶段构建，使用更小的基础镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制可执行文件
COPY --from=builder /app/datamgr-cli /usr/local/bin/datamgr-cli

# 设置工作目录
WORKDIR /data

# 创建用于保存配置和数据的卷
VOLUME ["/data"]

# 设置入口点
ENTRYPOINT ["datamgr-cli"]
CMD ["help"] 