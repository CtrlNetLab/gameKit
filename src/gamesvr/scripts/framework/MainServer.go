package framework

import (
	cs "gamekit/pkg/cs"
	"gamekit/pkg/game_third"
	"gamekit/pkg/oss"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	relay "gamekit/src/base/relay"
	"gamekit/src/base/snowflake"
	"gamekit/src/base/utils"
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"gamesvr/scripts/apps"
	"gamesvr/scripts/config"
	"gamesvr/scripts/model"
	"math/rand"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gobwas/ws"
	"github.com/pierrec/lz4"

	gnet "github.com/walkon/wsgnet"
	gerrors "github.com/walkon/wsgnet/pkg/errors"
	"github.com/xxtea/xxtea-go/xxtea"

	"google.golang.org/protobuf/proto"
)

type ServerState int

const (
	ServerState_Init    ServerState = 0
	ServerState_NoReady ServerState = 1
	// 正常运行阶段
	ServerState_Ready ServerState = 2
	// 准备退出阶段，需要做停服退出准备，不再处理客户端的请求
	ServerState_PreStop ServerState = 3
	// 退出
	ServerState_Stop ServerState = 4
	//滚服更新
	ServerState_RollingUpdate ServerState = 5
)

const (
	LengthFieldLength int = 4
)

var (
	UnknownRequestUri error = errors.New("unknown request uri")
	NotGameRequest    error = errors.New("not game request")
)

type MainServer struct {
	// gnet.BuiltinEventEngine
	eng gnet.Engine

	// svr          gnet.Server
	network      string
	addr         string
	multicore    bool
	async        bool
	started      int32
	connected    int32
	clientActive int32
	disconnected int32
	// workerPool   *goroutine.Pool
	codec gnet.ICodec

	state ServerState
	// 健康检测失败开始时间(秒)
	unHealthStartTime int64
	// 是否通知玩家滚动更新
	notifyRollingUpdate bool
	// tick 相关
	// 1s tick 最后一次 tick 的时间戳
	tick1sTm int64

	// 人为制造时间延迟，单位: 毫秒
	latency int

	// 系统信号处理
	signalChan chan os.Signal
}

var ms *MainServer

func GetMainServer() *MainServer {
	return ms
}

func (this *MainServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	this.eng = eng
	this.state = ServerState_NoReady
	this.tick1sTm = time.Now().UnixNano() / 1e6

	relay.RelayMgr.Init(config.GetServiceId(), int32(ss.ServiceType_GameSvr), config.GetAreaId(), &eng, logger.Log)
	utils.AddEventHandler(ss.EventType_EId_StopService, this.handleStopServiceEvent)
	utils.AddEventHandler(ss.EventType_EId_RollServerUpdate, this.handleRollingUpdateEvent)
	utils.AddEventHandler(ss.EventType_EId_Latency, this.latencyEvent)
	utils.AddTimer("MainServer::tick1s", time.Second, this.tick1s)
	rand.Seed(time.Now().UnixNano())

	return
}

func (this *MainServer) PollerPreInit() {
	now := time.Now().UnixNano() / 1e6
	this.tick1sTm = now

	logger.Log.Debug("PollerPreInit ...")
}

// epoll_wait timeout, 默认值 10ms
func (this *MainServer) PollerWaitTimeOut() int {
	return 10
}

func (this *MainServer) isReady() bool {
	return (this.state == ServerState_Ready || this.state == ServerState_RollingUpdate)
}

func (this *MainServer) isRollingUpdate() bool {
	return this.state == ServerState_RollingUpdate
}

func (this *MainServer) handleStopServiceEvent(event *utils.Event) {
	model.PlayerMgr.AllPlayerExec(func(player *model.Player) {
		if player == nil || !player.IsOnlineOK() {
			return
		}
		model.PlayerMgr.Logout(player, model.LogoutReason_Normal)
	})

	this.state = ServerState_PreStop
	logger.Log.Error("service stoping ...")
}

func (this *MainServer) handleRollingUpdateEvent(event *utils.Event) {
	this.state = ServerState_RollingUpdate
	logger.Log.Error("service rollingUpdate ...")
}

func (this *MainServer) latencyEvent(event *utils.Event) {
	if event == nil || len(event.Param) != 1 {
		return
	}
	this.latency = int(event.Param[0])
}

