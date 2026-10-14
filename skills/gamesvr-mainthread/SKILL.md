---
name: gamesvr-mainthread
description: gamesvr 主线程模型与协程使用规则，包括 WorkerPool 的异步任务与回调约束。
---

# gamesvr 主线程模型

用于解释 gamesvr 的主线程/协程模型，以及什么时候使用 `WorkerPool` 的不同接口。

## 规则边界

- 本 skill 只说明 gamesvr 并发模型的实现入口与使用方式。
- 跨场景硬约束（如 `OnTick` 禁业务、`AddTask -> AddCallback`）以 `doc/agent-coding-rules.md` 为准。

## 线程模型要点

- gamesvr 主逻辑在 `MainServer.PollerProc` 内单线程执行，主循环每帧会处理回调与定时器。
- `MainServer.OnTick` 明确标注为子线程，禁止在其中添加业务逻辑。
- 项目不是严格“全局单线程”，而是**主逻辑单线程 + 辅助协程并发**。

相关代码：
- `src/gamesvr/scripts/framework/MainServer.go`

## WorkerPool 使用规则

- `WorkerPool.AddTask(primary, taskName, task)`:
  - 任务进入 worker 队列（按 `primary` 分片），在后台协程执行。
  - 适合 Redis/HTTP 等耗时操作。
- `WorkerPool.AddCallback(name, cb)`:
  - 将回调加入主线程队列。
  - 回调在 `MainServer.PollerProc` 内调用 `WorkerPool.Callback()` 时执行。
  - **涉及公共内存/游戏数据的操作必须放在这里执行**。
- `WorkerPool.Do(primary, taskName, task)`:
  - 直接起 goroutine 执行，只用于“延迟性特别高”的任务（源码注释）。
  - 不进入 worker 队列，不保证串行。

相关代码：
- `src/base/utils/WorkerPool.go`
- `src/gamesvr/scripts/framework/MainServer.go`

## 实践建议

- I/O 密集：`AddTask` + `AddCallback` 组合。
- 数据修改：只在主线程回调里进行。
- `OnTick` 内避免业务逻辑，具体约束遵循 `doc/agent-coding-rules.md`。
