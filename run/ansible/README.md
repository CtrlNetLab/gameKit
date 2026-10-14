# ansible 部署目录(脚手架占位)

本目录在原项目中是一个独立 git submodule，包含各环境(开发/测试/RC/生产)的
主机清单与部署 playbook。脚手架不内置任何环境配置，请按以下步骤初始化:

1. `git init` 后将本目录初始化为 submodule，或直接作为普通目录管理
2. 参考 `run/Makefile` 中的 `dist_deploy` / `dist_rolling_update` 目标,
   它们依赖以下文件:
   - `playbooks/dist_deploy.yaml`        部署 playbook
   - `playbooks/dist_rolling_update.yaml` gamesvr 滚动更新 playbook
   - `config/<env>/<env>-<region>.yaml`  环境变量(对应 run/config/*.yaml 的 Roles)
   - `vault/`                            ansible-vault 加密的密钥文件
3. `run/Makefile` 的 `ANSIBLE_PLAYBOOK_ARGS` 中引用了
   `$(ANSIBLE_HOME)/vault/gamekit.yaml`，请按需修改

部署变量约定(与 run/config/local.yaml 对应):
- `Roles.<svc>.Port`      服务监听端口
- `Roles.<svc>.CtrlPort`  UDP 控制端口
- `Roles.<svc>.HealthPort` 健康检查端口(gamesvr/miscsvr)
- `Roles.<svc>.ServiceId`  服务 ID
- `dist_user.Roles`       部署渲染时由 config.py 注入(单机多实例部署时使用)