func (this *MainServer) isStop() bool {
	return (this.state == ServerState_PreStop || this.state == ServerState_Stop)
}

func (this *MainServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if this.isStop() {
		action = gnet.Close
		logger.Log.Error("[OnOpened] server stoping, cann't process anything, state: %d", this.state)
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
			logger.Log.Error("panic: %s", osslog.OSS.PanicTrace(err))
		}
	}()

	atomic.AddInt32(&this.connected, 1)

	// 服务之间的连接
	if c.Context() != nil {
		// service connect success
		logger.Log.Error("service connect success. connIdx: %#v", c.Context())
		return
	}

	if !this.isReady() {
		action = gnet.Close
		logger.Log.Error("gamesvr not ready, can't support service ...")
		return
	}

	return
}

func (this *MainServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if this.isStop() {
		logger.Log.Error("[OnClosed] server stoping, cann't process anything")
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
			logger.Log.Error("panic: %s", osslog.OSS.PanicTrace(err))
		}
	}()

	atomic.AddInt32(&this.disconnected, 1)

	iContext := c.Context()
	if hc, ok := iContext.(string); ok && hc == "healthCheck" {
		return
	}

	connIdx, cok := iContext.(uint64)
	if !cok {
		// 客户端连接上来之后就立刻关闭，不需要做任何处理
		return
	}

	if err != nil {
		logger.Log.Error("%d|error occurred on closed, err: %v", connIdx, err)
	}

	logger.Log.Error("OnClosed, ConnIdx: %d", connIdx)

	// service close
	if connIdx < utils.MaxServiceConnIdx {
		relay.RelayMgr.CloseRelay(connIdx, err)
		return
	}

	player := model.PlayerMgr.GetPlayerByConnIdx(connIdx)
	if nil == player {
		// 主动 Close 也会回调，无法区分是主动 close 还是客户端 close，不过只要是 nil 一般都是服务器主动 close
		logger.Log.Error("close connect callback. not found player. ConnIdx: %d", connIdx)
	} else {
		if utils.IsNetLog() {
			logger.Log.Debug("%s|connect close, player logout.", player.String())
		}
		player.Disconnect(model.LogoutReason_CloseSock)
	}

	return
}

func isNoPrintLogCmd(cmdId cs.EnmCmdValue) bool {
	// @TODO ryder: 密码等关键信息不能输出到日志当中
	return false
}

