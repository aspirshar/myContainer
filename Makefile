# MyContainer Makefile
# 确保启用CGO以支持C代码编译

# 变量定义
BINARY_NAME=myContainer
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')

# Go 编译参数
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CGO_ENABLED=1

# 编译标志
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GoVersion=${GO_VERSION}"

# 默认目标
.PHONY: all
all: clean build

# 编译项目
.PHONY: build
build:
	@echo "Building ${BINARY_NAME} with CGO_ENABLED=1..."
	@echo "Version: ${VERSION}"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Go Version: ${GO_VERSION}"
	CGO_ENABLED=1 go build ${LDFLAGS} -o ${BINARY_NAME} .

# 编译并安装到系统路径
.PHONY: install
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	@sudo cp ${BINARY_NAME} /usr/local/bin/
	@echo "Installation completed!"

# 编译到指定目录
.PHONY: build-dir
build-dir:
	@echo "Building ${BINARY_NAME} to ${BUILD_DIR}..."
	@mkdir -p ${BUILD_DIR}
	CGO_ENABLED=1 go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} .

# 交叉编译（Linux amd64）
.PHONY: build-linux
build-linux:
	@echo "Cross-compiling for Linux amd64..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-amd64 .

# 交叉编译（Linux arm64）
.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "Cross-compiling for Linux arm64..."
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o ${BINARY_NAME}-linux-arm64 .

# 交叉编译（所有平台）
.PHONY: build-all
build-all: build-linux build-linux-arm64
	@echo "Cross-compilation completed!"

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	CGO_ENABLED=1 go test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	CGO_ENABLED=1 go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

# 代码检查（使用go vet）
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# 清理编译产物
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f ${BINARY_NAME}
	@rm -f ${BINARY_NAME}-linux-*
	@rm -rf ${BUILD_DIR}
	@rm -f coverage.out coverage.html
	@echo "Clean completed!"

# 深度清理（包括依赖）
.PHONY: clean-all
clean-all: clean
	@echo "Cleaning all dependencies..."
	go clean -modcache
	@echo "Deep clean completed!"

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 更新依赖
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# 验证依赖
.PHONY: deps-verify
deps-verify:
	@echo "Verifying dependencies..."
	go mod verify

# 运行容器示例
.PHONY: demo
demo: build
	@echo "Running demo container..."
	@echo "Make sure you have root privileges and busybox.tar in images/ directory"
	@echo "Example: sudo ./${BINARY_NAME} run -it busybox /bin/sh"

# 运行基本测试
.PHONY: demo-basic
demo-basic: build
	@echo "Running basic container test..."
	@echo "Starting test container..."
	@sudo ./${BINARY_NAME} run -d -name test-container busybox /bin/sh -c "while true; do sleep 1; done" || true
	@sleep 2
	@echo "Container list:"
	@sudo ./${BINARY_NAME} ps
	@echo "Executing command in container:"
	@sudo ./${BINARY_NAME} exec test-container /bin/sh -c "echo 'Hello from container!'" || true
	@echo "Stopping container:"
	@sudo ./${BINARY_NAME} stop test-container || true
	@echo "Removing container:"
	@sudo ./${BINARY_NAME} rm test-container || true

# 检查系统要求
.PHONY: check
check:
	@echo "Checking system requirements..."
	@echo "Go version: $(shell go version)"
	@echo "CGO enabled: $(shell go env CGO_ENABLED)"
	@echo "GCC available: $(shell which gcc 2>/dev/null || echo 'GCC not found')"
	@echo "Root privileges: $(shell id -u 2>/dev/null | grep -q '^0$$' && echo 'Yes' || echo 'No - need sudo')"
	@echo "iptables available: $(shell which iptables 2>/dev/null || echo 'iptables not found')"
	@echo "bridge-utils available: $(shell which brctl 2>/dev/null || echo 'bridge-utils not found')"

# 显示帮助信息
.PHONY: help
help:
	@echo "MyContainer Makefile"
	@echo "==================="
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary with CGO_ENABLED=1"
	@echo "  install        - Build and install to /usr/local/bin"
	@echo "  build-dir      - Build to build/ directory"
	@echo "  build-linux    - Cross-compile for Linux amd64"
	@echo "  build-all      - Cross-compile for all platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (if available)"
	@echo "  vet            - Run go vet"
	@echo "  clean          - Clean build artifacts"
	@echo "  clean-all      - Deep clean including dependencies"
	@echo "  deps           - Download dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  deps-verify    - Verify dependencies"
	@echo "  demo           - Show demo instructions"
	@echo "  demo-basic     - Run basic container test"
	@echo "  check          - Check system requirements"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS           - Target OS (default: current OS)"
	@echo "  GOARCH         - Target architecture (default: current arch)"
	@echo "  CGO_ENABLED    - Enable CGO (default: 1)"
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build for current platform"
	@echo "  make build-linux              # Cross-compile for Linux"
	@echo "  make install                  # Install to system"
	@echo "  make demo-basic               # Run basic test"

# 默认目标
.DEFAULT_GOAL := help 