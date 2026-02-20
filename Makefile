APP := syl-md2ppt
GO ?= go
BIN_DIR ?= bin
BIN := $(BIN_DIR)/$(APP)
DESTDIR ?=
GO_BIN_DIR ?= $(shell sh -c 'gobin="$$( $(GO) env GOBIN )"; if [ -n "$$gobin" ]; then printf "%s" "$$gobin"; else gopath="$$( $(GO) env GOPATH )"; printf "%s/bin" "$${gopath%%:*}"; fi')
INSTALL_BIN_DIR := $(DESTDIR)$(GO_BIN_DIR)
INSTALL_BIN := $(INSTALL_BIN_DIR)/$(APP)
DEFAULT_GOAL := default

SOURCE ?=
OUTPUT ?=
CONFIG ?=

.DEFAULT_GOAL := $(DEFAULT_GOAL)

.PHONY: default help build test fmt tidy clean run run-build install uninstall

default:
	@$(MAKE) fmt
	@$(MAKE) test
	@$(MAKE) install

help:
	@echo "Targets:"
	@echo "  make              - 默认流程：fmt -> test -> install"
	@echo "  make build        - 编译二进制到 $(BIN)"
	@echo "  make test         - 运行全部测试"
	@echo "  make fmt          - gofmt 全部 Go 文件"
	@echo "  make tidy         - 整理 go.mod/go.sum"
	@echo "  make run          - 直跑入口（需要 SOURCE）"
	@echo "  make run-build    - build 子命令入口（需要 SOURCE）"
	@echo "  make install      - 安装到 Go bin 目录（GOBIN 或 GOPATH/bin）"
	@echo "  make uninstall    - 卸载已安装二进制"
	@echo "  make clean        - 删除构建产物"
	@echo ""
	@echo "Variables:"
	@echo "  SOURCE=/path/to/SPI      必填（run/run-build）"
	@echo "  OUTPUT=/path/out.pptx    可选"
	@echo "  CONFIG=/path/conf.yaml   可选"
	@echo "  GO_BIN_DIR=...           覆盖安装目录（默认 GOBIN 或 GOPATH/bin）"
	@echo "  DESTDIR=                 打包场景根目录"

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN) .

test:
	$(GO) test ./...

fmt:
	@gofmt -w $$(find . -name '*.go' -type f)

tidy:
	$(GO) mod tidy

run:
	@if [ -z "$(SOURCE)" ]; then echo "还没传 SOURCE 数据源目录"; exit 1; fi
	$(GO) run . "$(SOURCE)" $(if $(OUTPUT),--output "$(OUTPUT)",) $(if $(CONFIG),--config "$(CONFIG)",)

run-build:
	@if [ -z "$(SOURCE)" ]; then echo "还没传 SOURCE 数据源目录"; exit 1; fi
	$(GO) run . build "$(SOURCE)" $(if $(OUTPUT),--output "$(OUTPUT)",) $(if $(CONFIG),--config "$(CONFIG)",)

clean:
	rm -rf $(BIN_DIR)

install: build
	@mkdir -p "$(INSTALL_BIN_DIR)"
	install -m 0755 "$(BIN)" "$(INSTALL_BIN)"

uninstall:
	rm -f "$(INSTALL_BIN)"