func (this *MainServer) acReactClient(player *model.Player, frame []byte, c gnet.Conn) (action gnet.Action) {
	if player == nil {
		return
	}

	if !this.isReady() {
		action = gnet.Close
		logger.Log.Error("gamesvr not ready, can't support service ...")
		return
	}

	var length uint32 = uint32(len(frame))
	if length <= 2 || length < (2+uint32(frame[1])) {
		action = gnet.Close
		logger.Log.Error("%s|length too small. length: %d", player.String(), length)
		return
	}

	var (
		xxteaKey []byte = []byte(config.Conf.Common.XXTeaKey)
		buffer   []byte
		err      error
	)

	// 加密 && 压缩
	bitmask := uint32(frame[0])
	// 头部长度
	headlen := uint32(frame[1])

	buffer = frame[2:]

	// 解密
	if (bitmask & 0xf) != 0 {
		buffer = xxtea.Decrypt(buffer, xxteaKey)
	}

	// 解压缩
	if (bitmask & 0xf0) != 0 {
		compressBuffer := make([]byte, 10*len(buffer))
		n, err := lz4.UncompressBlock(buffer[4:], compressBuffer)
		if err != nil {
			logger.Log.Error("lz4 uncompress failed:%s", err.Error())
			action = gnet.None
			return
		}

		buffer = compressBuffer[:n]
	}

	if len(buffer) < int(headlen) {
		action = gnet.None
		logger.Log.Error("%s|buffer is too small. buf.len: %d, headlen: %d, frame: %#v", player.String(), len(buffer), headlen, frame)
		return
	}

	head := &cs.CSHead{}
	err = proto.Unmarshal(buffer[:headlen], head)
	if err != nil {
		action = gnet.None
		logger.Log.Error("%s|unpack head failed. length: %d, headlen: %d, err: %s", player.String(), length, headlen, err.Error())
		return
	}

	now := time.Now()
	defer func() {
		err := recover()
		if err != nil {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
			logger.Log.Error("panic: %s", osslog.OSS.PanicTrace(err))
		}
		cost := utils.TimeSince(now)
		utils.StatMgr.StatCmdCall(head.GetCmdId(), len(frame), cost)
	}()

	imessage, ok := cs.NewClientMessage(int32(head.GetCmdId()))
	if !ok || imessage == nil {
		action = gnet.None
		logger.Log.Error("%s|not found message. CmdId: %d", player.String(), head.GetCmdId())
		return
	}

	moffset := headlen
	err = proto.Unmarshal(buffer[moffset:], imessage)
	if err != nil {
		action = gnet.None
		logger.Log.Error("%s|unpack message failed. Head: %s, Err: %s", player.String(), head.String(), err.Error())
		return
	}

	cmdId := cs.EnmCmdValue(head.GetCmdId())
	if utils.IsNetLog() && !isNoPrintLogCmd(cmdId) {
		logger.Log.Debug("%s|Recv Message(client). Head: %s, Message: %s", player.String(), head.String(), imessage.String())
	}

	if !player.IsOnlineOK() {
		if cmdId != cs.EnmCmdValue_ECV_CS_AC_LoginReq &&
			cmdId != cs.EnmCmdValue_ECV_CSHeartReq {
			// 未登陆成功就发其他协议，直接返回
			logger.Log.Error("%s|player is not online ok, but have other message. Head: %s", player.String(), head.String())
			return
		}
	}

	player.SetLastRecvMsg()

	// @TODO ryder: 直接转发到 roomsvr, gamesvr 无需处理
	// 直接转发到 roomsvr, gamesvr 无需处理
	if cmdId > cs.EnmCmdValue_ECV_CSTransRoomReq {
		player.SendRoomMsg(head.RoomId, head, imessage)
		return
	}

	handler := apps.GetClientHandle(uint32(cmdId))
	if handler == nil {
		logger.Log.Error("%s|not found handler. Head: %s", player.String(), head.String())
		return
	}

	handler(player, head, imessage)

	return
}

func (this *MainServer) reactService(connIdx uint64, frame []byte, c gnet.Conn) (action gnet.Action) {
	var length uint32 = uint32(len(frame))
	if length <= 2 || length < (2+uint32(frame[1])) {
		action = gnet.Close
		logger.Log.Error("%d|length too small. length: %d", connIdx, length)
		return
	}

	// 加密 && 压缩
	// bit := uint32(frame[0])
	// 头部长度
	headlen := uint32(frame[1])
	head := &ss.SSHead{}

	if err := proto.Unmarshal(frame[2:2+headlen], head); err != nil {
		action = gnet.Close
		logger.Log.Error("%d|unpack head failed. length: %d, headlen: %d", connIdx, length, headlen)
		return
	}

	now := time.Now()
	defer func() {
		err := recover()
		if err != nil {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
			logger.Log.Error("panic: %s", osslog.OSS.PanicTrace(err))
		}
		utils.StatMgr.StatCmdCall(head.GetCmdId(), len(frame), utils.TimeSince(now))
	}()

	moffset := 2 + headlen
	cmdId := head.GetCmdId()
	if cmdId != uint32(ss.EnmCmdValue_ECV_SSTransReq) {
		if head.RetCode == uint32(ss.EnmRetCode_NEED_REGIST) {
			logger.Log.Error("service not regist first. Head: %s", head.String())
			serv := relay.RelayMgr.GetRelay(connIdx)
			if nil == serv {
				logger.Log.Error("not found service. ConnIdx: %d", connIdx)
			} else {
				serv.RegistService(int32(ss.ServiceType_GameSvr), config.GetServiceId(), config.GetAreaId())
			}
			return
		}

		var (
			imessage utils.IMessage
			ok       bool
		)
		if cmdId == uint32(ss.EnmCmdValue_ECV_SSMiscReq) {
			imessage = &game_third.SSMiscReq{}
			ok = true
		} else {
			imessage, ok = ss.NewServiceMessage(int32(cmdId))
		}
		if !ok || imessage == nil {
			action = gnet.Close
			logger.Log.Error("%d|not found message. Head: %s", connIdx, head.String())
			return
		}

		if merr := proto.Unmarshal(frame[moffset:], imessage); merr != nil {
			action = gnet.Close
			logger.Log.Error("%d|unpack message failed. Head: %s, Err: %#v", connIdx, head.String(), merr)
			return
		}

		if utils.IsNetLog() && utils.LogFlag(head.CmdId) {
			logger.Log.Debug("%d|Recv Message(service). Head: %s, Message: %s", connIdx, head.String(), imessage.String())
		}

		handler := apps.GetServiceHandle(cmdId)

		if handler == nil {
			logger.Log.Error("%d|not found handler. Head: %s, Message: %s", connIdx, head.String(), imessage.String())
			return
		}

		handler(c, head, imessage)
	} else {
		apps.HandleTransReq(head, frame[moffset:])
	}

	return
}

