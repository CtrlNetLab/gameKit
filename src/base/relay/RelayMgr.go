package relay

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"gamekit/pkg/oss"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/snowflake"
	"gamekit/src/base/utils"

	log "github.com/walkon/log4go"
	gnet "github.com/walkon/wsgnet"
	"google.golang.org/protobuf/proto"
)

var (
	RelayMgr *relayMgr
	Log      *log.Filter
)

var (
	NotFoundRelay error = fmt.Errorf("not found relay service")
)

func init() {
	RelayMgr = &relayMgr{}
}

type DelayBuffer struct {
	buffer    []byte
	timestamp int64

	dst   uint64
	accId uint64
	cmdId uint32
}

type relayMgr struct {
	relays []*Relay

	eng *gnet.Engine

	// 心跳包请求列表，heartSeqId --> *Relay
	hearts     map[uint32]*Relay
	heartSeqId uint32

	// 如果所有 relay 突然断掉，要发送的数据临时保存起来，等到连接上之后再发送
	delays []DelayBuffer

	serviceId   int32
	serviceType int32
	// 服务本身所在地区Id
	areaId int32

	// relay发送消息固定relay,保证消息发送顺序
	relayConnIdxMap map[uint64]uint64
}

func (this *relayMgr) GetSelfConnIdx() uint64 {
	return (((uint64(this.serviceType) & 0xff) << 16) | (uint64(this.serviceId) & 0xffff))
}

func (this *relayMgr) Init(serviceId, serviceType, areaId int32, eng *gnet.Engine, log *log.Filter) {
	Log = log

	InitRelayConfig()

	this.relays = make([]*Relay, 0)
	this.eng = eng
	this.serviceId = serviceId
	this.serviceType = serviceType
	this.areaId = areaId
	this.relayConnIdxMap = make(map[uint64]uint64)

	this.hearts = make(map[uint32]*Relay)
	this.heartSeqId = 10000

	var relayIdx uint64 // 这里只是防止向gnet插入连接时报错
	for areaStr, hosts := range RelayConf.RelaySvr {
		relayAreaId := ss.EnmLocation_value[fmt.Sprintf("EL_%s", areaStr)]
		// 需要跨地区通信的，才连接其他地区的relay
		if relayAreaId != 0 && relayAreaId != int32(areaId) {
			if !IsWorldService(serviceType) {
				logger.Log.Debug("not world service, serviceType: %d", serviceType)
				continue
			}
		}
		for _, host := range hosts {
			relayIdx++
			serv := &Relay{}
			serv.init(relayIdx, host)
			this.relays = append(this.relays, serv)
		}
		logger.Log.Debug("world service, serviceType: %d", serviceType)
	}

	utils.AddTimer("RelayMgr::tick200ms", time.Millisecond*200, this.tick200ms)
}

