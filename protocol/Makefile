
# protocol/game_third
GAME_THIRD_PROTO_DIR := $(PROTOCOL_DIR)/game_third
GAME_THIRD_PKG_DIR := $(PKG_DIR)/game_third
GAME_THIRD_PROTO_FILES := $(shell find $(GAME_THIRD_PROTO_DIR) -type f -name '*.proto')
GAME_THIRD_PB_FILES := $(patsubst $(GAME_THIRD_PROTO_DIR)/%.proto, $(GAME_THIRD_PKG_DIR)/%.pb.go, $(GAME_THIRD_PROTO_FILES))

$(GAME_THIRD_PB_FILES): $(GAME_THIRD_PROTO_FILES)
	mkdir -p $(GAME_THIRD_PKG_DIR) && \
	$(PROTOC) $(PROTOC_ARGS) $(PROTOC_PATH) $(GAME_THIRD_PROTO_FILES)

# protocol/share
PROTOC_PATH_SHARE := --proto_path=$(PROTOCOL_DIR)/share

# protocol/share/oss
OSS_PROTO_DIR := $(PROTOCOL_DIR)/share/oss
OSS_PKG_DIR := $(PKG_DIR)/oss
OSS_PROTO_FILES := $(shell find $(OSS_PROTO_DIR) -type f -name '*.proto')
OSS_PB_FILES := $(patsubst $(OSS_PROTO_DIR)/%.proto, $(OSS_PKG_DIR)/%.pb.go, $(OSS_PROTO_FILES))

$(OSS_PB_FILES): $(OSS_PROTO_FILES)
	mkdir -p $(OSS_PKG_DIR) && \
	$(PROTOC) $(PROTOC_ARGS) $(PROTOC_PATH_SHARE) $(OSS_PROTO_FILES)

# protocol/share/database
DATABASE_PROTO_DIR := $(PROTOCOL_DIR)/share/database
DATABASE_PKG_DIR := $(PKG_DIR)/database
DATABASE_PROTO_FILES := $(shell find $(DATABASE_PROTO_DIR) -type f -name '*.proto')
DATABASE_PB_FILES := $(patsubst $(DATABASE_PROTO_DIR)/%.proto, $(DATABASE_PKG_DIR)/%.pb.go, $(DATABASE_PROTO_FILES))

#! protocol/share/database 没有外部依赖, 因此只需要 $(PROTOC_PATH_SHARE)
$(DATABASE_PB_FILES): $(DATABASE_PROTO_FILES)
	mkdir -p $(DATABASE_PKG_DIR) && \
	$(PROTOC) $(PROTOC_ARGS) $(PROTOC_PATH_SHARE) $(DATABASE_PROTO_FILES)

# protocol/share/cs
CS_PROTO_DIR := $(PROTOCOL_DIR)/share/cs
CS_PKG_DIR := $(PKG_DIR)/cs
CS_PROTO_FILES := $(shell find $(CS_PROTO_DIR) -type f -name '*.proto')
CS_PB_FILES := $(patsubst $(CS_PROTO_DIR)/%.proto, $(CS_PKG_DIR)/%.pb.go, $(CS_PROTO_FILES))

#! protocol/share/cs 没有外部依赖, 因此只需要 $(PROTOC_PATH_SHARE)
$(CS_PB_FILES): $(CS_PROTO_FILES) $(DATABASE_PB_FILES)
	mkdir -p $(CS_PKG_DIR) && \
	$(PROTOC) $(PROTOC_ARGS) $(PROTOC_PATH_SHARE) $(CS_PROTO_FILES)

# protocol/ss
SS_PROTO_DIR := $(PROTOCOL_DIR)/ss
SS_PKG_DIR := $(PKG_DIR)/ss
SS_PROTO_FILES := $(shell find $(SS_PROTO_DIR) -type f -name '*.proto')
SS_PB_FILES := $(patsubst $(SS_PROTO_DIR)/%.proto, $(SS_PKG_DIR)/%.pb.go, $(SS_PROTO_FILES))

#! 注意: 此处的 $(PROTOC) 额外添加了 $(PROTOC_PATH_SHARE), 其余 target 无需如此,
#! 因为 SS_PROTO_FILES 中有 `import "database/ac_*.proto";`, 而 database 位于 protocol/share/database
$(SS_PB_FILES): $(SS_PROTO_FILES)
	mkdir -p $(SS_PKG_DIR) && \
	$(PROTOC) $(PROTOC_ARGS) $(PROTOC_PATH) $(PROTOC_PATH_SHARE) $(SS_PROTO_FILES)

# protocol/opt
PROTOCOL_OPT_DIR := $(GIT_ROOT)/protocol/opt
PROTOCOL_OPT_TEMPDIR := $(GIT_ROOT)/protocol/opt_tempdir
PROTOC_ARGS_OPT := --go_out=paths=source_relative:$(PKG_DIR)/opt
PROTOC_PATH_OPT := --proto_path=$(PROTOCOL_OPT_TEMPDIR)

# protocol/opt/scommon
OPTSCOMMON_PROTO_DIR := $(PROTOCOL_OPT_DIR)/scommon
OPTSCOMMON_PROTO_TEMPDIR := $(PROTOCOL_OPT_TEMPDIR)/scommon
OPTSCOMMON_PKG_DIR := $(PKG_DIR)/opt/scommon
OPTSCOMMON_PROTO_FILES := $(shell find $(OPTSCOMMON_PROTO_DIR) -type f -name '*.proto')
OPTSCOMMON_PB_FILES := $(patsubst $(OPTSCOMMON_PROTO_DIR)/%.proto, $(OPTSCOMMON_PKG_DIR)/%.pb.go, $(OPTSCOMMON_PROTO_FILES))
OPTSCOMMON_GO_PACKAGE := gamekit/pkg/opt/scommon;optscommon

#! 注意: 此处通过 opt_tempdir 中转并替换 go_package 占位符 __GO_OPTS__, 不要直接改 opt/scommon 下的 proto
$(OPTSCOMMON_PB_FILES): $(OPTSCOMMON_PROTO_FILES)
	mkdir -p $(OPTSCOMMON_PKG_DIR) $(OPTSCOMMON_PROTO_TEMPDIR) && \
		rsync -a $(OPTSCOMMON_PROTO_DIR)/ $(OPTSCOMMON_PROTO_TEMPDIR) && \
		sed -E -i '/^option go_package/s#__GO_OPTS__#'"$(OPTSCOMMON_GO_PACKAGE)"'#' \
			$$(find $(OPTSCOMMON_PROTO_TEMPDIR) -type f -name '*.proto') && \
		$(PROTOC) $(PROTOC_ARGS_OPT) $(PROTOC_PATH_OPT) \
			$$(find $(OPTSCOMMON_PROTO_TEMPDIR) -type f -name '*.proto')

.PHONY: proto
## proto: .proto 转 .pb.go
proto: $(GAME_THIRD_PB_FILES) $(OSS_PB_FILES) $(CS_PB_FILES) $(SS_PB_FILES) $(OPTSCOMMON_PB_FILES)