func (this *MainServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	if this.isStop() {
		action = gnet.Close
		logger.Log.Error("[React] server stoping, cann't process anything")
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
			logger.Log.Error("panic: %s", osslog.OSS.PanicTrace(err))
		}
	}()

	// 初始化连接上下文，如果uri是/game，上下文则为connIdx
	context := c.Context()
	if context == nil {
		err, action := this.initialConnContext(c)
		if err != nil {
			return action
		}
	}

	connIdx, ok := c.Context().(uint64)
	if ok && connIdx < utils.MaxServiceConnIdx {
		for {
			// 服务器消息：解包
			frame, _ := this.codec.Decode(c)
			if frame == nil {
				return
			}

			if action = this.reactService(connIdx, frame, c); action != gnet.None {
				return
			}
		}
		return
	}

	// @NOTE 人为制造时间延时，通过 control latency 控制
	if this.latency > 0 {
		time.Sleep(time.Duration(this.latency) * time.Millisecond)
	}

	player := model.PlayerMgr.GetPlayerByConnIdx(connIdx)
	if player == nil {
		logger.Log.Error("not found player by connIdx: %d", connIdx)
		return
	}

	if !player.WSUpgrade {
		if utils.IsNetLog() {
			logger.Log.Error("start player upgrade success. ConnIdx: %d", connIdx)
		}
		_, werr := ws.Upgrade(c)
		if werr != nil {
			logger.Log.Error("ws upgrade failed. ConnIdx: %d, werr: %v", connIdx, werr)
			return
		}
		player.WSUpgrade = true
		if utils.IsNetLog() {
			logger.Log.Error("finish player upgrade success. ConnIdx: %d", connIdx)
		}
		return
	}

	for {
		ok, frame, opCode, _, werr := utils.WsServDecode(c)
		if werr != nil {
			logger.Log.Error("[WsServDecode] read data failed. ConnIdx: %d, werr: %s, OpCode: %d", connIdx, werr.Error(), opCode)
			osslog.OSS.ElkAlarm(oss.EnmElk_Network_Error, player.AccId, fmt.Sprintf("WsServDecode read data failed. ConnIdx: %d, werr: %s, OpCode: %d", connIdx, werr.Error(), opCode))
			action = gnet.Close
			return
		}
		// 不是一个完整包
		if !ok {
			break
		}

		switch opCode {
		case utils.OpPing:
			// 挂机状态没有命令字，AC没有心跳包，使用ws ping包维持心跳
			player.SetLastRecvMsg()
			pongPkg, err := utils.WsEncodePong()
			if err == nil {
				c.AsyncWrite(pongPkg, nil)
			}
			logger.Log.Debug("%s|recv ws ping", player)
			player.Heart()
			continue
		case utils.OpClose:
			if len(frame) > 4 {
				logger.Log.Error("[WsServDecode] read close data. ConnIdx: %d, frame: %v", connIdx, string(frame[2:]))
			}
			if !player.IsSendCloseFrame {
				player.IsSendCloseFrame = true
				closePkg, _ := utils.WsEncodeClose(0)
				c.AsyncWrite(closePkg, nil)
				action = gnet.None
			} else {
				action = gnet.Close
			}
			return
		}

		// 只处理 OpBinary 数据
		if opCode != utils.OpBinary {
			logger.Log.Error("[WsServDecode] read not binary data. ConnIdx: %d, OpCode: %d", connIdx, opCode)
			continue
		}

		// 数据包过小，认为是非法连接
		if len(frame) < LengthFieldLength {
			logger.Log.Error("[WsServDecode] read frame too small. ConnIdx: %d, frame.len:", connIdx, len(frame))
			action = gnet.Close
			return
		}

		// 客户端消息
		action = this.acReactClient(player, frame[LengthFieldLength:], c)
		if action != gnet.None {
			return
		}
	}

	return
}

