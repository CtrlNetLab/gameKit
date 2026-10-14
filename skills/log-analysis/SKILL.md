---
name: log-analysis
description: 日志分析 skill - 当用户说"分析日志"、"帮我看日志"、"日志报错"、"log 有问题"、"排查 gamesvr/matchsvr/relaysvr/roomsvr/globalsvr/miscsvr 问题"、"看看这个 log"、"log 里有异常"、"检查异常操作"时必须触发此 skill。触发后先确认日志位置（本地默认 run/<service>/log/，与玩家相关默认 gamesvr），再识别进程并聚焦对应代码目录；Loki 查询使用 service_type label。
---

# 日志分析 Skill

## 第一步：确认日志位置与环境

**触发 skill 后立即询问用户**，不要直接开始分析。先确认两件事：

### 1.1 确认日志位置

> 请提供日志位置或内容。常见默认位置：
>
> - **本地工程**（开发机）默认 `run/<service>/log/<service>.log`
>   - 与玩家相关问题默认查 `run/gamesvr/log/gamesvr.log`
>   - 其他服务：`run/matchsvr/log/`、`run/relaysvr/log/`、`run/roomsvr/log/`、`run/globalsvr/log/`、`run/miscsvr/log/`
> - **内网测试服务器** 默认走 Loki MCP 工具查询

如果用户没有明确给路径，按以下默认：

| 上下文关键字             | 默认日志位置 / Loki label         |
| ------------------------ | --------------------------------- |
| 玩家/登录/英雄/关卡/资源 | `run/gamesvr/log/gamesvr.log`     |
| 匹配/排位                | `run/matchsvr/log/matchsvr.log`   |
| 中继/网关                | `run/relaysvr/log/relaysvr.log`   |
| 房间/战斗服              | `run/roomsvr/log/roomsvr.log`     |
| 全局服/世界服            | `run/globalsvr/log/globalsvr.log` |
| 杂项                     | `run/miscsvr/log/miscsvr.log`     |

### 1.2 确认日志环境

> 日志来源环境是？
>
> 1. **当前开发工程** - 本地日志文件，可直接读源码
> 2. **内网测试服务器** - 已收集到 Loki，通过 Loki MCP 工具查询
> 3. **其他**（请说明）

等用户回复后继续。

---

## 第二步：识别进程，确定代码范围

从日志内容识别是哪个进程的日志，**只查找该进程对应的源码目录**。

### 进程识别规则

日志中的 `(包路径/文件:行号)` 字段包含进程标识：

| 日志包路径前缀          | 进程      | 源码目录                 | 本地日志路径                      |
| ----------------------- | --------- | ------------------------ | --------------------------------- |
| `gamesvr/scripts/...`   | gamesvr   | `src/gamesvr/scripts/`   | `run/gamesvr/log/gamesvr.log`     |
| `matchsvr/scripts/...`  | matchsvr  | `src/matchsvr/scripts/`  | `run/matchsvr/log/matchsvr.log`   |
| `relaysvr/scripts/...`  | relaysvr  | `src/relaysvr/scripts/`  | `run/relaysvr/log/relaysvr.log`   |
| `roomsvr/scripts/...`   | roomsvr   | `src/roomsvr/scripts/`   | `run/roomsvr/log/roomsvr.log`     |
| `globalsvr/scripts/...` | globalsvr | `src/globalsvr/scripts/` | `run/globalsvr/log/globalsvr.log` |
| `miscsvr/scripts/...`   | miscsvr   | `src/miscsvr/scripts/`   | `run/miscsvr/log/miscsvr.log`     |
| `base/...`              | 公共库    | `src/base/`              | （无独立服务日志）                |

也可以从日志分类字段 `[分类|用户@IP]` 辅助判断，例如 `[MainLog|...]` 通常是 gamesvr 主进程。

> **小贴士**：与玩家直接相关的日志（登录、英雄、关卡、资源、邮件、好友、PVP 等）几乎都在 gamesvr，对应 `run/gamesvr/log/gamesvr.log`。先看这个文件能解决大部分问题。

**确认进程后，告知用户**："检测到日志来自 `<进程名>`，本地路径 `<run-path>`，后续代码查找范围限定在 `src/<进程名>/scripts/`。"

---

## 第三步：获取日志

### 环境 1：当前开发工程

日志在本地，默认按 `run/<service>/log/<service>.log` 定位，常用文件：

- `run/gamesvr/log/gamesvr.log` - 主业务日志
- `run/gamesvr/log/gamesvr_monitor.log` - 监控/统计日志
- `run/gamesvr/log/gamesvr_oss.log` - OSS/外部服务日志
- `run/<service>/log/stdout` - 启动与 panic 输出

