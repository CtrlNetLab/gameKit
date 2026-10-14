package model

// Player roomsvr 侧的玩家视图(脚手架最小实现)
// roomsvr 处理 gamesvr 透传过来的客户端协议, 玩家数据由 gamesvr 维护, 此处仅保留房间内所需信息
type Player struct {
	AccId   uint64
	Sid     string
	ConnIdx uint64
	RoomId  uint64
}

func (player *Player) String() string {
	if player == nil {
		return "<nil player>"
	}
	return "accId:%d|sid:%s|roomId:%d"
}

func (player *Player) GetAccId() uint64 {
	return player.AccId
}
