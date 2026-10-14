---
name: add-new-event
description: gamekit 事件订阅与派发（EventType 枚举 → pb 重新生成 → AddEventHandler 注册 → DispatchEvent 派发）。
---

# 新增事件订阅

当需要新增一个跨模块事件（如登录成功、数据变更通知、服务停止）时使用此 skill。
事件机制由 `src/base/utils/EventMgr.go` 提供，采用**发布-订阅**模式。

## 规则边界

- 本 skill 说明新增事件的完整链路：proto 枚举 → pb 重新生成 → Handler 注册 → 派发点。
- 跨场景硬约束（并发安全、主线程模型等）以 `doc/agent-coding-rules.md` 为准。
- 若涉及新的 proto message 定义或 pb.go 重新生成，参考 `skills/protocol-and-register/SKILL.md`。

## 前置知识

事件链路如下：

```
proto 定义 EventType 枚举值
  → 业务代码 utils.DispatchEvent
  → EventMgr 按 EventId 查 eventHandlerMap
  → 同步调用所有已注册的 IEventHandler
```

关键文件：

| 路径 | 职责 |
|---|---|
| `protocol/ss/ss_eventtype.proto` | EventType 枚举定义（框架事件已内置，业务事件在此追加） |
| `pkg/ss/ss_eventtype.pb.go` | 生成的 Go 常量 |
| `src/base/utils/EventMgr.go` | `Event` 结构、`AddEventHandler` / `DispatchEvent` |

框架内置事件（各服务 `scripts/framework/MainServer.go` 已注册）：
`EId_Create` / `EId_Login` / `EId_Logout` / `EId_StopService` / `EId_Latency` /
`EId_ServerReady_GlobalSvr` / `EId_ServerReady_GameSvr` / `EId_OfflineDataOver` /
`EId_ActiStart` / `EId_ActiEnd` / `EId_CrossDay` / `EId_RollServerUpdate` /
`EId_OnlinePlayerReconnect` / `EId_CtrlCommand` / `EId_LoginSuccess`。

## 操作流程

### Step 1: 定义 EventType 枚举

在 `protocol/ss/ss_eventtype.proto` 的 `enum EventType` 中新增枚举值：

```protobuf
// 示例功能数据变更，P1：变更类型，P2：变更数量
EId_ExampleChange = 60;
```

- 枚举值**接续已有最大值 +1**，不可重复、不可修改已发布值（枚举值即协议）。
- 注释写明 P1/P2 参数含义，供配置表填表人参考。
- 若需要携带结构化参数，在文件中新增对应的 `EvnXxx` message。

### Step 2: 重新生成 pb.go

```bash
# 全部协议（推荐）
./AllMake.sh proto

# 或只生成 ss 协议
protoc --go_out=paths=source_relative:pkg --proto_path=protocol --proto_path=protocol/share protocol/ss/ss_eventtype.proto
```

确认 `pkg/ss/ss_eventtype.pb.go` 中包含新的常量。

### Step 3: 注册 Handler

在服务初始化处（框架用 `scripts/framework/MainServer.go` 的初始化函数，业务用对应 Manager 的 `Init`）调用：

```go
utils.AddEventHandler(ss.EventType_EId_ExampleChange, this.handleExampleChangeEvent)
```

- 同一事件可注册多个 Handler，按注册顺序同步执行。
- 注册时机必须早于首次派发（服务启动初始化阶段完成）。

### Step 4: 派发事件

在触发该事件的业务代码位置调用：

```go
utils.DispatchEvent(&utils.Event{
    IObj:    player,                              // 实现 utils.IObjString 的对象（无玩家场景可传 nil 之外的最小对象）
    EventId: ss.EventType_EId_ExampleChange,
    Param:   []uint32{uint32(changeType), uint32(count)},
    // EventParam: 需要传递结构化参数时使用（proto.Message，如 &ss.EvnParam{...}）
})
```

**Event 字段说明：**

| 字段 | 说明 |
|---|---|
| `IObj` | 事件关联对象（通常是 player，需实现 `IObjString` 接口；注意避免循环引用） |
| `EventId` | 事件 ID |
| `Param` | `[]uint32` 简单参数，注释中约定 P1/P2 语义 |
| `EventParam` | 结构化参数（proto.Message），与 `Param` 二选一或并用 |

派发点选择原则：
- **阈值型**：在数据发生变化的上游派发（如数据保存后）。
- **累计型**：在行为发生的位置派发（如购买后、充值后）。
- 若事件已在其他系统中有派发点，尽量复用或紧邻已有派发点追加。

### Step 5: 编译验证

```bash
cd src/gamesvr && GOOS=linux GOARCH=amd64 go build ./...
```

## 注意事项

- Handler **同步执行**，禁止在 Handler 中做 Redis/HTTP/DB 等阻塞调用（用 `utils.WorkerPool` 异步化）。
- 没有注册任何 Handler 的事件被派发时，`DispatchEvent` 会打 Error 日志（便于发现漏注册）。
- `IObj` 不要传 nil（`Event.String()` 会解引用）；无关联对象时传一个最小的 `IObjString` 实现。

## 快速检查清单

- [ ] proto 枚举已定义，注释含 P1/P2 说明，编号接续最大值
- [ ] pb.go 已重新生成，常量可用
- [ ] Handler 已通过 `utils.AddEventHandler` 注册，且注册早于首次派发
- [ ] 业务代码中已添加 `utils.DispatchEvent` 调用
- [ ] Handler 内无阻塞调用
- [ ] `go build` 编译通过
