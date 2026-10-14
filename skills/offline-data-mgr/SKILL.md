---
name: offline-data-mgr
description: 玩家离线数据补发模式：目标玩家不在线时经 Redis 暂存、登录后拉取分发，以及新增离线类型与处理器映射的实现步骤。用于好友消息、送礼、结算、GM 命令等离线补发场景。
---

# 离线数据补发模式

> 脚手架未内置 OfflineDataMgr 实现，以下为 ac-server 沉淀的可复用模式：
> 在 gamesvr model 层新建 `OfflineDataMgr.go` 并按本模式实现。

## 适用场景

业务侧需要给**不在线**的玩家投递数据（好友消息、送礼、结算奖励、GM 命令等），
不能直接写玩家内存对象，也不能丢弃——需要暂存 + 登录补发。

## 核心流程

1. 定义离线数据消息与类型枚举（放在 `protocol/ss/` 下，业务自建 message）：

```protobuf
// 离线数据类型
enum SSEnmOfflineType {
    EOT_Init = 0;
    EOT_SendGift = 1;
    // ... 业务类型追加
}

// 离线数据载体（一个 message 承载所有类型，按 Type 取对应字段）
message SSOfflineData {
    SSEnmOfflineType Type = 1;
    // ... 各业务字段
}
```

2. 业务侧发现目标玩家不在线时，调用（示例 API，项目自行实现同名 Manager）：

```go
OfflineDataMgr.UpdateOfflineData(receiverId, &ss.SSOfflineData{
    Type: ss.SSEnmOfflineType_EOT_SendGift,
    // GiftSendMsg: giftSendMsg,
})
```

3. `UpdateOfflineData` 将 `SSOfflineData` 做 `proto.Marshal` 后，异步执行 `rdao.ExecScript(...)` 写入 Redis 暂存列表。
4. 玩家登录后调用 `GetAllOfflineData(player)`，异步执行 Lua 拉取离线列表。
5. 拉取到的数据逐条 `proto.Unmarshal` 为 `SSOfflineData`，进入 `operateOfflineData`。
6. `operateOfflineData` 根据 `data.Type` 在 `operateOfflineMap` 找处理器并执行，全部处理完后分发 `ss.EventType_EId_OfflineDataOver`。

## 线程模型约束

- Redis 读写必须在 `utils.WorkerPool.AddTask` 异步任务中执行。
- 涉及玩家对象状态、通知、事件分发等逻辑，放在回调或主线程安全路径执行。
- 处理器函数签名固定为：

```go
type operateOfflineHandler func(player *Player, offlinedata *ss.SSOfflineData)
```

## 新增离线类型步骤

1. 在 ss 协议中新增 `SSEnmOfflineType` 枚举值与 `SSOfflineData` 对应字段。
2. 在 `OfflineDataMgr.init()` 中注册映射：

```go
OfflineDataMgr.operateOfflineMap[ss.SSEnmOfflineType_XXX] = handleOfflineXXX
```

3. 新增 `handleOfflineXXX(player, offlinedata)`：
- 做空值和在线状态校验；
- 保持幂等（避免重复处理造成脏数据）；
- 只处理本类型字段，避免混用其他字段。
4. 在业务触发点（目标玩家离线分支）调用 `UpdateOfflineData` 写入该类型数据。

## 排查要点

- `operateOfflineMap` 未注册会出现 handler not found 错误日志。
- `proto.Unmarshal` 失败通常是协议字段不匹配或写入数据异常。
- 若登录后业务依赖离线数据处理完成，需监听 `ss.EventType_EId_OfflineDataOver`（见 `skills/add-new-event`）。
