package model

import (
	"sync"
	"sync/atomic"
	"time"

	"gamekit/src/base/logger"
	"gamekit/src/base/utils"

	"benchmark/scripts/config"
)

var (
	PlayerMgr *playerMgr
)

func init() {
	PlayerMgr = &playerMgr{}
}

type playerMgr struct {
	// 链接到玩家的映射, connIdx --> *Player
	cidMapLock sync.RWMutex
	connIdxMap map[uint64]*Player

	// accId --> *Player
	accMapLock sync.RWMutex
	accIdMap   map[uint64]*Player

	// (PlatType + "-" + UserId) --> *Player
	platMapLock sync.RWMutex
	platUserMap map[string]*Player

	SuccessLoginNum int32
}

func (mgr *playerMgr) Init() {
	mgr.connIdxMap = make(map[uint64]*Player)
	mgr.accIdMap = make(map[uint64]*Player)
	mgr.platUserMap = make(map[string]*Player)

	utils.AddTimer("PlayerMgr::tick1s", time.Second, mgr.tick1s)
}

func (mgr *playerMgr) Create(connIdx uint64) *Player {
	player := &Player{
		RecvMsgs: make(chan WsFrame, 128),
		SendMsgs: make(chan WsFrame, 128),
	}
	player.Init(connIdx)
	mgr.cidMapLock.Lock()
	mgr.connIdxMap[connIdx] = player
	return player
}

func (mgr *playerMgr) GetPlayerByConnIdx(connIdx uint64) *Player {
	mgr.cidMapLock.RLock()
	defer mgr.cidMapLock.RUnlock()
	return mgr.connIdxMap[connIdx]
}

func (mgr *playerMgr) Delete(player *Player) {
	connIdx := player.GetConnIdx()

	mgr.cidMapLock.Lock()
	if player, ok := mgr.connIdxMap[connIdx]; ok {
		logger.Log.Error("%s|delete player", player.String())
		delete(mgr.connIdxMap, connIdx)
	} else {
		logger.Log.Error("delete player not found, ConnIdx: %d", connIdx)
	}
	mgr.cidMapLock.Unlock()

	platUser := utils.CombinePlatUser(player.LoginType, player.UserId)
	mgr.platMapLock.Lock()
	delete(mgr.platUserMap, platUser)
	mgr.platMapLock.Unlock()

	accId := player.GetAccId()
	mgr.accMapLock.Lock()
	delete(mgr.accIdMap, accId)
	mgr.accMapLock.Unlock()
}

func (mgr *playerMgr) Logout(player *Player) {
	logger.Log.Error("%s|player logout", player.String())

	conn := player.GetConn()
	mgr.Delete(player)

	if conn != nil {
		conn.Close()
	}
}

func (mgr *playerMgr) AddConnIdx(connIdx uint64, player *Player) {
	mgr.cidMapLock.Lock()
	defer mgr.cidMapLock.Unlock()
	mgr.connIdxMap[connIdx] = player
}

func (mgr *playerMgr) AddPlatUser(loginType int32, userId string, player *Player) {
	platUser := utils.CombinePlatUser(loginType, userId)
	mgr.platMapLock.Lock()
	defer mgr.platMapLock.Unlock()
	mgr.platUserMap[platUser] = player
}

func (mgr *playerMgr) LoginSuccess(player *Player) {
	if player == nil {
		return
	}
	mgr.accMapLock.Lock()
	defer mgr.accMapLock.Unlock()
	mgr.accIdMap[player.GetAccId()] = player

	atomic.AddInt32(&mgr.SuccessLoginNum, 1)
}

func (mgr *playerMgr) GetPlayerByAccId(accId uint64) *Player {
	mgr.accMapLock.RLock()
	defer mgr.accMapLock.RUnlock()
	return mgr.accIdMap[accId]
}

func (mgr *playerMgr) GetPlayerByPlatUser(loginType int32, userId string) *Player {
	platUser := utils.CombinePlatUser(loginType, userId)
	mgr.platMapLock.RLock()
	defer mgr.platMapLock.RUnlock()
	return mgr.platUserMap[platUser]
}

func (mgr *playerMgr) GetOnlineNum() int {
	mgr.cidMapLock.RLock()
	defer mgr.cidMapLock.RUnlock()
	return len(mgr.connIdxMap)
}

func (mgr *playerMgr) IsFull() bool {
	mgr.cidMapLock.RLock()
	defer mgr.cidMapLock.RUnlock()
	return len(mgr.connIdxMap) >= config.Conf.Common.KeepOnline
}

func (mgr *playerMgr) tick1s(now time.Time) {}

func (mgr *playerMgr) StopServer() {
	mgr.cidMapLock.RLock()
	defer mgr.cidMapLock.RUnlock()
	for _, player := range mgr.connIdxMap {
		if player != nil && player.RecvSendCancel != nil {
			player.RecvSendCancel()
		}

		mgr.Logout(player)
	}
}
