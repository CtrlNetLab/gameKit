package model

import (
	"time"

	cs "gamekit/pkg/cs"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"gamekit/src/base/utils"

	gnet "github.com/walkon/wsgnet"
)

var (
	PlayerMgr *playerMgr
)

func init() {
	PlayerMgr = &playerMgr{}
}

type playerMgr struct {
	// 链接到玩家的映射, connIdx --> *Player
	connIdxMap map[uint64]*Player

	// accId --> *Player
	accIdMap map[uint64]*Player
}

func (this *playerMgr) Init() {
	this.connIdxMap = make(map[uint64]*Player)
	this.accIdMap = make(map[uint64]*Player)
	utils.AddTimer("PlayerMgr::Tick1s", time.Second, this.Tick1s)
}

// Tick1s 每秒定时器, 主线程执行(见 gamesvr-mainthread skill)
func (this *playerMgr) Tick1s(now time.Time) {
	this.AllPlayerExec(func(player *Player) {
		if player == nil {
			return
		}
		// 业务开发: 在此处理玩家秒级逻辑(如离线检测)
	})
}

// Create 客户端连接建立时创建玩家对象(框架 MainServer.initialConnContext 调用)
func (this *playerMgr) Create(c gnet.Conn, cidx uint64) *Player {
	player := &Player{
		conn:    c,
		connIdx: cidx,
	}
	this.connIdxMap[cidx] = player
	logger.Log.Debug("%s|player create success.", player)
	return player
}

// GetPlayerByConnIdx 通过连接索引查找玩家
func (this *playerMgr) GetPlayerByConnIdx(connIdx uint64) *Player {
	if player, ok := this.connIdxMap[connIdx]; ok {
		return player
	}
	return nil
}

// GetPlayerByAccId 通过账号ID查找玩家
func (this *playerMgr) GetPlayerByAccId(accId uint64) *Player {
	if player, ok := this.accIdMap[accId]; ok {
		return player
	}
	return nil
}

// AddPlayer 玩家登陆成功后加入在线映射(业务登陆流程调用)
func (this *playerMgr) AddPlayer(player *Player) {
	if player == nil {
		return
	}
	this.accIdMap[player.AccId] = player
}

// AllPlayerExec 遍历所有在线玩家执行指定逻辑
func (this *playerMgr) AllPlayerExec(f func(player *Player)) {
	for _, player := range this.connIdxMap {
		f(player)
	}
}

// GetOnlineNum 在线人数
func (this *playerMgr) GetOnlineNum() int {
	return len(this.connIdxMap)
}

// Logout 玩家下线, 清理在线映射
func (this *playerMgr) Logout(player *Player, reason LogoutReason) {
	if player == nil {
		return
	}
	logger.Log.Error("%s|player logout. reason: %d", player, reason)

	delete(this.connIdxMap, player.connIdx)
	if old, ok := this.accIdMap[player.AccId]; ok && old == player {
		delete(this.accIdMap, player.AccId)
	}
	player.loginState = LoginState_Init
}

// 跨天回调, 业务开发时扩展为各模块的跨天处理
func (this *playerMgr) CrossZero() {
	this.AllPlayerExec(func(player *Player) {
		if player != nil && player.IsOnlineOK() {
			logger.Log.Debug("%s|cross zero.", player)
		}
	})
}

// CrossMonth 跨月回调
func (this *playerMgr) CrossMonth() {}

// CrossWeek 跨周回调
func (this *playerMgr) CrossWeek() {}

// CrossFour 跨天4点回调
func (this *playerMgr) CrossFour() {}

// CrossFive 跨天5点回调
func (this *playerMgr) CrossFive() {}

// ServerRollingUpdate 滚动更新, 通知在线玩家重连
func (this *playerMgr) ServerRollingUpdate() {
	this.AllPlayerExec(func(player *Player) {
		if player == nil || !player.IsOnlineOK() {
			return
		}
		player.SendClientNotify(cs.EnmCmdValue_ECV_CSServerRollingUpdateNotify, &cs.CSServerRollingUpdateNotify{})
	})
}

// GenAccId 生成账号ID(脚手架示例: 使用 snowflake, 业务可替换为平台账号体系)
func GenAccId() uint64 {
	return snowflake.GenUniqId(snowflake.IDType_Player)
}

// 编译期检查: 确认 ss 事件枚举存在(框架 OnBoot 中注册的三个事件)
var _ = ss.EventType_EId_StopService
