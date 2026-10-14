package model

import (
	"context"
	"fmt"
	"net"
	"time"

	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"benchmark/scripts/config"

	"github.com/gorilla/websocket"
	gnet "github.com/walkon/wsgnet"
)

var (
	LoginMgr *loginMgr = &loginMgr{}
	WsDialer *websocket.Dialer
)

const (
	PlayerIndexKey string = "PlayerIndex"
)

func init() {
	WsDialer = &websocket.Dialer{
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   config.Conf.Common.DialerTimeout,
				KeepAlive: time.Duration(3600) * time.Second,
			}
			return netDialer.DialContext(ctx, network, addr)
		},
		HandshakeTimeout: config.Conf.Common.HandshakeTimeout,
	}
}

type loginMgr struct {
	eng    *gnet.Engine
	client *gnet.Client

	KeepOnlineOK bool
}

func (this *loginMgr) Init(eng *gnet.Engine, client *gnet.Client) {
	this.eng = eng
	this.client = client
	this.KeepOnlineOK = false
}

func (this *loginMgr) IsKeepOnlineOK() bool {
	return this.KeepOnlineOK
}

// StartPlayer 启动一个机器人玩家: 建连 -> 登陆 -> 主循环
func StartPlayer(ctx context.Context) {
	connIdx := snowflake.GenUniqId(snowflake.IDType_ConnIdx)
	player := NewPlayer()
	player.SetConnIdx(connIdx)

	start := time.Now()

	// 初始化连接 : TODO 多服务器随机
	wsConn, _, err := WsDialer.Dial(fmt.Sprintf("%s://%s/game", config.Conf.Network.Proto, config.Conf.Network.GameSvr[0]), nil)
	if err != nil {
		logger.Log.Error("connIdx: %d, dial ws failed. err: %s", connIdx, err.Error())
		return
	}
	logger.Log.Error("connIdx: %d, websocket start conn cost: %v ms", connIdx, time.Since(start).Milliseconds())

	PlayerMgr.AddConnIdx(connIdx, player)
	player.SetWsConn(wsConn)
	player.StartWebSocket(ctx)

	var userName string = fmt.Sprintf("cIdx-%d", connIdx)
	index, ok := ctx.Value(PlayerIndexKey).(int)
	if ok {
		userName = fmt.Sprintf("robot-%d", index)
	}
	player.StartLogin(userName)
	player.StartModuleBench(ctx)
}
