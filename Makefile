# str: tarball name prefix
GameId := gs

# str: auto generated build date
BUILD_DATE := $(shell date "+%F-%H%M%S")

# str: current git repo root_dir
GIT_ROOT := $(shell git rev-parse --show-toplevel)
# 原来 GIT_BRANCH 和 GIT_COMMIT 都是 Jenkins 的 builtin variables
# The remote branch name, if any.
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
# The commit hash being checked out.
GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD)
EXCEL_DIR_COMMIT ?= $(shell git rev-parse HEAD:./protocol/configs/resourse)
EXCEL_DIR_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD:./protocol/configs/resourse)

# path: shared directories
PKG_DIR := $(GIT_ROOT)/pkg
RUN_DIR := $(GIT_ROOT)/run
SRC_DIR := $(GIT_ROOT)/src
TOOLS_DIR := $(GIT_ROOT)/tools
PROTOCOL_DIR := $(GIT_ROOT)/protocol
CONFIGS_DIR := $(GIT_ROOT)/protocol/configs
DIST_DIR := $(GIT_ROOT)/dist

# Ansible configuration
ANSIBLE_DIR := $(GIT_ROOT)/run/ansible
ANSIBLE_HOME := $(HOME)/.ansible
ANSIBLE_BINDIR := $(HOME)/.local/bin
ANSIBLE_PLAYBOOK := $(ANSIBLE_BINDIR)/ansible-playbook

# path: run/config/ansible_*.yaml
CONFIG ?= $(ANSIBLE_DIR)/config/ansible-internal_test/ansible-internal_test-NA.yaml

# str: {配置表dev, 配置表test, 配置表stable, 配置表master}/{NA, EU}
EXCEL_DIR ?= 配置表dev/NA
# str: 将 / 替换为 - 用于 tarball 文件名
EXCEL_DIR_DASHED := $(subst /,-,$(EXCEL_DIR))
# 提取环境部分 (配置表dev)
EXCEL_DIR_ENV := $(firstword $(subst /, ,$(EXCEL_DIR)))
# 提取地区部分 (NA)
EXCEL_DIR_REGION := $(lastword $(subst /, ,$(EXCEL_DIR)))

# protobuf args
PROTOC ?= protoc
PROTOC_PATH := --proto_path=$(PROTOCOL_DIR)
PROTOC_ARGS := --go_out=paths=source_relative:$(PKG_DIR)

# go args
GO_EXEC ?= go
CGO_ENABLED ?= 0
GOOS ?= linux
GOARCH ?= amd64
# arguments to pass on each "go tool compile" invocation
# ref: https://pkg.go.dev/cmd/compile
# -N    disable optimizations
# -l    disable inlining
# GO_BUILD_GCFLAGS ?= -gcflags=""
GO_BUILD_GCFLAGS ?= -gcflags="all=-N -l"
# arguments to pass on each "go tool link" invocation
# ref: https://pkg.go.dev/cmd/link
# GO_BUILD_LDFLAGS ?= -ldflags=""
GO_BUILD_LDFLAGS ?=
# go build env
GO_BUILD_ENV := CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH)
# go build command
GO_BUILD := $(GO_BUILD_ENV) $(GO_EXEC) build $(GO_BUILD_GCFLAGS) $(GO_BUILD_LDFLAGS)

.DEFAULT_GOAL=help
.PHONY: help
## help: 显示这段帮助信息
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/    /'

# include Makefiles
include $(PROTOCOL_DIR)/Makefile
include $(TOOLS_DIR)/Makefile
include $(CONFIGS_DIR)/Makefile
include $(SRC_DIR)/Makefile
include $(RUN_DIR)/Makefile

.PHONY: clean
## clean: 清理各种中间产物 (.pb.go, blob; 保留脚手架已提交的示例文件)
clean:
	rm -rf $(DIST_DIR)
	rm -rf $(PROTOCOL_OPT_TEMPDIR)
	rm -f $(BUILD_EXEC)
	rm -f $(TOOLS_EXEC) $(TOOLS_EXTRA)
	rm -f $(RUN_DIR)/tools/db/env.sh
	rm -f $(SRC_DIR)/base/control/build.go
	find $(PKG_DIR) -type f -name '*.pb.go' -delete
	find $(CONFIGS_DIR) -type f -name '*.proto' \
		! -name 'ConfigInitArgsConfig.proto' ! -name 'WordsConfig.proto' -delete
	find $(RUN_DIR) -type f \
		-not -name 'patchAccIdArea.json' \
		-not -name 'ConfigInitArgsConfig.json' -not -name 'ConfigInitArgsConfig.bytes' \
		-not -name 'WordsConfig.json' -not -name 'WordsConfig.bytes' \
		-and -regextype awk -regex '.+/run/binconf/.+\.(json|ya?ml|bytes)$$' -delete
	find $(RUN_DIR) -type f \
		-not -name 'zodiac-heros-firebase.json' \
		-and -regextype awk -regex '.+/run/.+/conf/.+\.(json|ya?ml)$$' -delete
