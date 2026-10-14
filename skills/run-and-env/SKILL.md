---
name: run-and-env
description: 使用 run/ 脚本与配置模板完成本地环境配置、渲染服务配置、启动/停止服务。
---

# 运行与环境

当你需要配置开发机、渲染服务配置或启停服务时使用此 skill。

## 规则边界

- 本 skill 聚焦环境配置与服务启停流程。
- 验收与构建基线（含 Windows 场景）以 `doc/agent-coding-rules.md` 为准，避免在 skill 中重复维护全局规则。

## 关键位置

- `run/admin.sh`: init/config/start/stop/restart/control/clearlog 的统一入口。
- `run/config/local.yaml`: 本地开发环境配置（脚手架内置示例, 单机全服务）。
- `run/config/machine.yaml`: 本机 IP + 用户名 -> 配置文件映射（多机部署时按需创建）。
- `run/config/*.yaml`: `run/config.py` 使用的个人环境配置。
- `run/*/conf/*.j2`: Jinja2 模板，渲染为服务配置文件。
- `run/tools/db/env.sh.j2`: 数据库环境模板。

## 操作流程

1) 本地开发直接使用内置的 `run/config/local.yaml`（单机 127.0.0.1）；多机环境在 `run/config/machine.yaml` 配置本机内网 IP 和用户名。
2) 渲染配置：
   - `cd run`
   - `bash admin.sh config local.yaml`（显式指定配置文件）
   - 或 `bash admin.sh config`（按 machine.yaml 的 IP 匹配自动选择）
4) 启动服务：
   - `bash admin.sh start`（按当前 `machine` 启动全部服务）
   - `bash admin.sh start gamesvr`（启动单个服务）
5) 停止或重启：
   - `bash admin.sh stop` / `bash admin.sh restart`

## 注意事项

- `admin.sh` 根据 `machine` 环境变量选择服务列表（如 `game`、`develop`、`world`、`area`）。
- `admin.sh init` / `reinit` 仅用于 develop 环境，会执行 `run/tools/db` 下的脚本。
- `run/config.py` 渲染 `*.j2` 模板为配置文件：传参时用指定 yaml，不传参时由 `run/config/machine.yaml` 的 IP + 用户名匹配决定。