func (this *relayMgr) connect() {
	var (
		now        int64    = utils.GetTimeMS()
		reConnTm   int64    = RelayConf.Common.ReConnTm
		reConnNum  int32    = RelayConf.Common.ReConnNum
		reConnList []*Relay = make([]*Relay, 0)
	)

	if reConnNum < 0 {
		reConnNum = math.MaxInt32
	}

	for _, serv := range this.relays {
		if (serv.state == RelayState_Init || serv.state == RelayState_Failed) &&
			((serv.lastConnTm+reConnTm) < now && (serv.connNum+1) < reConnNum) {
			serv.setReConn(now)
			reConnList = append(reConnList, serv)
		}
	}

	if len(reConnList) <= 0 {
		return
	}

	Log.Error("have %d service need connect no delay", len(reConnList))

	task := func(serv *Relay) (int32, gnet.Conn) {
		var (
			state int32 = RelayState_Success
			conn  gnet.Conn
		)

		// err为nil表示打开socket成功, 不代表链接建立成功; 如果对端直接close Socket, 则链接建立失败
		pconn, cerr := gnet.AddTCPConnector(this.eng, "tcp", serv.host, serv.GetConnIdx(), gnet.WithTCPNoDelay(gnet.TCPNoDelay))
		if cerr != nil {
			state = RelayState_Failed
			Log.Error("add service to poller failed. Host: %s, Err: %#v", serv.host, cerr)
		} else {
			conn = gnet.Conn(pconn)
			Log.Error("connect %s success", serv.host)
		}
		return state, conn
	}

	connect := func(connIdx uint64, serv *Relay) {
		utils.WorkerPool.Do(connIdx, "RelayMgr:connect", func() {
			state, conn := task(serv)
			callback := func() {
				if serv.state == RelayState_Connecting {
					// 如果对端直接close Socket, SetState会覆盖Close的状态
					serv.SetState(state)
					if state == RelayState_Success {
						serv.SetConn(conn)
						serv.RegistService(this.serviceType, this.serviceId, this.areaId)
					}
				} else {
					serv.lastConnTm = now
					Log.Error("relay connect callback is not connecting. Serv: %s, State: %d", serv.String(), state)
				}
				Log.Error("connect callback. Serv: %s, State: %d", serv.String(), state)
			}
			utils.WorkerPool.AddCallback("RelayMgr:connect", callback)
		})
	}

	for idx, serv := range reConnList {
		connect(uint64(idx), serv)
	}
}

func (this *relayMgr) tick200ms(now time.Time) {
	this.connect()
	this.checkHeart()
	this.SendHeart()
}

func (this *relayMgr) checkHeart() {
	now := utils.GetTimeMS()
	heartTimeOut := RelayConf.Common.HeartTimeOut
	maxHeartNum := RelayConf.Common.MaxHeartNum

	var heartTimeOutList []uint32
	for seqId, serv := range this.hearts {
		if !serv.IsReady() || serv.lastHeartTm <= 0 {
			continue
		}

		// 心跳包超时
		if (serv.lastHeartTm + heartTimeOut) < now {
			serv.heartTimeOutNum++
			// 超时认为 RTT = heartTimeOut
			serv.addRTT(heartTimeOut)
			heartTimeOutList = append(heartTimeOutList, seqId)
			Log.Error("%s|heart timeout. HeartTimeOutNum: %d", serv.String(), serv.heartTimeOutNum)
			// 心跳包超时次数过多，需要重连
			if serv.heartTimeOutNum >= maxHeartNum {
				Log.Error("%s|service HeartTimeOutNum: %d, try reconnect", serv.String(), serv.heartTimeOutNum)
				serv.Close()
			}
		}
	}

	for _, seqId := range heartTimeOutList {
		delete(this.hearts, seqId)
	}
}

func (this *relayMgr) SendHeart() {
	now := utils.GetTimeMS()
	relayHeartTm := RelayConf.Common.RelayHeartTm

	head := utils.NewSSHead(ss.EnmCmdValue_ECV_SSHeartReq, 0, ss.EnmRetCode_SUCCESS)
	heartReq := &ss.SSHeartReq{}

	for _, serv := range this.relays {
		if !serv.IsReady() || serv.lastHeartSeqId > 0 {
			continue
		}

		if serv.lastHeartTm+relayHeartTm < now {
			head.SeqId = this.genHeartSeqId()
			serv.lastHeartTm = now
			serv.lastHeartSeqId = head.SeqId
			this.hearts[head.SeqId] = serv

			utils.SendMsg(serv.conn, head, heartReq)
		}
	}
}

func (this *relayMgr) genHeartSeqId() uint32 {
	this.heartSeqId++
	return this.heartSeqId
}

func (this *relayMgr) GetRelay(connIdx uint64) *Relay {
	for _, serv := range this.relays {
		if serv != nil && serv.GetConnIdx() == connIdx {
			return serv
		}
	}
	return nil
}

func (this *relayMgr) CloseRelay(connIdx uint64, err error) {
	serv := this.GetRelay(connIdx)
	if serv == nil {
		Log.Error("not found service, ConnIdx: %d", connIdx)
		return
	}

	Log.Error("%s|service close, Err: %#v", serv.String(), err)

	serv.Close()
}