**读取建议**：

- 当天日志在 `gamesvr.log`，历史日志在 `gamesvr.log.YYYYMMDD.N`（按 100M rotate）。
- 玩家操作类问题先看 `gamesvr.log`；若是网络/超时/外部依赖问题再叠加 `gamesvr_oss.log`。
- 大文件优先用 `grep`/`rg` 过滤后再 `Read`，避免一次性加载整个文件。

### 环境 2：内网测试服务器（Loki）

内网服务器日志已统一收集到 Loki，使用 **Loki MCP 工具**查询，无需 SSH。

> **Loki label 约定**：本项目日志使用 `service_type` 标识服务类型（不是 `job`），查询时必须显式带上 `service_type="<service>"` 才能命中目标服务日志。

**常用 LogQL 查询模板：**

```logql
# 查询某服务的错误日志（最近 1 小时）
{service_type="gamesvr"} |= "[EROR]"

# 按关键词过滤
{service_type="gamesvr"} |= "config not found"

# 查询某用户的日志（AccId）
{service_type="gamesvr"} |= "3516010829381956"

# 查询多个级别
{service_type="gamesvr"} |~ "\[EROR\]|\[FATAL\]"

# 多服务联合查询（如玩家问题同时看 gamesvr 和 roomsvr）
{service_type=~"gamesvr|roomsvr"} |= "3516010829381956"
```

**查询规范**：

- 始终以 `{service_type="<service>"}` 起手，`<service>` 取值 `gamesvr` / `matchsvr` / `relaysvr` / `roomsvr` / `globalsvr` / `miscsvr` 之一。
- 玩家相关问题默认 `service_type="gamesvr"`，需要联动战斗/房间时再加 `roomsvr`。
- 不确定可用 label 时，用 `loki_label_values` 查 `service_type` 的取值后再构造查询。

---

## 第四步：分析日志

### 日志格式

```
[时间戳] [分类|用户@IP] [级别] (包路径/文件.类型.方法:行号) 用户ID|消息
```

示例：

```
[2026/05/26 21:35:47.623416] [MainLog|ac10@10.162.0.7] [EROR] (gamesvr/scripts/framework.(*MainServer).OnClose:229) 3516010829381956|error occurred on closed, err: read: connection reset by peer
```

`(gamesvr/scripts/framework.(*MainServer).OnClose:229)` → 源码在 `src/gamesvr/scripts/framework/` 下，行号 229。

### 错误级别

| 级别      | 含义         | 处理                 |
| --------- | ------------ | -------------------- |
| `[FATAL]` | 服务即将退出 | 立即排查             |
| `[EROR]`  | 操作失败     | 判断是代码/配置/网络 |
| `[WARN]`  | 意外但可恢复 | 评估是否有设计问题   |
| `[INFO]`  | 正常操作     | 排查时通常忽略       |
| `[DEBG]`  | 详细调试     | 生产环境不应存在     |

### 根因分类

**代码逻辑问题**（研发修复）：

- `nil pointer` / `panic` / `index out of bounds`
- `config not found` / `configId is` / `quality value was illegal`
- `eventToken is nil` / `parse.*error` / `json.*error`
- `deadlock` / `race condition` / `duplicate key`

**配置问题**（运营/策划修复）：

- `activity.*not open` / `activity.*end`
- `feature disabled` / `not enabled` / `mapping not found`

**网络/环境问题**（监控）：

- `EOF` / `timeout` / `connection reset` / `broken pipe`
- `adjust.*error` / `firebase.*error`

**正常行为**（忽略）：

- `connection reset by peer` / `error occurred on closed` / `context canceled`

---

## 第五步：代码定位（仅当环境为"当前开发工程"时）

识别到需要分析代码时，**只在该进程的源码目录内查找**。

使用 CodeGraph 工具精确定位：

- `codegraph_search` 按方法名查找定义
- `codegraph_node` 查看方法源码与签名
- `codegraph_callers` 追溯调用链

---

## 输出格式

每个问题按以下格式输出：

```
### [问题类型]
- **错误类型**: 具体错误模式
- **出现次数**: 时间内总次数
- **根因分类**: 代码逻辑 / 配置问题 / 网络环境 / 正常行为
- **代码位置**: `包路径/文件.方法:行号`（从日志直接提取）
- **影响范围**: 受影响用户数/比例
- **建议**: 修复代码 / 更新配置 / 监控 / 忽略
```

代码位置填写规则：直接复制日志 `(...)` 内的内容，去掉括号；多个位置列 Top 3；无位置信息则注明。
