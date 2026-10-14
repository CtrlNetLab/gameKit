package model

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	cs "gamekit/pkg/cs"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"gamesvr/scripts/config"

	gnet "github.com/walkon/wsgnet"
	"github.com/pierrec/lz4/v4"
	"github.com/xxtea/xxtea-go/xxtea"
	"google.golang.org/protobuf/proto"
)

var (
	errNilConn = errors.New("player conn is nil")
	// Err_SendMsgOffline 玩家不在线, 无法发送消息
	Err_SendMsgOffline = errors.New("player is offline")
)

// Player 玩家对象(脚手架最小实现)
// 业务开发时在此结构体上扩展自己的玩家数据字段与方法
type Player struct {
	Primary   uint64
	AccId     uint64
	UserId    string
	LoginType int32
	Sid       string

	// 保存玩家连接
	conn gnet.Conn

	connIdx uint64

	// 登陆状态
	loginState LoginState

	DeviceId string

	// 客户端真实 IP
	IP string
	// 是否已完成 websocket 升级
	WSUpgrade bool
	// 是否已向客户端发送关闭帧
	IsSendCloseFrame bool

	// 最后一次收到消息的时间戳，单位: 毫秒
	lastRecvMsgTm int64

	// 通知序号, 服务端主动下发的消息自增
	NotifySeqId uint32

	// 是否控制actor(GM 代打), 控制actor 不向客户端发送消息
	IsControlActor bool
}

func (player *Player) String() string {
	if player == nil {
		return "<nil player>"
	}
	return fmt.Sprintf("accId:%d|connIdx:%d|sid:%s", player.AccId, player.connIdx, player.Sid)
}

func (player *Player) GetConnIdx() uint64 {
	return player.connIdx
}

func (player *Player) GetAccId() uint64 {
	return player.AccId
}

func (player *Player) IsOnlineOK() bool {
	return player.loginState == LoginState_Success
}

func (player *Player) SetLastRecvMsg() {
	// 毫秒
	player.lastRecvMsgTm = time.Now().UnixMilli()
}

func (player *Player) GetLastRecvMsg() int64 {
	return player.lastRecvMsgTm
}

// SetLoginState 设置登陆状态(业务登陆流程完成后调用)
func (player *Player) SetLoginState(state LoginState) {
	player.loginState = state
}

// Heart 心跳处理, 执行业务侧的心跳逻辑(如邮件过期检查)
func (player *Player) Heart() {
	if player == nil || !player.IsOnlineOK() {
		return
	}
}

// Disconnect 玩家连接断开
func (player *Player) Disconnect(reason LogoutReason) {
	if player == nil {
		return
	}
	if !player.IsOnlineOK() {
		logger.Log.Error("%s|player disconnect, but player is not ok. reason: %d", player, reason)
		return
	}

	PlayerMgr.Logout(player, reason)
}

// SendRoomMsg 透传消息给 roomsvr(脚手架未接入 roomsvr 时仅打印日志)
func (player *Player) SendRoomMsg(roomId uint64, head *cs.CSHead, msg utils.IMessage) error {
	logger.Log.Debug("%s|forward msg to room. roomId: %d, cmdId: %d", player, roomId, head.GetCmdId())
	return nil
}

// SendClientMsg 向客户端发送带返回码的响应消息
func (player *Player) SendClientMsg(cmdId cs.EnmCmdValue, seqId uint32, retCode cs.EnmRetCode, retMessage string, message utils.IMessage) error {
	if !player.IsOnlineOK() || player.connIdx <= 0 {
		return Err_SendMsgOffline
	}

	return SendClientMsg(player, cmdId, seqId, retCode, retMessage, message)
}

// SendClientNotify 向客户端发送通知消息(无请求响应语义)
func (player *Player) SendClientNotify(cmdId cs.EnmCmdValue, message utils.IMessage) error {
	if !player.IsOnlineOK() || player.connIdx <= 0 {
		return Err_SendMsgOffline
	}
	return SendClientMsg(player, cmdId, 0, cs.EnmRetCode_SUCCESS, "", message)
}

// SendClientMsg 打包 CSHead + 消息体并发送给客户端
func SendClientMsg(player *Player, cmdId cs.EnmCmdValue, seqId uint32, retCode cs.EnmRetCode, retMessage string, message utils.IMessage) error {
	head := utils.GetCSHead(cmdId, seqId, retCode, retMessage)
	if seqId == 0 {
		player.NotifySeqId++
		head.NotifySeqId = player.NotifySeqId
	}
	if retCode != cs.EnmRetCode_SUCCESS {
		logger.Log.Info("%s|send client err message. CmdId: %d, SeqId: %d, RetCode: %d, Message: %s", player, head.CmdId, head.SeqId, head.RetCode, message.String())
	}
	return sendClientMsg(player, head, message)
}

