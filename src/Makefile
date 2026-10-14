
BUILD_REQUIREMENTS := proto configs tools/register tools/db

# src/benchmark
BENCHMARK := benchmark
BENCHMARK_SRC_DIR := $(SRC_DIR)/$(BENCHMARK)
BENCHMARK_BIN_DIR := $(DIST_DIR)/$(BENCHMARK)/bin
BENCHMARK_SRC_FILES := $(shell find $(BENCHMARK_SRC_DIR) -type f -name '*.go') $(BENCHMARK_SRC_DIR)/go.mod $(BENCHMARK_SRC_DIR)/go.sum

$(BENCHMARK_BIN_DIR)/$(BENCHMARK): $(BENCHMARK_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(BENCHMARK_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(BENCHMARK_BIN_DIR)/$(BENCHMARK) scripts/main.go
.PHONY: benchmark
## benchmark: 构建 src/benchmark
benchmark: $(BENCHMARK_BIN_DIR)/$(BENCHMARK)

# src/control
CONTROL := control
CONTROL_SRC_DIR := $(SRC_DIR)/$(CONTROL)
CONTROL_BIN_DIR := $(DIST_DIR)/$(CONTROL)/bin
CONTROL_SRC_FILES := $(shell find $(CONTROL_SRC_DIR) -type f -name '*.go') $(CONTROL_SRC_DIR)/go.mod $(CONTROL_SRC_DIR)/go.sum

$(CONTROL_BIN_DIR)/$(CONTROL): $(CONTROL_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(CONTROL_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(CONTROL_BIN_DIR)/$(CONTROL) scripts/main.go
.PHONY: control
## control: 构建 src/control
control: $(CONTROL_BIN_DIR)/$(CONTROL)

# src/gamesvr
GAMESVR := gamesvr
GAMESVR_SRC_DIR := $(SRC_DIR)/$(GAMESVR)
GAMESVR_BIN_DIR := $(DIST_DIR)/$(GAMESVR)/bin
GAMESVR_SRC_FILES := $(shell find $(GAMESVR_SRC_DIR) -type f -name '*.go') $(GAMESVR_SRC_DIR)/go.mod $(GAMESVR_SRC_DIR)/go.sum

$(GAMESVR_BIN_DIR)/$(GAMESVR): $(GAMESVR_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(GAMESVR_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(GAMESVR_BIN_DIR)/$(GAMESVR) scripts/main.go
.PHONY: gamesvr
## gamesvr: 构建 src/gamesvr
gamesvr: $(GAMESVR_BIN_DIR)/$(GAMESVR)

# src/globalsvr
GLOBALSVR := globalsvr
GLOBALSVR_SRC_DIR := $(SRC_DIR)/$(GLOBALSVR)
GLOBALSVR_BIN_DIR := $(DIST_DIR)/$(GLOBALSVR)/bin
GLOBALSVR_SRC_FILES := $(shell find $(GLOBALSVR_SRC_DIR) -type f -name '*.go') $(GLOBALSVR_SRC_DIR)/go.mod $(GLOBALSVR_SRC_DIR)/go.sum

$(GLOBALSVR_BIN_DIR)/$(GLOBALSVR): $(GLOBALSVR_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(GLOBALSVR_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(GLOBALSVR_BIN_DIR)/$(GLOBALSVR) scripts/main.go
.PHONY: globalsvr
## globalsvr: 构建 src/globalsvr
globalsvr: $(GLOBALSVR_BIN_DIR)/$(GLOBALSVR)

# src/matchsvr
MATCHSVR := matchsvr
MATCHSVR_SRC_DIR := $(SRC_DIR)/$(MATCHSVR)
MATCHSVR_BIN_DIR := $(DIST_DIR)/$(MATCHSVR)/bin
MATCHSVR_SRC_FILES := $(shell find $(MATCHSVR_SRC_DIR) -type f -name '*.go') $(MATCHSVR_SRC_DIR)/go.mod $(MATCHSVR_SRC_DIR)/go.sum

$(MATCHSVR_BIN_DIR)/$(MATCHSVR): $(MATCHSVR_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(MATCHSVR_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(MATCHSVR_BIN_DIR)/$(MATCHSVR) scripts/main.go
.PHONY: matchsvr
## matchsvr: 构建 src/matchsvr
matchsvr: $(MATCHSVR_BIN_DIR)/$(MATCHSVR)

# src/miscsvr
MISCSVR := miscsvr
MISCSVR_SRC_DIR := $(SRC_DIR)/$(MISCSVR)
MISCSVR_BIN_DIR := $(DIST_DIR)/$(MISCSVR)/bin
MISCSVR_SRC_FILES := $(shell find $(MISCSVR_SRC_DIR) -type f -name '*.go') $(MISCSVR_SRC_DIR)/go.mod $(MISCSVR_SRC_DIR)/go.sum

$(MISCSVR_BIN_DIR)/$(MISCSVR): $(MISCSVR_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(MISCSVR_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(MISCSVR_BIN_DIR)/$(MISCSVR) scripts/main.go
.PHONY: miscsvr
## miscsvr: 构建 src/miscsvr
miscsvr: $(MISCSVR_BIN_DIR)/$(MISCSVR)

# src/relaysvr
RELAYSVR := relaysvr
RELAYSVR_SRC_DIR := $(SRC_DIR)/$(RELAYSVR)
RELAYSVR_BIN_DIR := $(DIST_DIR)/$(RELAYSVR)/bin
RELAYSVR_SRC_FILES := $(shell find $(RELAYSVR_SRC_DIR) -type f -name '*.go') $(RELAYSVR_SRC_DIR)/go.mod $(RELAYSVR_SRC_DIR)/go.sum

$(RELAYSVR_BIN_DIR)/$(RELAYSVR): $(RELAYSVR_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(RELAYSVR_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(RELAYSVR_BIN_DIR)/$(RELAYSVR) scripts/main.go
.PHONY: relaysvr
## relaysvr: 构建 src/relaysvr
relaysvr: $(RELAYSVR_BIN_DIR)/$(RELAYSVR)

# src/roomsvr
ROOMSVR := roomsvr
ROOMSVR_SRC_DIR := $(SRC_DIR)/$(ROOMSVR)
ROOMSVR_BIN_DIR := $(DIST_DIR)/$(ROOMSVR)/bin
ROOMSVR_SRC_FILES := $(shell find $(ROOMSVR_SRC_DIR) -type f -name '*.go') $(ROOMSVR_SRC_DIR)/go.mod $(ROOMSVR_SRC_DIR)/go.sum

$(ROOMSVR_BIN_DIR)/$(ROOMSVR): $(ROOMSVR_SRC_FILES) $(BUILD_REQUIREMENTS)
	cd $(ROOMSVR_SRC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(ROOMSVR_BIN_DIR)/$(ROOMSVR) scripts/main.go
.PHONY: roomsvr
## roomsvr: 构建 src/roomsvr
roomsvr: $(ROOMSVR_BIN_DIR)/$(ROOMSVR)

# game service
BUILD_EXEC := \
	$(CONTROL_BIN_DIR)/$(CONTROL) \
	$(GAMESVR_BIN_DIR)/$(GAMESVR) \
	$(GLOBALSVR_BIN_DIR)/$(GLOBALSVR) \
	$(MATCHSVR_BIN_DIR)/$(MATCHSVR) \
	$(MISCSVR_BIN_DIR)/$(MISCSVR) \
	$(RELAYSVR_BIN_DIR)/$(RELAYSVR) \
	$(ROOMSVR_BIN_DIR)/$(ROOMSVR) \
	$(BENCHMARK_BIN_DIR)/$(BENCHMARK)

.PHONY: build
## build: 编译 src/ 目录下游戏相关模块
build: buildInfo $(BUILD_EXEC)
