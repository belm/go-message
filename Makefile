# Makefile for go-message project

# 项目名称
APP_NAME := go-message

# 版本信息（可以从 git tag 获取）
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 构建目录
BUILD_DIR := build

# Go 相关变量
GO := go
GOFMT := gofmt
GOFLAGS := -ldflags "-X main.Version=$(VERSION)"

# 支持的平台列表
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64 \
	darwin/amd64 \
	darwin/arm64

.PHONY: all
all: clean fmt vet build ## 默认目标：格式化、检查并构建当前平台

.PHONY: build
build: ## 本地构建（当前平台）
	@echo "构建 $(APP_NAME) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)$(shell go env GOEXE) .

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOFMT) -s -w .

.PHONY: vet
vet: ## 代码检查
	@echo "代码检查..."
	$(GO) vet ./...

.PHONY: test
test: ## 运行测试
	@echo "运行测试..."
	$(GO) test -v ./...

.PHONY: build-all
build-all: clean fmt vet ## 构建所有平台的二进制文件
	@echo "构建所有平台的二进制文件..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT=$(BUILD_DIR)/$(APP_NAME)-$$OS-$$ARCH; \
		if [ "$$OS" = "windows" ]; then \
			OUTPUT=$$OUTPUT.exe; \
		fi; \
		echo "构建 $$OS/$$ARCH -> $$OUTPUT"; \
		GOOS=$$OS GOARCH=$$ARCH $(GO) build $(GOFLAGS) -o $$OUTPUT .; \
	done
	@echo "构建完成！二进制文件位于 $(BUILD_DIR)/ 目录"

.PHONY: build-linux
build-linux: ## 构建 Linux 平台 (amd64, arm64)
	@echo "构建 Linux 平台..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "构建 linux/$$arch..."; \
		GOOS=linux GOARCH=$$arch $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-$$arch .; \
	done

.PHONY: build-windows
build-windows: ## 构建 Windows 平台 (amd64, arm64)
	@echo "构建 Windows 平台..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "构建 windows/$$arch..."; \
		GOOS=windows GOARCH=$$arch $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-$$arch.exe .; \
	done

.PHONY: build-darwin
build-darwin: ## 构建 macOS 平台 (amd64, arm64)
	@echo "构建 macOS 平台..."
	@mkdir -p $(BUILD_DIR)
	@for arch in amd64 arm64; do \
		echo "构建 darwin/$$arch..."; \
		GOOS=darwin GOARCH=$$arch $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-$$arch .; \
	done

.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@echo "清理完成"

.PHONY: deps
deps: ## 下载并整理依赖
	@echo "下载依赖..."
	$(GO) mod download
	$(GO) mod tidy

.PHONY: run-producer
run-producer: build ## 运行生产者服务
	@echo "运行生产者服务..."
	./$(BUILD_DIR)/$(APP_NAME) -type=producer

.PHONY: run-consumer
run-consumer: build ## 运行消费者服务
	@echo "运行消费者服务..."
	./$(BUILD_DIR)/$(APP_NAME) -type=consumer

.PHONY: help
help: ## 显示帮助信息
	@echo "可用的 Make 目标:"
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(firstword $(MAKEFILE_LIST)) | awk 'BEGIN {FS = ":.*?## "} {printf "  %-20s %s\n", $$1, $$2}' | sort

