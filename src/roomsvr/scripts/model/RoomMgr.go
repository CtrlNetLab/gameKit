package model

import (
	"sync"
	"time"

	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
)

var (
	RoomMgr *roomMgr
)

func init() {
	RoomMgr = &roomMgr{}
}

// Room 房间(脚手架最小实现)
type Room struct {
	RoomId   uint64
	CreateTm int64
	Players  map[uint64]*Player
}

type roomMgr struct {
	mu    sync.Mutex
	rooms map[uint64]*Room
}

func (this *roomMgr) Init() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.rooms = make(map[uint64]*Room)
	utils.AddTimer("RoomMgr::tick1s", time.Second, this.tick1s)
	logger.Log.Error("RoomMgr init success.")
}

// CreateRoom 创建房间
func (this *roomMgr) CreateRoom(roomId uint64) *Room {
	this.mu.Lock()
	defer this.mu.Unlock()
	room := &Room{RoomId: roomId, CreateTm: time.Now().Unix(), Players: make(map[uint64]*Player)}
	this.rooms[roomId] = room
	return room
}

// GetRoom 获取房间
func (this *roomMgr) GetRoom(roomId uint64) *Room {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.rooms[roomId]
}

// DestroyRoom 销毁房间
func (this *roomMgr) DestroyRoom(roomId uint64) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.rooms, roomId)
}

func (this *roomMgr) tick1s(now time.Time) {
	// 业务开发: 在此处理房间超时销毁等逻辑
}