func (this *relayMgr) GetAliveRelayArea(areaId int32) *Relay {
	relayList := make([]*Relay, 0)
	for _, serv := range this.relays {
		if !serv.IsReady() || (areaId != 0 && serv.areaId != areaId) {
			continue
		}

		relayList = append(relayList, serv)
		//	if retServ == nil {
		//		retServ = serv
		//	} else if retServ.rtt > serv.rtt {
		//		retServ = serv
		//	}
	}

	return utils.RandomSliceOne(relayList)
}

func (this *relayMgr) GetAliveRelayConnIdx(connIdx uint64, toArea int32) *Relay {
	var (
		retServ *Relay = nil
	)
	// 优先hash
	relayConnIdx, ok := this.relayConnIdxMap[connIdx]
	if !ok {
		relayList := make([]*Relay, 0)
		for _, serv := range this.relays {
			if !serv.IsReady() || (toArea != 0 && serv.areaId != toArea) {
				continue
			}

			relayList = append(relayList, serv)
		}
		retServ = utils.RandomSliceOne(relayList)
	} else {
		for _, serv := range this.relays {
			if !serv.IsReady() {
				continue
			}

			if serv.GetConnIdx() == relayConnIdx {
				return serv
			}
			if retServ == nil {
				retServ = serv
			} else if retServ.rtt > serv.rtt {
				retServ = serv
			}
		}
	}
	if retServ != nil {
		this.relayConnIdxMap[connIdx] = retServ.GetConnIdx()
	}
	return retServ
}

func (this *relayMgr) HeartRes(connIdx uint64, seqId uint32) {
	serv, sok := this.hearts[seqId]
	if !sok || connIdx != serv.GetConnIdx() {
		Log.Error("heart response, but not found service. Ok: %t, ConnIdx: %d, SeqId: %d", sok, connIdx, seqId)
		return
	}

	delete(this.hearts, seqId)
	now := utils.GetTimeMS()
	if serv.lastHeartTm <= 0 || serv.lastHeartTm > now {
		Log.Error("%s|service lastHeartTm invalid. LastHeartTm: %d, Now: %d", serv.String(), serv.lastHeartTm, now)
		return
	}

	serv.addRTT(now - serv.lastHeartTm)
}

func (this *relayMgr) RegistRes(conn gnet.Conn, connIdx uint64) {
	if len(this.delays) <= 0 {
		return
	}

	Log.Error("start handle relay delay message. Count: %d, ConnIdx: %d", len(this.delays), connIdx)

	var (
		now          int64 = utils.GetTimeMS()
		delayTimeOut int64 = RelayConf.Common.DelayTimeOut
	)

	for _, delay := range this.delays {
		var timeout bool = delay.timestamp+delayTimeOut < now
		Log.Error("delay message. ConnIdx: %d, Dst: %d, AccId: %d, CmdId: %d, TimeOut: %v", connIdx, delay.dst, delay.accId, delay.cmdId, timeout)
		if timeout {
			continue
		}

		if err := utils.ConnAsyncWrite(conn, delay.buffer); err != nil {
			Log.Error("delay message failed. ConnIdx: %d, Dst: %d, AccId: %d, CmdId: %d, Err: %#v", connIdx, delay.dst, delay.accId, delay.cmdId, err)
			continue
		}

		Log.Error("delay message success. ConnIdx: %d, Dst: %d, AccId: %d, CmdId: %d", connIdx, delay.dst, delay.accId, delay.cmdId)
	}
	this.delays = this.delays[:0]
}

func (this *relayMgr) GetAllRelay() []*Relay {
	return this.relays
}

func (this *relayMgr) RelayString() string {
	var result string
	for _, serv := range this.relays {
		if len(result) <= 0 {
			result = serv.String()
		} else {
			result = result + "\n" + serv.String()
		}
	}
	result = result + "\n" + fmt.Sprintf("len(hearts): %d", len(this.hearts))
	return result
}

