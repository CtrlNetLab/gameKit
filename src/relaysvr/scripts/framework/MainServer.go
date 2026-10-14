package framework

import (
	"gamekit/pkg/oss"
	"gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/utils"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"relaysvr/scripts/apps"
	"relaysvr/scripts/config"
	"relaysvr/scripts/service"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	gnet "github.com/walkon/wsgnet"
	gerrors "github.com/walkon/wsgnet/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type ServerState int

const (
	ServerState_Init ServerState = 0
	// 正常运行阶段
	ServerState_Normal ServerState = 1
	// 准备退出阶段，需要做停服退出准备，不再处理客户端的请求
	ServerState_PreStop ServerState = 2
	// 退出
	ServerState_Stop ServerState = 3
)

type MainServer struct {
	gnet.BuiltinEventEngine
	eng          gnet.Engine
	network      string
	addr         string
	multicore    bool
	async        bool
	connected    int32
	disconnected int32

	// workerPool   *goroutine.Pool
	codec gnet.ICodec

	state ServerState

	// 人为制造时间延迟，单位: 毫秒
	latency int
}

func (this *MainServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	this.eng = eng
	this.state = ServerState_Normal

	service.ServiceMgr.Init()
	rand.Seed(time.Now().UnixNano())
	utils.AddEventHandler(ss.EventType_EId_StopService, this.handleStopServiceEvent)
	utils.AddEventHandler(ss.EventType_EId_Latency, this.latencyEvent)
	utils.AddTimer("MainServer::tick1s", time.Second, this.tick1s)

	return
}

func (this *MainServer) handleStopServiceEvent(event *utils.Event) {
	this.state = ServerState_PreStop
	logger.Log.Error("service stoping ...")
}

func (this *MainServer) latencyEvent(event *utils.Event) {
	if event == nil || len(event.Param) != 1 {
		return
	}
	this.latency = int(event.Param[0])
}

func (this *MainServer) PollerPreInit() {
}

func (this *MainServer) isStop() bool {
	return (this.state == ServerState_PreStop || this.state == ServerState_Stop)
}

// epoll_wait timeout, 默认值 10ms
func (this *MainServer) PollerWaitTimeOut() int {
	return 10
}

func (this *MainServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	if this.isStop() {
		action = gnet.Close
		logger.Log.Error("[OnOpened] server stoping, cann't process anything, state: %d", this.state)
		return
	}

	atomic.AddInt32(&this.connected, 1)

	return
}

func (this *MainServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if this.isStop() {
		logger.Log.Error("[OnClosed] server stoping, cann't process anything")
		return
	}

	var (
		connIdx uint64
		stype   int32
		id      int32
		ok      bool
	)

	ctx := c.Context()
	if ctx == nil {
		return gnet.Close
	}

	connIdx, ok = ctx.(uint64)
	if ok && connIdx > 0 {
		if connIdx <= utils.MaxServiceConnIdx {
			stype, id = utils.ServiceTypeIdx(connIdx)
			service.ServiceMgr.DelService(connIdx)
		}
	}

	if err != nil {
		logger.Log.Error("OnClosed, Err: %s, Type: %d, Id: %d, connIdx: %d", err.Error(), stype, id, connIdx)
	} else {
		logger.Log.Error("OnClosed, Type: %d, Id: %d, connIdx: %d", stype, id, connIdx)
	}
	atomic.AddInt32(&this.disconnected, 1)

	return
}

func (this *MainServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	if this.isStop() {
		action = gnet.Close
		logger.Log.Error("[React] server stoping, can't process anything, state: %d", this.state)
		return
	}

	var (
		connIdx uint64
		stype   int32
		id      int32
		ok      bool
	)

	// @NOTE 人为制造时间延时，通过 control latency 控制
	if this.latency > 0 {
		time.Sleep(time.Duration(this.latency) * time.Millisecond)
	}

	if c.Context() != nil {
		connIdx, ok = c.Context().(uint64)
		if ok {
			stype, id = utils.ServiceTypeIdx(connIdx)
		}
	}

	for {
		// 解包
		frame, _ := this.codec.Decode(c)
		if frame == nil {
			return
		}

		var length uint32 = uint32(len(frame))
		if length <= 2 || length < (2+uint32(frame[1])) {
			action = gnet.Close
			logger.Log.Error("length too small. Type: %d, Id: %d, length: %d", stype, id, length)
			return
		}

		// 加密 && 压缩
		// bit := uint32(frame[0])
		// 头部长度
		headlen := uint32(frame[1])
		head := &ss.SSHead{}

		if err := proto.Unmarshal(frame[2:2+headlen], head); err != nil {
			action = gnet.Close
			logger.Log.Error("unpack head failed. Type: %d, Id: %d, length: %d, headlen: %d", stype, id, length, headlen)
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

		if head.GetCmdId() != uint32(ss.EnmCmdValue_ECV_SSTransReq) {
			imessage, ok := ss.NewServiceMessage(int32(head.GetCmdId()))
			if !ok || imessage == nil {
				action = gnet.Close
				logger.Log.Error("not found message. Type: %d, Id: %d, CmdId: %d", stype, id, head.GetCmdId())
				return
			}

			if merr := proto.Unmarshal(frame[moffset:], imessage); merr != nil {
				action = gnet.Close
				logger.Log.Error("%s|unpack message failed. Type: %d, Id: %d, Head: %s, Err: %#v", stype, id, head.String(), merr)
				return
			}

			if utils.IsNetLog() && utils.LogFlag(head.GetCmdId()) {
				logger.Log.Debug("Recv Message. Type: %d, Id: %d, Head: %s, Message: %s", stype, id, head.String(), imessage.String())
			}
			handler := apps.GetServiceHandle(head.GetCmdId())

			if handler == nil {
				logger.Log.Error("not found handler. Type: %d, Id: %d, Head: %s, Message: %s", stype, id, head.String(), imessage.String())
				return
			}

			if serv := service.ServiceMgr.GetService(connIdx); serv != nil {
				serv.SetLastRecvMsgTm(utils.GetTimeMS())
			} else if head.CmdId != uint32(ss.EnmCmdValue_ECV_SSRegistReq) {
				// 还未注册，需要先注册
				resCmdId := ss.GetResCmdId(int32(head.CmdId))

				resMsg := &ss.SSRegistRes{}
				headRes := &ss.SSHead{CmdId: uint32(resCmdId), SeqId: head.SeqId, RetCode: uint32(ss.EnmRetCode_NEED_REGIST)}
				utils.SendMsg(c, headRes, resMsg)

				logger.Log.Error("service mustbe regist first. Head: %s, Msg: %s", head.String(), imessage.String())

				return
			}

			handler(c, head, imessage)
		} else {
			// 处理转发协议
			apps.HandleTransReq(head, frame[moffset:])
		}
	}
}

func (this *MainServer) tick1s(now time.Time) {
	if this.state == ServerState_PreStop {
		logger.Log.Error("server PreStop: all goroutine is done")
		this.state = ServerState_Stop
	}
}

func (this *MainServer) Tick() (delay time.Duration, action gnet.Action) {
	// @NOTE: 不能在这里添加代码, 这里是子线程
	delay = time.Millisecond * 1000
	return
}

// 返回: errors.ErrServerShutdown 则关闭服务器
func (this *MainServer) PollerProc() error {
	now := time.Now()

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

func StartServe() {
	go healthChecker()

	utils.InitTime(false)

	reuseport := true
	// multicore 为 false，否则有可能会造成主逻辑是多线程
	multicore := false
	async := true
	network := "tcp"
	addr := config.Conf.Network.ServerService

	encoderConfig := gnet.EncoderConfig{
		ByteOrder:         binary.BigEndian,
		LengthFieldLength: 4,
	}

	decoderConfig := gnet.DecoderConfig{
		ByteOrder:         binary.BigEndian,
		LengthFieldLength: 4,
	}

	codec := gnet.NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)

	ms := &MainServer{
		network:   network,
		addr:      addr,
		multicore: multicore,
		async:     async,
		// workerPool: goroutine.Default(),
		codec: codec,
	}

	utils.SetCodec(codec)

	logger.Log.Error("start server service ...")

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

	r.POST("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "{}")
	})

	r.Run(fmt.Sprintf(":%d", config.Conf.Common.HealthPort))
}