func sendClientMsg(player *Player, head *cs.CSHead, message utils.IMessage) error {
	if player.conn == nil {
		return errNilConn
	}

	connIdx, ok := player.conn.Context().(uint64)
	if !ok {
		return errNilConn
	}

	var marshalBuffer []byte
	marshalOpts := proto.MarshalOptions{Deterministic: true, UseCachedSize: true}
	marshalBuffer, err := marshalOpts.MarshalAppend(marshalBuffer, head)
	if err != nil {
		logger.Log.Error("%d|proto.Marshal head failed. Head: %s, Err: %s", connIdx, head, err.Error())
		return err
	}

	// NOTE: proto.Size(message) 防止反射修改字段值导致序列化失败，勿删
	_, err = marshalOpts.MarshalAppend(marshalBuffer, message)
	if err != nil {
		logger.Log.Error("%d|proto.Marshal body failed. Head: %s, Message: %s, Err: %s", connIdx, head, message, err.Error())
		return err
	}

	if utils.IsNetLog() && utils.LogFlag(head.GetCmdId()) {
		logger.Log.Debug("%d|SendMsg success. Head: {%s}, Message: %s", connIdx, head.String(), message.String())
	}

	// 压缩发送数据, 并加密
	sendBuffer, err := PressAndEncrypt(player, marshalBuffer, proto.Size(head))
	if err != nil {
		logger.Log.Error("%d|pack send buffer failed. Head: %s, Message: %s, Err: %#v", connIdx, head.String(), message.String(), err)
		return err
	}

	utils.StatMgr.StatCmdCall(head.CmdId, len(sendBuffer), time.Duration(0))
	if err := utils.ConnAsyncWrite(player.conn, sendBuffer); err != nil {
		logger.Log.Error("%d|conn.Write failed. Head: %s, Message: %s, Err: %#v, len(buffer): %d",
			connIdx, head.String(), message.String(), err, len(sendBuffer))
		return err
	}

	return nil
}

// PressAndEncrypt 压缩(lz4) + 加密(xxtea), 与客户端解包协议保持一致
func PressAndEncrypt(player *Player, pressBuffer []byte, headLen int) ([]byte, error) {
	var (
		xxteaOpen     bool   = config.Conf.Common.XXTeaOpen
		xxteaKey      []byte = []byte(config.Conf.Common.XXTeaKey)
		pressOpen     bool   = config.Conf.Common.PressOpen
		bitMask       byte
		xxteaBuffer   []byte
		encryptBuffer []byte
	)

	if pressOpen {
		// 只加密: 1; 只压缩: 1 << 4; 加密+压缩: (1 | (1 << 4))
		bitMask |= 1 << 4
		lz4Buffer := make([]byte, lz4.CompressBlockBound(len(pressBuffer))+4)
		var c lz4.Compressor
		lz4Len, err := c.CompressBlock(pressBuffer, lz4Buffer[4:])
		if err != nil {
			logger.Log.Error("lz4 press encode failed. err: %s", err.Error())
			return nil, err
		}
		// 压缩后的长度放入头部4字节， 客户端需要长度信息解压缩
		binary.LittleEndian.PutUint32(lz4Buffer[0:4], uint32(len(pressBuffer)))
		xxteaBuffer = lz4Buffer[:lz4Len+4]
	} else {
		xxteaBuffer = pressBuffer
	}

	if xxteaOpen {
		bitMask |= 1
		encryptBuffer = xxtea.Encrypt(xxteaBuffer, xxteaKey)
	} else {
		encryptBuffer = xxteaBuffer
	}

	sendBuffer := make([]byte, 2+len(encryptBuffer))
	sendBuffer[0] = bitMask
	sendBuffer[1] = byte(headLen)
	copy(sendBuffer[2:], encryptBuffer)

	return sendBuffer, nil
}

const (
	// MaxWsMessageSize websocket 单包最大字节数
	MaxWsMessageSize = 1024 * 1024
	// OverLoadAlarmRate 发送包大小超限告警比例
	OverLoadAlarmRate = 0.9
)

// GetConn 获取玩家连接(透传发送时使用)
func (player *Player) GetConn() gnet.Conn {
	return player.conn
}
