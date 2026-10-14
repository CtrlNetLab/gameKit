package model

import (
	"benchmark/scripts/config"
	"context"
	"errors"
	"fmt"
	"time"

	"gamekit/pkg/cs"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	gnet "github.com/walkon/wsgnet"
)

type WsFrame []byte

const (
	heartTimeOut int64 = 30
)

const (
	LoginState_Init    int = 0
	LoginState_Success int = 1
)

type Player struct {
	UserId    string
	LoginType int32
	AccId     uint64

	// 保存玩家连接
	conn   gnet.Conn
	WsConn *websocket.Conn

	SendMsgs chan WsFrame
	RecvMsgs chan WsFrame
	PingMsg  chan []any

	RecvSendCtx    context.Context
	RecvSendCancel context.CancelFunc

	connIdx uint64

	// 登陆状态
	loginState int

	// 玩家数据
	playerData *cs.CS_AC_LoginRes

	heartUnix int64

	// 命令耗时统计: reqCmdId --> 发包时间
	CmdStatMap map[cs.EnmCmdValue]*time.Time
}

func (player *Player) Init(cidx uint64) {
	player.connIdx = cidx
	player.loginState = LoginState_Init
	player.playerData = nil
	player.CmdStatMap = make(map[cs.EnmCmdValue]*time.Time)
}

func (player *Player) Init_New() {
	player.Init(0)
}

func NewPlayer() *Player {
	player := &Player{
		RecvMsgs: make(chan WsFrame, 128),
		SendMsgs: make(chan WsFrame, 128),
		PingMsg:  make(chan []any, 16),
	}
	player.Init_New()
	return player
}

// StartLogin 发起登陆(脚手架最小示例, 可按业务替换登陆协议)
func (player *Player) StartLogin(loginAccount string) {
	if player.loginState != LoginState_Init {
		logger.Log.Error("%s|start login failed. State: %d", player, player.loginState)
		return
	}
	userId := loginAccount
	var loginType int32 = 0
	token := fmt.Sprintf("token-%d", utils.RandInt64(1000000, 9999999))

	loginReq := &cs.CS_AC_LoginReq{
		LoginType:     loginType,
		PlatformId:    userId,
		PlatformToken: token,
		TouristId:     userId,
		DeviceId:      uuid.NewString(),
		ClientVer:     "1.0.0",
		Lang:          "zh",
	}

	player.SetUserInfo(loginType, userId)
	player.SendClientMsg(cs.EnmCmdValue_ECV_CS_AC_LoginReq, cs.EnmRetCode_SUCCESS, loginReq)
}

// StartModuleBench 登陆后的机器人主循环: 心跳 + 收包分发
// 业务压测模块可参照此结构扩展(按 QPS 发送业务请求)
func (player *Player) StartModuleBench(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			player.TrySendHeart(time.Now().Unix())
		case buff := <-player.RecvMsgs:
			DispatchMsg(player, buff)
		}
	}
}

func (player *Player) SetLoginState(state int) {
	logger.Log.Debug("%s|set player state. NewState: %d", player.String(), state)
	player.loginState = state
}

func (player *Player) SetUserInfo(loginType int32, userId string) {
	player.UserId = userId
	player.LoginType = loginType
	PlayerMgr.AddPlatUser(loginType, userId, player)
}

func (player *Player) String() string {
	if player.playerData != nil {
		return fmt.Sprintf("[UserId:%s, cidx: %d, State: %d, AccId: %d]", player.UserId, player.connIdx, player.loginState, player.GetAccId())
	}
	return fmt.Sprintf("[UserId:%s, cidx: %d, State: %d]", player.UserId, player.connIdx, player.loginState)
}

func (player *Player) GetConn() gnet.Conn {
	return player.conn
}

func (player *Player) GetLoginState() int {
	return player.loginState
}

func (player *Player) LoginSuccess(playerData *cs.CS_AC_LoginRes) {
	player.playerData = playerData
	player.AccId = playerData.AccId

	player.SetLoginState(LoginState_Success)
	player.heartUnix = time.Now().Unix() + heartTimeOut
	PlayerMgr.LoginSuccess(player)

	logger.Log.Error("%s|login success", player)
}

func (player *Player) SetConnIdx(connIdx uint64) {
	player.connIdx = connIdx
}

func (player *Player) SetConn(conn gnet.Conn) {
	player.conn = conn
}

func (player *Player) SetWsConn(wsConn *websocket.Conn) {
	player.WsConn = wsConn
}

func (player *Player) GetConnIdx() uint64 {
	return player.connIdx
}

func (player *Player) GetAccId() uint64 {
	if player.IsOnlineOK() {
		return player.playerData.AccId
	}
	return 0
}

func (player *Player) IsOnlineOK() bool {
	return player.loginState == LoginState_Success
}

func (player *Player) GetPlayerData() *cs.CS_AC_LoginRes {
	return player.playerData
}

func (player *Player) TrySendHeart(nowUnix int64) {
	if !player.IsOnlineOK() || player.heartUnix > nowUnix {
		return
	}
	player.heartUnix = nowUnix + heartTimeOut

	heartReq := &cs.CSHeartReq{}
	player.SendClientMsg(cs.EnmCmdValue_ECV_CSHeartReq, cs.EnmRetCode_SUCCESS, heartReq)
	player.PingMsg <- []any{nowUnix}
}

func (player *Player) SendClientMsg(cmdId cs.EnmCmdValue, retCode cs.EnmRetCode, message utils.IMessage) error {
	seqId := GenSeqId()

	csHead := utils.GetCSHead(cmdId, seqId, retCode, "")
	switch config.Conf.Network.Proto {
	case "tcp":
		return utils.SendMsg(player.conn, csHead, message)
	case "ws":
		fallthrough
	case "wss":
		return SendWsMsgChan(player, csHead, message)
	default:
		return errors.New("unknow proto")
	}
}

func (player *Player) StartWebSocket(ctx context.Context) {
	player.RecvSendCtx, player.RecvSendCancel = context.WithCancel(ctx)
	go RecvWsMsg(player)
	go SendWsMsg(player)
}
