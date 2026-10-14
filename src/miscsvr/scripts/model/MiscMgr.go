package model

import (
	"sync"
	"time"

	"gamekit/src/base/logger"
	"gamekit/src/base/utils"

	gnet "github.com/walkon/wsgnet"
)

var (
	MiscMgr *miscMgr
)

func init() {
	MiscMgr = &miscMgr{}
}

// miscMgr 管理 miscsvr 与 gamesvr 之间的长连接(脚手架最小实现)
type miscMgr struct {
	mu    sync.Mutex
	conns map[uint64]gnet.Conn
}

func (this *miscMgr) Init() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.conns = make(map[uint64]gnet.Conn)
	utils.AddTimer("MiscMgr::tick1s", time.Second, this.tick1s)
	logger.Log.Error("MiscMgr init success.")
}

// AddConn 记录 gamesvr 连接(框架在收到 ECV_SSMiscReq 时调用)
func (this *miscMgr) AddConn(c gnet.Conn) {
	this.mu.Lock()
	defer this.mu.Unlock()
	if connIdx, ok := c.Context().(uint64); ok {
		this.conns[connIdx] = c
		logger.Log.Error("MiscMgr add conn. connIdx: %d", connIdx)
	}
}

// GetConn 获取指定 gamesvr 连接
func (this *miscMgr) GetConn(connIdx uint64) gnet.Conn {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.conns[connIdx]
}

// RemoveConn 移除连接
func (this *miscMgr) RemoveConn(connIdx uint64) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.conns, connIdx)
}

func (this *miscMgr) tick1s(now time.Time) {
	// 业务开发: 在此处理连接心跳超时检测
}