func crossDay(last, now int64) {
	lastTm := time.Unix(last/1000, 0)
	nowTm := time.Unix(now/1000, 0)
	zeroTm := utils.GetZeroTimeStamp(nowTm.Unix(), 0)

	// 跨天0点
	if lastTm.Unix() < zeroTm && nowTm.Unix() >= zeroTm {
		// 跨天0点
		model.PlayerMgr.CrossZero()

		// 跨月
		if nowTm.Month() != lastTm.Month() {
			model.PlayerMgr.CrossMonth()
		}

		// 跨周
		if !utils.IsSameWeek(lastTm.Unix(), nowTm.Unix()) {
			model.PlayerMgr.CrossWeek()
		}
	} else if lastTm.Unix() < (zeroTm+utils.FOUR_CLOCK_OFFSET) && nowTm.Unix() >= (zeroTm+utils.FOUR_CLOCK_OFFSET) {
		// 跨天4点
		model.PlayerMgr.CrossFour()
	} else if lastTm.Unix() < (zeroTm+utils.FIVE_CLOCK_OFFSET) && nowTm.Unix() >= (zeroTm+utils.FIVE_CLOCK_OFFSET) {
		// 跨天5点
		model.PlayerMgr.CrossFive()
	}
}

func (this *MainServer) tick1s(now time.Time) {
	t := now.UnixNano() / 1e6
	// 检查跨天
	crossDay(this.tick1sTm, t)
	this.tick1sTm = t

	this.checkReady()

	if this.state == ServerState_PreStop {
		logger.Log.Error("server PreStop: all goroutine is done")
		this.state = ServerState_Stop
	}
}

func (this *MainServer) OnShutdown(_ gnet.Engine) {
}

func (this *MainServer) OnTick() (delay time.Duration, action gnet.Action) {
	// @NOTE: 不能在这里添加代码, 这里是子线程
	delay = time.Millisecond * 1000
	return
}

// 返回: errors.ErrServerShutdown 则关闭服务器
func (this *MainServer) PollerProc() error {
	now := time.Now()

	defer func() {
		err := recover()
		if err != nil {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
			logger.Log.Error("panic: %s", osslog.OSS.PanicTrace(err))
		}
	}()

	select {
	case sig := <-this.signalChan:
		logger.Log.Error("recv system signal: %s", sig.String())
		this.handleStopServiceEvent(nil)
	default:
	}

	utils.WorkerPool.Callback()
	utils.ResetTime()
	utils.ProcessTimer()

	// 服务器退出
	if this.state == ServerState_Stop {
		logger.Log.Error("server Stop ...")
		// action = gnet.Shutdown
		return gerrors.ErrEngineShutdown
	}
	utils.StatMgr.StatFuncCall("PollerProc", utils.TimeSince(now))

	return nil
}

func (this *MainServer) checkReady() {
	if this.state != ServerState_NoReady {
		return
	}

	if !relay.RelayMgr.IsReady() {
		logger.Log.Error("relay service is not ready ...")
		return
	}

	this.state = ServerState_Ready
	logger.Log.Error("gamesvr is ready ...")
}

func (this *MainServer) SetState(stat ServerState) {
	this.state = stat
}

