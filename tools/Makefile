
# tools/excelToProtobuf
TOOLS_EXCEL := excelToProtobuf
TOOLS_EXCEL_DIR := $(TOOLS_DIR)/$(TOOLS_EXCEL)
TOOLS_EXCEL_SRC := $(shell find $(TOOLS_EXCEL_DIR) -type f -name '*.go') $(TOOLS_EXCEL_DIR)/go.mod $(TOOLS_EXCEL_DIR)/go.sum

$(TOOLS_EXCEL_DIR)/$(TOOLS_EXCEL): $(TOOLS_EXCEL_SRC)
	cd $(TOOLS_EXCEL_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(TOOLS_EXCEL) main.go

# tools/client
TOOLS_CLIENT := client
TOOLS_CLIENT_DIR := $(TOOLS_DIR)/$(TOOLS_CLIENT)
TOOLS_CLIENT_SRC := $(shell find $(TOOLS_CLIENT_DIR) -type f -name '*.go') $(TOOLS_CLIENT_DIR)/go.mod $(TOOLS_CLIENT_DIR)/go.sum

# import "gamekit/pkg/cs"
$(TOOLS_CLIENT_DIR)/$(TOOLS_CLIENT): $(TOOLS_CLIENT_SRC) $(CS_PB_FILES)
	cd $(TOOLS_CLIENT_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(TOOLS_CLIENT) scripts/main.go

# tools/misc
TOOLS_MISC := misc
TOOLS_MISC_DIR := $(TOOLS_DIR)/$(TOOLS_MISC)
TOOLS_MISC_SRC := $(shell find $(TOOLS_MISC_DIR) -type f -name '*.go') $(TOOLS_MISC_DIR)/go.mod $(TOOLS_MISC_DIR)/go.sum

# import "gamekit/pkg/ss", "gamekit/pkg/game_third"
$(TOOLS_MISC_DIR)/$(TOOLS_MISC): $(TOOLS_MISC_SRC) $(SS_PB_FILES) $(GAME_THIRD_PB_FILES)
	cd $(TOOLS_MISC_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(TOOLS_MISC) scripts/main.go

# tools/register
TOOLS_REGISTER := register
TOOLS_REGISTER_DIR := $(TOOLS_DIR)/$(TOOLS_REGISTER)
TOOLS_REGISTER_SRC := $(shell find $(TOOLS_REGISTER_DIR) -type f -name '*.go') $(TOOLS_REGISTER_DIR)/go.mod $(TOOLS_REGISTER_DIR)/go.sum

# import "gamekit/pkg/cs", "gamekit/pkg/ss"
$(TOOLS_REGISTER_DIR)/$(TOOLS_REGISTER): $(TOOLS_REGISTER_SRC) $(CS_PB_FILES) $(SS_PB_FILES)
	cd $(TOOLS_REGISTER_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(TOOLS_REGISTER) scripts/main.go
.PHONY: tools/register
## tools/register: 构建 tools/register 并生成必要的 .pb.go
tools/register: $(TOOLS_REGISTER_DIR)/$(TOOLS_REGISTER)
	cd $(TOOLS_REGISTER_DIR) && ./register

# tools/db
TOOLS_DB := db
TOOLS_DB_DIR := $(TOOLS_DIR)/$(TOOLS_DB)
TOOLS_DB_SRC := $(shell find $(TOOLS_DB_DIR) -type f -name '*.go') $(TOOLS_DB_DIR)/go.mod $(TOOLS_DB_DIR)/go.sum

# import "gamekit/pkg/database"
$(TOOLS_DB_DIR)/$(TOOLS_DB): $(TOOLS_DB_SRC) $(DATABASE_PB_FILES)
	cd $(TOOLS_DB_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(TOOLS_DB) scripts/main.go
.PHONY: tools/db
## tools/db: 构建 tools/db 并生成必要的 .pb.go
tools/db: $(TOOLS_DB_DIR)/$(TOOLS_DB)
	cd $(TOOLS_DB_DIR) && ./db

# tools/buildInfo
TOOLS_BUILDINFO := buildInfo
TOOLS_BUILDINFO_DIR := $(TOOLS_DIR)/$(TOOLS_BUILDINFO)
TOOLS_BUILDINFO_SRC := $(shell find $(TOOLS_BUILDINFO_DIR) -type f -name '*.go') $(TOOLS_BUILDINFO_DIR)/go.mod $(TOOLS_BUILDINFO_DIR)/go.sum

$(TOOLS_BUILDINFO_DIR)/$(TOOLS_BUILDINFO): $(TOOLS_BUILDINFO_SRC)
	cd $(TOOLS_BUILDINFO_DIR) && $(GO_EXEC) mod tidy && $(GO_BUILD) -o $(TOOLS_BUILDINFO) cmd/main.go
## buildInfo: 构建 tools/buildInfo 并生成 src/base/control/build.go
buildInfo: $(TOOLS_BUILDINFO_DIR)/$(TOOLS_BUILDINFO)
	$(TOOLS_BUILDINFO_DIR)/$(TOOLS_BUILDINFO) -git-dir $(GIT_ROOT) -output $(SRC_DIR)/base/control/build.go

TOOLS_EXEC := \
	$(TOOLS_EXCEL_DIR)/$(TOOLS_EXCEL) \
	$(TOOLS_CLIENT_DIR)/$(TOOLS_CLIENT) \
	$(TOOLS_MISC_DIR)/$(TOOLS_MISC)

TOOLS_EXTRA := \
	$(TOOLS_REGISTER_DIR)/$(TOOLS_REGISTER) \
	$(TOOLS_DB_DIR)/$(TOOLS_DB) \
	$(TOOLS_BUILDINFO_DIR)/$(TOOLS_BUILDINFO)

.PHONY: tools
## tools: 编译 tools/ 目录下的工具
tools: $(TOOLS_EXEC)