func (this *relayMgr) IsReady() bool {
	for _, serv := range this.relays {
		if serv.IsReady() {
			return true
		}
	}
	return false
}

func HandleServHeartRes(conn gnet.Conn, head *ss.SSHead, iMsg interface{}) {
	connIdx, ok := conn.Context().(uint64)
	if ok {
		RelayMgr.HeartRes(connIdx, head.SeqId)
	}
}

func HandleServRegistRes(conn gnet.Conn, head *ss.SSHead, iMsg interface{}) {
	response, ok := iMsg.(*ss.SSRegistRes)
	if !ok {
		return
	}

	connIdx, ok := conn.Context().(uint64)
	if !ok {
		Log.Error("registService failed, not found connIdx. Context: %+v", conn.Context())
		return
	}

	RelayMgr.RegistRes(conn, connIdx)
	serv := RelayMgr.GetRelay(connIdx)

	if head.RetCode == uint32(ss.EnmRetCode_DUPLICATE_REGIST) {
		// 判断是否已经有了相同id的 gamesvr，如果是则直接 panic
		Log.Error("already have same service, please change service.yaml:ServiceId")
		osslog.OSS.ElkAlarm(oss.EnmElk_Network_Error, 0, "already have same service, please change service.yaml:ServiceId. serviceType: %d, serviceId:  %d", RelayMgr.serviceType, RelayMgr.serviceId)
		// panic("already have same service, please change service.yaml:ServiceId")
		return
	}

	if serv != nil {
		serv.areaId = response.RelayAreaId
		Log.Error("%s|registService success", serv.String())
	} else {
		Log.Error("registService success, not found service")
	}
}

func TransToClientMsg(src, dst uint64, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	playerAreaId := snowflake.ParseAreaId(accId)
	relay := RelayMgr.GetAliveRelayArea(playerAreaId)
	return TransMsgImpl(relay, src, dst, accId, subHead, subMsg, ss.EnmMsgFlag_EMF_ToClient)
}

func TransToClientMsgExt(src, dst, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	playerAreaId := snowflake.ParseAreaId(accId)
	// TODO: 这里参数用accId不妥，应该用数据包来源的relay connIdx
	relay := RelayMgr.GetAliveRelayConnIdx(accId, playerAreaId)
	return TransMsgImplExt(relay, src, dst, accId, accId, subHead, subMsg, ss.EnmMsgFlag_EMF_ToClient)
}

func TransToAllGameSvr(src uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	relay := RelayMgr.GetAliveRelayArea(RelayMgr.areaId)
	return TransMsgImpl(relay, src, 0, 0, subHead, subMsg, ss.EnmMsgFlag_EMF_ToAllGameSvr)
}

func TransFromClientMsg(src, dst uint64, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	playerAreaId := snowflake.ParseAreaId(accId)
	relay := RelayMgr.GetAliveRelayArea(playerAreaId)
	return TransMsgImpl(relay, src, dst, accId, subHead, subMsg, ss.EnmMsgFlag_EMF_FromClient)
}

func TransMatchSvr(src, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	relay := RelayMgr.GetAliveRelayArea(RelayMgr.areaId)
	return TransMsgImpl(relay, src, uint64(ss.ServiceType_MatchSvr), accId, subHead, subMsg, ss.EnmMsgFlag_EMF_Init)
}

func TransGlobalSvr(src, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	relay := RelayMgr.GetAliveRelayArea(RelayMgr.areaId)
	return TransMsgImpl(relay, src, uint64(ss.ServiceType_GlobalSvr), accId, subHead, subMsg, ss.EnmMsgFlag_EMF_Init)
}

func TransMiscSvr(src, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	playerAreaId := snowflake.ParseAreaId(accId)
	relay := RelayMgr.GetAliveRelayArea(playerAreaId)
	return TransMsgImpl(relay, src, uint64(ss.ServiceType_MiscSvr), accId, subHead, subMsg, ss.EnmMsgFlag_EMF_Init)
}