func (this *MainServer) OnHealtyCheck() (httpCode int) {
	if this.isRollingUpdate() {
		// 健康检测健康->不健康阈值
		healthMaxNum := config.Conf.Common.HealthMaxNum
		// 健康检测间隔时间
		healthCheckInterval := config.Conf.Common.HealthCheckInterval
		now := time.Now().Unix()
		if this.unHealthStartTime == 0 {
			this.unHealthStartTime = now
		}

		// 不能判断检测次数，云厂商可能有多个节点并发检测
		if this.unHealthStartTime+healthCheckInterval*int64(healthMaxNum+1) < now {
			if !this.notifyRollingUpdate {
				logger.Log.Error("start rollingUpdate, notify online player reconnect.")
				model.PlayerMgr.ServerRollingUpdate()
				this.notifyRollingUpdate = true
			}
		}
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func StartServe() {
	utils.InitTime(false)

	// 云平台健康检测端口
	utils.WorkerPool.Do(0, "healthChecker", healthChecker)

	reuseport := true
	// multicore 为 false，否则有可能会造成主逻辑是多线程
	multicore := false
	async := true
	network := "tcp"
	addr := config.Conf.Network.ClientService

	encoderConfig := gnet.EncoderConfig{
		ByteOrder:         binary.BigEndian,
		LengthFieldLength: LengthFieldLength,
	}

	decoderConfig := gnet.DecoderConfig{
		ByteOrder:         binary.BigEndian,
		LengthFieldLength: LengthFieldLength,
	}

	codec := gnet.NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)
	ms = &MainServer{
		network:   network,
		addr:      addr,
		multicore: multicore,
		async:     async,
		// workerPool: goroutine.Default(),
		codec:      codec,
		signalChan: make(chan os.Signal, 1),
	}
	utils.SetCodec(codec)
	signal.Notify(ms.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	logger.Log.Error("start client service ...")

	gnet.Run(ms, network+"://"+addr,
		gnet.WithLockOSThread(async),
		gnet.WithMulticore(multicore),
		gnet.WithReusePort(reuseport),
		gnet.WithTicker(false),
		gnet.WithTCPKeepAlive(time.Minute*1),
		// gnet.WithCodec(codec),
		gnet.WithTCPNoDelay(gnet.TCPNoDelay),
		gnet.WithLoadBalancing(gnet.SourceAddrHash),
		// gnet.WithLogger(logger.Log),
	)
}

func healthChecker() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		server := GetMainServer()
		code := http.StatusOK
		if server != nil {
			code = server.OnHealtyCheck()
		}
		c.JSON(code, "{}")
	})

	r.Run(fmt.Sprintf(":%d", config.Conf.Common.HealthPort))
}

func (server *MainServer) handleGnetHealthCheck(c gnet.Conn) (action gnet.Action) {
	code := server.OnHealtyCheck()
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n", code, http.StatusText(code))
	c.Write([]byte(response))
	c.SetContext("healthCheck")
	return gnet.Close
}

func (server *MainServer) handleGnetMetrics(c gnet.Conn, request *http.Request) (action gnet.Action) {
	start := time.Now()
	defer func() {
		utils.StatMgr.StatFuncCall("PrometheusMetrics", utils.TimeSince(start))
	}()
	c.SetContext("healthCheck")

	if model.HttpHandler != nil {
		response := httptest.NewRecorder()
		model.HttpHandler.ServeHTTP(response, request)
		result := response.Result()
		result.Write(c)
	} else {
		logger.Log.Error("promhttp.HandlerForTransactional failed. http handler is nil")
	}

	return gnet.Close
}

func (server *MainServer) initialConnContext(c gnet.Conn) (error, gnet.Action) {
	// 读取tcp缓冲区数据
	httpFrame, err := c.Peek(-1)
	if err != nil {
		logger.Log.Error("peek gnet.Conn buffer failed. err: %s", err.Error())
		return err, gnet.Close
	}

	reader := bufio.NewReader(bytes.NewReader(httpFrame))
	request, err := http.ReadRequest(reader)
	if err != nil {
		logger.Log.Error("read http request failed. err: %s", err.Error())
		return err, gnet.Close
	}

	switch request.URL.Path {
	case "/heathz":
		return NotGameRequest, server.handleGnetHealthCheck(c)
	case "/metrics":
		return NotGameRequest, server.handleGnetMetrics(c, request)
	case "/game":
		cidx := snowflake.GenUniqId(snowflake.IDType_ConnIdx)
		c.SetWebSock(true)
		c.SetContext(cidx)
		realIp := request.Header.Get("X-Forwarded-For")
		player := model.PlayerMgr.Create(c, cidx)
		if player != nil && realIp != "" {
			// 头部可能包含多个 IP 地址，格式如 client, proxy1, proxy2。通常，第一个是客户端的真实 IP。
			realIps := strings.Split(realIp, ",")
			if len(realIps) > 0 {
				logger.Log.Error("%s|player create success. remote address: %s", player, realIps[0])
				player.IP = realIps[0]
			}
		}
		return nil, gnet.None
	default:
		logger.Log.Warn("unknown http request uri: %s", request.URL.Path)
		return UnknownRequestUri, gnet.Close
	}
}
