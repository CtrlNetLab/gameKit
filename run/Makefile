# tarball 文件名
DIST_FILE_DIR ?= /tmp
# 允许外部覆盖自动生成的文件名, 为满足部署已经验证过的压缩包的需求
# 外部传入 DIST_FILE 覆盖时, 此时不能走 make upload_dist, 应该直接去 make deploy_dist, 否则已经存在的包会被新构建的包覆盖
DIST_FILE ?= $(GameId)-$(GIT_BRANCH)@$(GIT_COMMIT_SHORT)-$(EXCEL_DIR_DASHED)@$(EXCEL_DIR_COMMIT_SHORT)-$(BUILD_DATE).tar.gz
DIST_FILE_SHA256 := $(DIST_FILE).sha256
DIST_FILE_LATEST := $(GameId)-$(GIT_BRANCH)-$(EXCEL_DIR_DASHED)-latest.txt
# 待上传文件列表, 空格分隔, 内含指向最新压缩包的 DIST_FILE_LATEST
# DIST_FILES 与 ansible 中的 dist_files / dist_file 都包含 path, _basename 不包含 path
DIST_FILES := \
	$(DIST_FILE_DIR)/$(DIST_FILE) \
	$(DIST_FILE_DIR)/$(DIST_FILE_SHA256) \
	$(DIST_FILE_DIR)/$(DIST_FILE_LATEST)

# tarball 中需要包含的文件, 压缩包有新增文件需求时请修改这里
# 已替换为 rsync 的 new-style filter rules, - 为 exclude, + 为 include,
# 更高级的 FILTER RULES 没看明白, 可以在 man rsync 中搜索 "FILTER RULES IN DEPTH"
RSYNC_ARGS := \
	--archive \
	--verbose \
	--prune-empty-dirs \
	--filter='- /ansible/' \
	--filter='- /benchmark/' \
	--filter='- /config/' \
	--filter='- /*svr_[0-9]/' \
	--filter='+ */' \
	--filter='+ *.bytes' \
	--filter='+ *.json' \
	--filter='+ *.sql' \
	--filter='+ *.lua' \
	--filter='+ tools/db/*.sh' \
	--filter='+ control.sh' \
	--filter='- *'

.PHONY: tarball
## tarball: 将构建好的可执行程序与必要文件打成 .tar.gz 文件
tarball: build
	rsync $(RSYNC_ARGS) $(RUN_DIR)/ $(DIST_DIR)/ && { \
		cd $(DIST_DIR) && \
		tar -czf $(DIST_FILE_DIR)/$(DIST_FILE) . && \
		cd $(DIST_FILE_DIR) && \
		sha256sum $(DIST_FILE) >$(DIST_FILE_DIR)/$(DIST_FILE_SHA256) && \
		echo "$(DIST_FILE)" >$(DIST_FILE_DIR)/$(DIST_FILE_LATEST) ; \
	}

# str: 逗号分隔的字符串
# 可用 ANSIBLE_TAGS: init, upload, deploy, config
ANSIBLE_TAGS ?= init

# path: run/config/ansible_*.hosts.yaml
# 在 basename 和 suffix 中间插入 .hosts 字符串
ANSIBLE_INVENTORY := $(basename $(CONFIG)).hosts$(suffix $(CONFIG))

# ANSIBLE_INVENTORY 已对所有主机进行分组
# 可以传入 group name 如 gamesvr
playbook_hosts ?= all

# roles/dist_systemd
# str: systemd service 服务状态
# started / stopped / restarted / reloaded
systemd_service_state ?= restarted
# bool: true / false
systemd_service_enabled ?= true

# run/ansible/playbooks/dist_deploy.yaml
# str: 逗号分隔的待部署服务列表
DIST_SERVICES ?= gamesvr,relaysvr,miscsvr,roomsvr,roomsvr_global,globalsvr,matchsvr

# common ansible-playbook args
ANSIBLE_PLAYBOOK_ARGS := \
	--verbose \
	--tags='$(ANSIBLE_TAGS)' \
	--inventory='$(ANSIBLE_INVENTORY)' \
	--extra-vars=@$(CONFIG) \
	--extra-vars=@$(ANSIBLE_HOME)/vault/gamekit.yaml

ANSIBLE_DEPLOY_ARGS := \
	$(ANSIBLE_PLAYBOOK_ARGS) \
	--extra-vars='DIST_FILES="$(DIST_FILES)"' \
	--extra-vars='CONFIG=$(CONFIG:$(GIT_ROOT)/%=%)' \
	--extra-vars='GIT_BRANCH=$(GIT_BRANCH)' \
	--extra-vars='GIT_COMMIT=$(GIT_COMMIT)' \
	--extra-vars='EXCEL_DIR=$(EXCEL_DIR)' \
	--extra-vars='EXCEL_DIR_COMMIT=$(EXCEL_DIR_COMMIT)' \
	--extra-vars='systemd_service_state=$(systemd_service_state)' \
	--extra-vars='systemd_service_enabled=$(systemd_service_enabled)'

.PHONY: ansible
## ansible: 检查 run/ansible 目录(脚手架为占位目录, 部署前需按 run/ansible/README.md 初始化)
ansible:
	@test -f run/ansible/playbooks/dist_deploy.yaml || echo "警告: run/ansible 尚未初始化, 详见 run/ansible/README.md"

.PHONY: dist_deploy
## dist_deploy: 将 .tar.gz 部署到远程主机
dist_deploy: ansible tarball
	$(ANSIBLE_PLAYBOOK) $(ANSIBLE_DIR)/playbooks/dist_deploy.yaml \
		$(ANSIBLE_DEPLOY_ARGS) \
		--extra-vars='playbook_hosts=$(playbook_hosts)' \
		--extra-vars='DIST_SERVICES="$(DIST_SERVICES)"'

# 注意: 需要确保 playbook_hosts 指传入 gamesvr 主机组或者相关主机
.PHONY: dist_rolling_update
## dist_rolling_update: 对 gamesvr 执行 rollingUpdate (滚动更新)
dist_rolling_update: ansible tarball
	$(ANSIBLE_PLAYBOOK) $(ANSIBLE_DIR)/playbooks/dist_rolling_update.yaml \
		$(ANSIBLE_DEPLOY_ARGS) \
		--extra-vars='playbook_hosts=gamesvr'

# 注意: 需要确保 playbook_hosts 指传入 gamesvr 主机组或者相关主机
# make dist_rolling_update_only -e CONFIG=run/ansible/config/ansible-prod/ansible-prod-us-central1-NA.yaml -e ANSIBLE_TAGS=deploy
# make dist_rolling_update_only -e CONFIG=run/ansible/config/ansible-prod/ansible-prod-europe-west1-EU.yaml -e ANSIBLE_TAGS=deploy
.PHONY: dist_rolling_update_only
## dist_rolling_update_only: 只对 gamesvr 执行 rollingUpdate 后重启服务, 不更新代码
dist_rolling_update_only: ansible
	$(ANSIBLE_PLAYBOOK) $(ANSIBLE_DIR)/playbooks/dist_rolling_update_only.yaml \
		$(ANSIBLE_DEPLOY_ARGS) \
		--extra-vars='playbook_hosts=gamesvr'

.PHONY: dist_template
## dist_template: 本地渲染配置文件
dist_template:
	$(ANSIBLE_PLAYBOOK) $(ANSIBLE_DIR)/playbooks/dist_template.yaml \
		$(ANSIBLE_PLAYBOOK_ARGS) \
		--extra-vars='playbook_hosts=$(playbook_hosts)'