func TransMsg(src, dst uint64, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	playerAreaId := snowflake.ParseAreaId(accId)
	relay := RelayMgr.GetAliveRelayArea(playerAreaId)
	return TransMsgImpl(relay, src, dst, accId, subHead, subMsg, ss.EnmMsgFlag_EMF_Init)
}

// AccId无法解析Area时使用，需从业务需求选择调用
func TransAreaMsg(src, dst, accId uint64, relayAreadId int32, subHead utils.IHeadMsg, subMsg utils.IMessage) error {
	relay := RelayMgr.GetAliveRelayArea(relayAreadId)
	return TransMsgImpl(relay, src, dst, accId, subHead, subMsg, ss.EnmMsgFlag_EMF_Init)
}

/***
 * @brief 透传消息到 dst 服务
 */
func TransMsgImpl(relay *Relay, src, dst, accId uint64, subHead utils.IHeadMsg, subMsg utils.IMessage, flag ss.EnmMsgFlag) error {
	if src <= 0 {
		src = RelayMgr.GetSelfConnIdx()
	}
	cmdId := subHead.GetCmdId()
	tHead := &ss.SSHead{
		CmdId:    uint32(ss.EnmCmdValue_ECV_SSTransReq),
		SeqId:    0,
		Src:      src,
		Dst:      dst,
		External: 0,
		AccId:    accId,
		SubCmdId: cmdId,
		Flag:     uint32(flag),
		FromArea: RelayMgr.areaId,
	}

	// 子消息也是按统一方式进行封装 |length|bitmark|head-length|message-buffer|
	subHeadLen := proto.Size(subHead)
	subMsgLen := 4 + 2 + subHeadLen + proto.Size(subMsg)

	// 外层包装协议
	tHeadLen := proto.Size(tHead)
	totallen := 2 + tHeadLen + subMsgLen

	buffer := make([]byte, totallen, totallen)
	buffer[0] = byte(0)
	buffer[1] = byte(tHeadLen)

	headOffset := 2
	headBytes, err := proto.Marshal(tHead)
	if err != nil {
		Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", tHead.String(), err)
		return err
	}
	copy(buffer[headOffset:], headBytes)

	subLenOffset := headOffset + tHeadLen
	binary.BigEndian.PutUint32(buffer[subLenOffset:], uint32(subMsgLen))

	subBitMarkOffset := subLenOffset + 4
	buffer[subBitMarkOffset] = byte(0)

	subHeadLenOffset := subBitMarkOffset + 1
	buffer[subHeadLenOffset] = byte(subHeadLen)

	subHeadOffset := subHeadLenOffset + 1
	subHeadBytes, err := proto.Marshal(subHead)
	if err != nil {
		Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", subHead.String(), err)
		return err
	}
	copy(buffer[subHeadOffset:], subHeadBytes)

	subMsgOffset := subHeadOffset + subHeadLen
	subMsgBytes, err := proto.Marshal(subMsg)
	if err != nil {
		Log.Error("pack proto.Marshal failed. Head: %s, Message: %s, Err: %#v", subHead.String(), subMsg.String(), err)
		return err
	}
	copy(buffer[subMsgOffset:], subMsgBytes)

	if relay == nil {
		Log.Error("not found serv service. Head: %s, Msg: %s", subHead.String(), subMsg.String())
		RelayMgr.delays = append(RelayMgr.delays,
			DelayBuffer{
				buffer:    buffer,
				timestamp: utils.GetTimeMS(),
				dst:       dst,
				accId:     accId,
				cmdId:     cmdId})
		return NotFoundRelay
	}

	if err := utils.ConnAsyncWrite(relay.conn, buffer); err != nil {
		Log.Error("%s|serv conn.Write failed. subHead: %s, subMsg: %s, Err: %v, len(buffer): %d",
			relay, subHead.String(), subMsg.String(), err, len(buffer))
		return err
	}

	if utils.IsNetLog() && utils.LogFlag(cmdId) {
		Log.Debug("%s|transfer message success. Dst: %d, AccId: %d, Head: %s, Msg: %s",
			relay.host, dst, accId, subHead.String(), subMsg.String())
	}

	return nil
}

