---
name: server-architecture
description: gamekit 服务架构总览：入口、配置与主要服务关系。
---

# 服务架构

用于快速了解服务布局、入口文件，以及从哪里开始读代码。

## 服务入口

每个服务都有 `scripts/main.go` 入口：

- `src/gamesvr/scripts/main.go`
- `src/matchsvr/scripts/main.go`
- `src/relaysvr/scripts/main.go`
- `src/roomsvr/scripts/main.go`
- `src/globalsvr/scripts/main.go`
- `src/miscsvr/scripts/main.go`

这些入口展示初始化顺序（配置、控制服务、各类管理器、存储、框架启动）。

## 运行配置

- 模板位于 `run/<service>/conf/*.j2`，由 `run/config.py` 渲染。
- 启动脚本位于 `run/<service>/bin/`，由 `run/admin.sh` 统一编排。

## 通用模式

- 控制通道：各服务都会初始化 `gamekit/src/base/control`。
- 日志与指标：`gamekit/src/base/logger`、`gamekit/src/base/osslog`、`gamekit/src/base/utils`。
- 配置加载：需要读策划配置的服务会用 `gamekit/pkg/configs`。
- Redis/MySQL 初始化通常出现在 `globalsvr`、`gamesvr`、`relaysvr`、`miscsvr`。

## 服务职责（速览）

- `gamesvr`: 主游戏服，初始化多数业务管理器与 DAO。
- `matchsvr`: 匹配与房间协调。
- `relaysvr`: 消息中继与跨服分发，依赖 Redis。
- `roomsvr`: 房间/战斗逻辑相关管理器。
- `globalsvr`: 全局数据与服务管理，依赖 Redis/MySQL。
- `miscsvr`: 杂项功能与 Redis 管理器。

## 需要更深细节时

- 从服务的 `scripts/main.go` 入手，再跟到 `scripts/framework` 与 `scripts/model`。
- 面向客户端的流程可参考 `README.md` 的协议流程图。
