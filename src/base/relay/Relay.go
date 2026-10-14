package relay

import (
	"fmt"
	"math"

	ss "gamekit/pkg/ss"
	"gamekit/src/base/utils"

	gnet "github.com/walkon/wsgnet"
)

const (
	RelayState_Init       int32 = 0
	RelayState_Connecting int32 = 1
	RelayState_Success    int32 = 2
	RelayState_Failed     int32 = 3
)

type Relay struct {
	connIdx uint64
	conn    gnet.Conn

	host string

	// 最后一次心跳包时间戳, 单位: 毫秒
	lastHeartTm int64
	// lastHeartSeqId > 0 说明心跳包还没回复
	lastHeartSeqId uint32
	// 心跳包连续超时次数
	heartTimeOutNum int32

	// 最后三次心跳包 RTT 时间均值, 单位：毫秒
	rtt int64
	// 保存心跳包最后三次 RTT 时间，单位：毫秒
	rttList []int64

	state int32

	// 最后一次重连时间，单位: 毫秒
	lastConnTm int64
	// 重连次数
	connNum int32
	// relay所在的地区
	areaId int32
}

func (relay *Relay) reset() {
	relay.lastHeartTm = 0
	relay.lastHeartSeqId = 0
	relay.rtt = math.MaxInt64
	relay.rttList = make([]int64, 3)
	relay.lastConnTm = 0
}

func (this *Relay) init(connIdx uint64, host string) {
	this.connIdx = connIdx
	this.host = host
	this.state = RelayState_Init
	this.reset()
}

func (this *Relay) setReConn(now int64) {
	this.SetState(RelayState_Connecting)
	this.connNum++
	this.reset()
	this.lastConnTm = now
}

func (this *Relay) SetState(state int32) {
	Log.Error("%s|state: %d -> %d", this.String(), this.state, state)
	this.state = state
}

func (this *Relay) SetConn(conn gnet.Conn) {
	this.conn = conn
	this.connNum = 0
}

func (this *Relay) IsReady() bool {
	return this.state == RelayState_Success
}

func (this *Relay) Close() {
	if this.conn != nil {
		this.conn.Close()
	}
	this.lastHeartTm = 0
	this.heartTimeOutNum = 0
	this.rtt = math.MaxInt64
	this.rttList = make([]int64, 3)
	this.SetState(RelayState_Failed)
	this.lastConnTm = 0
	this.connNum = 0
}

func (this *Relay) GetConnIdx() uint64 {
	return (((uint64(ss.ServiceType_RelaySvr) & 0xff) << 16) | (uint64(this.connIdx) & 0xffff))
}

func (this *Relay) RegistService(stype, serviceId, areaId int32) {
	registReq := &ss.SSRegistReq{
		Type:   stype,
		Id:     serviceId,
		AreaId: areaId,
	}
	head := utils.NewSSHead(ss.EnmCmdValue_ECV_SSRegistReq, 0, ss.EnmRetCode_SUCCESS)
	utils.SendMsg(this.conn, head, registReq)

	Log.Error("%s|regist service", this.String())
}

func (this *Relay) addRTT(rtt int64) {
	this.lastHeartSeqId = 0
	if len(this.rttList) >= 3 {
		this.rttList = this.rttList[1:]
	}
	this.rttList = append(this.rttList, rtt)

	this.rtt = 0
	for _, r := range this.rttList {
		this.rtt += r
	}
	this.rtt = int64(this.rtt / 3)
}

func (this *Relay) GetConn() gnet.Conn {
	return this.conn
}

func (this *Relay) String() string {
	return fmt.Sprintf("ConnIdx: %d, host: %s, lastHeartTm: %d, heartTimeOutNum: %d, rtt: %d, state: %d, lastConnTm: %d, connNum: %d, areaId: %d, lastHeartSeqId: %d",
		this.connIdx, this.host, this.lastHeartTm, this.heartTimeOutNum, this.rtt, this.state, this.lastConnTm, this.connNum, this.areaId, this.lastHeartSeqId)
}
