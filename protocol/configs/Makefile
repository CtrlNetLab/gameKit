
# protocol/configs
excelPath := $(CONFIGS_DIR)/resourse/$(EXCEL_DIR)
excelPathLanguage := $(CONFIGS_DIR)/resourse/$(EXCEL_DIR_ENV)/language

# tools/excelToProtobuf
## build_excel: 使用 tools/excelToProtobuf 工具读取策划仓库 Excel 生成配置 .proto 和 .json
# excelToProtobuf 命令行选项各个目录中的 / 不可省略?
.PHONY: build_excel
build_excel: $(TOOLS_EXCEL_DIR)/$(TOOLS_EXCEL)
	@echo "策划表配置目录: $(EXCEL_DIR), 策划表提交: $(EXCEL_DIR_COMMIT)"
	$(TOOLS_EXCEL_DIR)/$(TOOLS_EXCEL) \
		--excelPath $(excelPath)/ \
		--dataPath $(RUN_DIR)/binconf/ \
		--protoPath $(PROTOCOL_DIR)/
	@echo "策划表配置目录: $(EXCEL_DIR), 策划表提交: $(EXCEL_DIR_COMMIT)"

.PHONY: build_language
build_language: $(TOOLS_EXCEL_DIR)/$(TOOLS_EXCEL)
	@echo "策划表配置目录: $(EXCEL_DIR), 策划表提交: $(EXCEL_DIR_COMMIT)"
	$(TOOLS_EXCEL_DIR)/$(TOOLS_EXCEL) \
		--excelPath $(excelPathLanguage)/ \
		--excelName "Language.xlsx" \
		--dataPath $(RUN_DIR)/binconf/ \
		--protoPath $(PROTOCOL_DIR)/
	@echo "策划表配置目录: $(EXCEL_DIR), 策划表提交: $(EXCEL_DIR_COMMIT)"

.PHONY: gen_primary
gen_primary:
	mkdir -p $(PKG_DIR)/configs && \
	PKG_DIR=$(PKG_DIR) CONFIGS_DIR=$(CONFIGS_DIR) $(CONFIGS_DIR)/gen-primary.sh && \
	cp $(CONFIGS_DIR)/ResLoadMgr.go.tmpl $(PKG_DIR)/configs/ResLoadMgr.go

.PHONY: configs
## configs: 将策划仓库 Excel 生成的 .proto 编译成 .pb.go
configs: build_excel build_language gen_primary
	$(PROTOC) $(PROTOC_ARGS) $(PROTOC_PATH) $(CONFIGS_DIR)/*.proto
