# 多阶段构建 Dockerfile
# 第一阶段：构建阶段
FROM golang:1.21-alpine AS builder

# 配置 Alpine 国内镜像源（阿里云）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 配置 Go 代理（使用国内镜像加速）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 设置工作目录
WORKDIR /build

# 安装必要的构建工具（使用国内源）
RUN apk update && \
    apk add --no-cache git make

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖（使用国内代理）
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件（构建所有平台，但这里只构建当前平台）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o go-message .

# 第二阶段：运行阶段
FROM alpine:latest

# 配置 Alpine 国内镜像源（阿里云）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时依赖（使用国内源）
RUN apk update && \
    apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/go-message .

# 复制配置文件
COPY --from=builder /build/config ./config

# 更改文件所有者
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口（HTTP 服务端口）
EXPOSE 8080

# 设置默认命令（可以通过 docker run 覆盖）
CMD ["./go-message", "-type=producer"]

