package model

import (
	"sync"

	"gamekit/src/base/logger"
)

var (
	GlobalMgr *globalMgr
)

func init() {
	GlobalMgr = &globalMgr{}
}

// globalMgr 世界服数据管理(脚手架最小实现)
type globalMgr struct {
	mu     sync.Mutex
	inited bool
	data   map[uint64][]byte
}

// IsInit 是否已初始化(框架 PollerProc 中调用)
func (this *globalMgr) IsInit() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.inited
}

// Init 延迟初始化(首次 PollerProc 时触发, 避免阻塞启动)
func (this *globalMgr) Init() {
	this.mu.Lock()
	defer this.mu.Unlock()
	if this.inited {
		return
	}
	this.inited = true
	this.data = make(map[uint64][]byte)
	logger.Log.Error("GlobalMgr init success.")
}
