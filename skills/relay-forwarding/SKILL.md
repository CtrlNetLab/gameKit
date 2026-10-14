---
name: relay-forwarding
description: gamekit 服务端之间消息转发(relay)机制：CS/SS 协议派发、RelayMgr 封装(Ecv_SSTransReq)、relaysvr TCP 中转、Redis 仅做在线查询、回包靠 Dst=Src。
---

# 服务端间消息转发(relay)机制

用于理解 gamesvr/matchsvr/roomsvr/globalsvr/miscsvr 之间如何通过 relaysvr 转发消息，以及新增跨服协议时如何收发。

## 架构
- 客户端只连 gamesvr，走 CS 协议(pkg/cs)。gamesvr 与其它服务走 SS 协议(pkg/ss)，统一经 relaysvr TCP 中转。
- relaysvr 是 gnet TCP 中转枢纽，不处理业务。Redis 仅存 accId→gamesvr 在线映射(hash onlineKey 的 gid 字段)，消息体走 socket，不走 Redis 队列/PubSub。

## 派发规则(各服务 apps/AppMgr.go 三张表)
- handleClientMap: CS 请求，RegistClientHandle(命令字, handler)，框架 MainServer.OnTraffic → GetClientHandle(cmdId) 分发。
- handleTransmitMap: 经 relay 转发的 SS 消息，RegistTransmitHandle(SubCmdId, handler)；框架识别 Ecv_SSTransReq → HandleTransReq → GetTransmitHandle(SubCmdId)。
- handleServiceMap: 直连 SS(如注册/心跳)。

## 发送封装(发送端必读)
- 统一封装在 src/base/relay/RelayMgr.go：relay.TransMatchSvr/TransGlobalSvr/TransMiscSvr/TransMsg/TransToClientMsg 等，全部汇入 TransMsgImpl(:482)。
- 外层包 SSHead{CmdId: Ecv_SSTransReq, SubCmdId: 真实命令字, Dst: 目标服务类型, AccId, Flag, Src}，子消息序列化进 buffer，ConnAsyncWrite(本服务 relay.conn)。
- Dst 填目标服务类型(如 ServiceType_MatchSvr)；AccId 用于按玩家定位 gamesvr。

## 转发(relaysvr)
- src/relaysvr/scripts/apps/GameApp.go:45 HandleTransReq 按 Flag/Dst/AccId 路由：
  - EMF_ToAllGameSvr → 广播；Dst==0 && AccId!=0 → 查 Redis 取 gid 后 TransBuffer；其它 → ServiceMgr.TransBuffer(:239) 按 Dst 取连接写 socket。
- ServiceMgr.go 用 map[connIdx]*Service 维护在线表，每秒 tick 踢心跳超时。

## 回包
- 接收端处理完把 Dst 设为 head.Src(来源 connIdx)，再 relay.TransMsg 原路返回。SeqId 沿用以对应请求/响应。

## 示例链路(普通 PvP 匹配)
1. Client→gamesvr CSPVPMatchReq → RoomApp.HandlePVPMatchReq → NormalRoomModel.Match
2. gamesvr relay.TransMatchSvr(SSPVPMatchReq) → relaysvr → matchsvr
3. matchsvr PVPMathApp.handlePVPMatchReq → PVPMatchMgr.Add → relay.TransMsg(0, head.Src, ...) 回包
4. gamesvr RoomApp.handlePVPMatchRes → player.SendClientMsg(CSPVPMatchRes)
5. 房间内战斗消息：gamesvr RoomApp.TransmitRoomMsg 透传(relay.TransMsg) 到 roomsvr，结算经 SSRoomSettleNotify 回 gamesvr 再下发 Client

## 常见坑
- 新增协议漏 RegistClientHandle/RegistTransmitHandle → 消息静默丢弃。
- relay.RelayMgr.IsReady() 为 false → RELAY_SEND_FAILED。
- 回包路由全靠 SSHead.Src/Dst/AccId；跨服定位玩家靠 AccId 查 Redis。