func TransMsgImplExt(serv *Relay, src, dst uint64, accId uint64, connIdx uint64, subHead utils.IHeadMsg, subMsg utils.IMessage, flag ss.EnmMsgFlag) error {
	if src <= 0 {
		src = RelayMgr.GetSelfConnIdx()
	}
	cmdId := subHead.GetCmdId()
	tHead := &ss.SSHead{
		CmdId:    uint32(ss.EnmCmdValue_ECV_SSTransReq),
		SeqId:    0,
		Src:      src,
		Dst:      dst,
		External: 0,
		AccId:    accId,
		SubCmdId: cmdId,
		Flag:     uint32(flag),
		FromArea: RelayMgr.areaId,
	}

	// 子消息也是按统一方式进行封装 |length|bitmark|head-length|message-buffer|
	subHeadLen := proto.Size(subHead)
	subMsgLen := 4 + 2 + subHeadLen + proto.Size(subMsg)

	// 外层包装协议
	tHeadLen := proto.Size(tHead)
	totallen := 2 + tHeadLen + subMsgLen

	buffer := make([]byte, totallen, totallen)
	buffer[0] = byte(0)
	buffer[1] = byte(tHeadLen)

	headOffset := 2
	headBytes, err := proto.Marshal(tHead)
	if err != nil {
		Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", tHead.String(), err)
		return err
	}
	copy(buffer[headOffset:], headBytes)

	subLenOffset := headOffset + tHeadLen
	binary.BigEndian.PutUint32(buffer[subLenOffset:], uint32(subMsgLen))

	subBitMarkOffset := subLenOffset + 4
	buffer[subBitMarkOffset] = byte(0)

	subHeadLenOffset := subBitMarkOffset + 1
	buffer[subHeadLenOffset] = byte(subHeadLen)

	subHeadOffset := subHeadLenOffset + 1
	subHeadBytes, err := proto.Marshal(subHead)
	if err != nil {
		Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", subHead.String(), err)
		return err
	}
	copy(buffer[subHeadOffset:], subHeadBytes)

	subMsgOffset := subHeadOffset + subHeadLen
	subMsgBytes, err := proto.Marshal(subMsg)
	if err != nil {
		Log.Error("pack proto.Marshal failed. Head: %s, Message: %s, Err: %#v", subHead.String(), subMsg.String(), err)
		return err
	}
	copy(buffer[subMsgOffset:], subMsgBytes)

	if serv == nil {
		Log.Error("not found serv service. Head: %s, Msg: %s", subHead.String(), subMsg.String())
		RelayMgr.delays = append(RelayMgr.delays,
			DelayBuffer{
				buffer:    buffer,
				timestamp: utils.GetTimeMS(),
				dst:       dst,
				accId:     accId,
				cmdId:     cmdId})
		return NotFoundRelay
	}

	if err := utils.ConnAsyncWrite(serv.conn, buffer); err != nil {
		Log.Error("%s|serv conn.Write failed. subHead: %s, subMsg: %s, Err: %v, len(buffer): %d",
			serv.String(), subHead.String(), subMsg.String(), err, len(buffer))
		return err
	}

	if utils.IsNetLog() && utils.LogFlag(cmdId) {
		Log.Debug("%s|transfer message success. Dst: %d, AccId: %d, Head: %s, Msg: %s",
			serv.host, dst, accId, subHead.String(), subMsg.String())
	}

	return nil
}

func IsWorldService(serviceType int32) bool {
	return serviceType == int32(ss.ServiceType_GlobalSvr) || serviceType == int32(ss.ServiceType_GlobalRoomSvr) || serviceType == int32(ss.ServiceType_MatchSvr)
}
