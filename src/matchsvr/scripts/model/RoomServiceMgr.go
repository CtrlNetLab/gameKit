package model

import (
	"sync"
	"time"

	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
)

var (
	RoomServiceMgr *roomServiceMgr
)

func init() {
	RoomServiceMgr = &roomServiceMgr{}
}

// RoomService roomsvr 在 matchsvr 侧的连接视图(脚手架最小实现)
type RoomService struct {
	ConnIdx      uint64
	LastMsgTime  int64
}

func (s *RoomService) UpdateLastMsgTime(subCmdId uint32) {
	s.LastMsgTime = time.Now().UnixMilli()
}

func (s *RoomService) String() string {
	return "roomService"
}

type roomServiceMgr struct {
	mu       sync.Mutex
	services map[uint64]*RoomService
}

func (this *roomServiceMgr) Init() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.services = make(map[uint64]*RoomService)
	utils.AddTimer("RoomServiceMgr::tick1s", time.Second, this.tick1s)
	logger.Log.Error("RoomServiceMgr init success.")
}

// GetRoomService 通过连接索引获取 roomsvr 视图
func (this *roomServiceMgr) GetRoomService(connIdx uint64) *RoomService {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.services[connIdx]
}

// AddRoomService 记录 roomsvr 连接(业务在匹配流程中调用)
func (this *roomServiceMgr) AddRoomService(connIdx uint64) *RoomService {
	this.mu.Lock()
	defer this.mu.Unlock()
	serv := &RoomService{ConnIdx: connIdx, LastMsgTime: time.Now().UnixMilli()}
	this.services[connIdx] = serv
	return serv
}

// RemoveRoomService 移除 roomsvr 连接
func (this *roomServiceMgr) RemoveRoomService(connIdx uint64) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.services, connIdx)
}

func (this *roomServiceMgr) tick1s(now time.Time) {
	// 业务开发: 在此处理 roomsvr 心跳超时检测
}
