---
name: redis-usage
description: Redis 使用规范、异步调用约束、Lua 脚本加载与 ExecScript 约定。
---

# Redis 使用规范

当你需要调用 Redis 或新增 Lua 脚本时使用此 skill。

## 规则边界

- 本 skill 聚焦 Redis/Lua 的落点与调用样式。
- 跨场景硬约束（异步执行、主线程回调、`ExecScript` 语义）以 `doc/agent-coding-rules.md` 为准。

## 基本规则

- Redis 操作按异步模型执行（避免阻塞主线程），硬约束以 `doc/agent-coding-rules.md` 为准。
- 简单查询优先检查 `RedisMgr.go` 是否已有封装方法。
- 涉及玩家或公共内存修改的逻辑，回到主线程回调中执行。

相关代码：
- `src/base/dao/RedisMgr.go`
- `src/base/utils/WorkerPool.go`
- `src/gamesvr/scripts/framework/MainServer.go`

## 推荐用法

```go
utils.WorkerPool.AddTask(player.AccId, "taskName", func() {
    // redis 操作...

    // 如需回调到主线程操作玩家/公共数据
    utils.WorkerPool.AddCallback("taskName", func() {
        // 主线程逻辑
    })
})
```

## Lua 脚本规则

- 复杂 Redis 操作用 Lua 实现，新脚本放在 `run/redislua/`。
- 在服务配置的 Redis Scripts 中加载脚本路径：
  - `run/<service>/conf/service.yaml.j2`（按实际服务选择, 如 `run/gamesvr/conf/service.yaml.j2`）

## ExecScript 约定

示例：

```go
utils.WorkerPool.AddTask(player.AccId, "taskName", func() {
// redis 操作... 
	rdao.ExecScript(rdao.GameCache_RedisName, rdao.Script, accId, redisArgs...)
// 如需回调到主线程操作玩家/公共数据
utils.WorkerPool.AddCallback("taskName", func() {
// 主线程逻辑
})
})
```

参数说明：

1) 第一个参数是 Redis 名称 key（决定走哪个 Redis 配置）。
2) 第二个参数是 Lua 脚本命名 key（需在 `service.yaml.j2` 的 Scripts 中注册）。
3) 第三个参数是玩家 `accId`（用于分片/路由）。
4) 第四个参数是传给 Lua 的参数数组。

Redis 名称 key 定义（见 `src/base/dao/RedisMgr.go`）：

- `Online_RedisName`
- `AccId_RedisName`
- `GameCache_RedisName`
- `Global_RedisName`
- `OnlineGlobal_RedisName`
