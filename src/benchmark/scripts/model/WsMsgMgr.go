package model

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"gamekit/pkg/cs"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"benchmark/scripts/config"

	"github.com/gorilla/websocket"
	"github.com/xxtea/xxtea-go/xxtea"
	"google.golang.org/protobuf/proto"
)

var (
	maxBufferSize int   = 512 * 1024
	overloadSize  error = fmt.Errorf("overload buffer size: %d", maxBufferSize)

	// 压测客户端消息序号(脚手架最小实现, 替代 StatCallMgr)
	seqIdCounter uint32
)

// GenSeqId 生成消息序号
func GenSeqId() uint32 {
	return atomic.AddUint32(&seqIdCounter, 1)
}

func PackMsg(head utils.IHeadMsg, message utils.IMessage) ([]byte, error) {
	var (
		headlen, bodylen int
		messageBytes     []byte
		headBytes        []byte
		err              error
		encryptBuffer    []byte
		bitMask          byte
		xxteaBuffer      []byte
		xxteaOpen        bool   = config.Conf.Common.XXTeaOpen
		xxteaKey         []byte = []byte(config.Conf.Common.XXTeaKey)
	)

	headBytes, err = proto.Marshal(head)
	if err != nil {
		logger.Log.Error("proto.Marshal failed. Head: %s, Err: %#v", head.String(), err)
		return nil, err
	}
	headlen = len(headBytes)

	messageBytes, err = proto.Marshal(message)
	if err != nil {
		logger.Log.Error("proto.Marshal failed. Head: %s, Message: %s, Err: %#v", head.String(), message.String(), err)
		return nil, err
	}

	if utils.IsNetLog() && utils.LogFlag(head.GetCmdId()) {
		logger.Log.Debug("SendMsg success. messageSize: %d, Head: {%s}, Message: %s", len(messageBytes), head.String(), message)
	}

	bodylen = len(messageBytes)
	// 生成加密缓存区
	pressBuffer := make([]byte, headlen+bodylen)
	copy(pressBuffer, headBytes)
	boffset := headlen
	copy(pressBuffer[boffset:], messageBytes)

	xxteaBuffer = pressBuffer

	if xxteaOpen {
		bitMask |= 1
		encryptBuffer = xxtea.Encrypt(xxteaBuffer, xxteaKey)
		if len(encryptBuffer)+2 >= maxBufferSize {
			logger.Log.Error("encrypt data overload max buffer size. Head: %s, TotalLen: %d", head.String(), len(pressBuffer))
			return nil, overloadSize
		}
	} else {
		encryptBuffer = xxteaBuffer
	}

	sendBuffer := make([]byte, 6+len(encryptBuffer))
	binary.LittleEndian.PutUint32(sendBuffer[0:4], uint32(len(encryptBuffer)+2))
	sendBuffer[4] = bitMask
	sendBuffer[5] = byte(headlen)
	copy(sendBuffer[6:], encryptBuffer)

	return sendBuffer, nil
}

func UnpackMsg(frame []byte) (*cs.CSHead, utils.IMessage, error) {
	var (
		xxteaKey []byte = []byte(config.Conf.Common.XXTeaKey)
		buffer   []byte
	)

	if len(frame) < 6 {
		return nil, nil, errors.New("frame length is too short")
	}

	// 加密 && 压缩
	bitmask := uint32(frame[4])
	// 头部长度
	headlen := uint32(frame[5])

	buffer = frame[6:]

	// 解密
	if (bitmask & 0xf) != 0 {
		buffer = xxtea.Decrypt(buffer, xxteaKey)
		if len(buffer) == 0 {
			logger.Log.Error("xxtea decrypt fialed")
			return nil, nil, errors.New("xxtea decrypt failed")
		}
	}

	if len(buffer) < int(headlen) {
		logger.Log.Error("buffer is too small. buf.len: %d, headlen: %d, frame: %#v", len(buffer), headlen, frame)
		return nil, nil, errors.New("buffer is too small")
	}

	head := &cs.CSHead{}
	err := proto.Unmarshal(buffer[:headlen], head)
	if err != nil {
		logger.Log.Error("unpack head failed. length: %d, headlen: %d, err: %s", len(buffer), headlen, err)
		return nil, nil, err
	}

	imessage, ok := cs.NewClientMessage(int32(head.GetCmdId()))
	if !ok || imessage == nil {
		return nil, nil, fmt.Errorf("cmdId not registed: %d", head.GetCmdId())
	}

	err = proto.Unmarshal(buffer[headlen:], imessage)
	if err != nil {
		logger.Log.Error("unpack message failed. Head: %s, Err: %s", head.String(), err)
		return nil, nil, err
	}

	return head, imessage, nil
}

func SendWsMsg(player *Player) {
	for {
		select {
		case <-player.RecvSendCtx.Done():
			logger.Log.Error("%s|send ws message exit", player)
			return
		case sendMsg := <-player.SendMsgs:
			if player.WsConn == nil {
				logger.Log.Error("player ws conn is nil. accId: %d", player.GetAccId())
				return
			}
			err := player.WsConn.WriteMessage(websocket.BinaryMessage, sendMsg)
			if err != nil {
				logger.Log.Error("send ws message failed. accId: %d, err: %s", player.GetAccId(), err.Error())
			}
		case <-player.PingMsg:
			if player.WsConn != nil {
				player.WsConn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}
}

func SendWsMsgChan(player *Player, head utils.IHeadMsg, message utils.IMessage) error {
	data, err := PackMsg(head, message)
	if err != nil {
		logger.Log.Error("pack send msg failed. err: %s head: %s", err.Error(), head.String())
		return err
	}

	now := time.Now()
	player.CmdStatMap[cs.EnmCmdValue(head.GetCmdId())] = &now
	player.SendMsgs <- data
	return nil
}

func RecvWsMsg(player *Player) {
	var (
		messageType int
		buffer      []byte
		err         error
	)

	for {
		select {
		case <-player.RecvSendCtx.Done():
			return
		default:
			messageType, buffer, err = player.WsConn.ReadMessage()
			if err != nil {
				return
			}

			switch messageType {
			case websocket.BinaryMessage:
				player.RecvMsgs <- buffer
			case websocket.PongMessage:
				logger.Log.Debug("%s|recv pong message", player)
			}
		}
	}
}

func DispatchMsg(player *Player, buffer []byte) {
	head, iMessage, err := UnpackMsg(buffer)
	if err != nil {
		logger.Log.Error("unpack message failed. err: %s", err.Error())
		return
	}

	defer func() {
		err := recover()
		if err != nil {
			logger.Log.Error("panic: %v", err)
		}
	}()

	handler := GetClientHandle(uint32(head.GetCmdId()))
	if handler == nil {
		logger.Log.Debug("not found handler. CmdId: %d", head.CmdId)
		return
	}
	if utils.IsNetLog() {
		logger.Log.Error("recv msg|AccId: %d, head: %s, message: %s", player.GetAccId(), head, iMessage)
	}

	reqCmd := cs.GetReqCmdId(int32(head.GetCmdId()))
	startTime, ok := player.CmdStatMap[cs.EnmCmdValue(reqCmd)]
	if ok {
		utils.StatMgr.StatCmdCall(uint32(reqCmd), len(buffer), utils.TimeSince(*startTime))
	} else {
		logger.Log.Debug("not found stat cmd. CmdId: %d", head.GetCmdId()-1)
	}
	handler(player, head, iMessage)
}
