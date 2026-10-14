package apps

import (
	"math"

	cs "gamekit/pkg/cs"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	relay "gamekit/src/base/relay"

	"google.golang.org/protobuf/proto"
)

func init() {
	// 测试代码
	RegistTransmitHandle(uint32(ss.EnmCmdValue_ECV_SSTransTestReq), handleTransTestReq)
	RegistTransmitHandle(uint32(ss.EnmCmdValue_ECV_SSTransTestRes), handleTransTestRes)
}

// HandleTransReq 处理 relaysvr 转发过来的透传协议
// 透传协议格式: [4字节子协议掩码][1字节子头长度][子头][子消息体]
func HandleTransReq(head *ss.SSHead, buffer []byte) {
	subBitMarkOffset := 4
	subHeadLenOffset := subBitMarkOffset + 1
	subHeadOffset := subHeadLenOffset + 1
	subHeadLen := int(buffer[subHeadLenOffset])

	handler := GetTransmitHandle(head.SubCmdId)
	if handler == nil {
		logger.Log.Error("not found trans cmd handler. Head: %s", head.String())
		return
	}

	if head.SubCmdId < uint32(ss.EnmCmdValue_ECV_SSRegistReq) {
		// 客户端消息转发
		subHead := &cs.CSHead{}
		subMsg, ok := cs.NewClientMessage(int32(head.SubCmdId))
		if !ok || subMsg == nil {
			logger.Log.Error("unpack trans not found message. CmdId: %d", head.SubCmdId)
			return
		}

		if err := proto.Unmarshal(buffer[subHeadOffset:subHeadOffset+subHeadLen], subHead); err != nil {
			logger.Log.Error("unpack trans subHead failed. Err: %v, CmdId: %d", err, head.SubCmdId)
			return
		}

		subMsgOffset := subHeadOffset + subHeadLen
		if err := proto.Unmarshal(buffer[subMsgOffset:], subMsg); err != nil {
			logger.Log.Error("unpack trans subMsg failed. Err: %v, CmdId: %d", err, head.SubCmdId)
			return
		}

		handler(head, subHead, subMsg)
	} else {
		// 服务器消息转发
		subHead := &ss.SSHead{}
		subMsg, ok := ss.NewServiceMessage(int32(head.SubCmdId))
		if !ok || subMsg == nil {
			logger.Log.Error("unpack trans not found message. CmdId: %d", head.SubCmdId)
			return
		}

		if err := proto.Unmarshal(buffer[subHeadOffset:subHeadOffset+subHeadLen], subHead); err != nil {
			logger.Log.Error("unpack trans subHead failed. Err: %v, CmdId: %d", err, head.SubCmdId)
			return
		}

		subMsgOffset := subHeadOffset + subHeadLen
		if err := proto.Unmarshal(buffer[subMsgOffset:], subMsg); err != nil {
			logger.Log.Error("unpack trans subMsg failed. Err: %v, CmdId: %d", err, head.SubCmdId)
			return
		}

		handler(head, subHead, subMsg)
	}
}

func handleTransTestReq(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subMsg := iSubMsg.(*ss.SSTransTestReq)
	logger.Log.Debug("call handleTransTestReq. Head: %s, Value: %d", head.String(), subMsg.Value)
	testRes := &ss.SSTransTestReq{Value: subMsg.Value}
	ssHead := &ss.SSHead{CmdId: uint32(ss.EnmCmdValue_ECV_SSTransTestRes), SeqId: uint32(math.MaxInt32)}
	relay.TransMsg(0, head.Src, 0, ssHead, testRes)
}

func handleTransTestRes(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subMsg := iSubMsg.(*ss.SSTransTestRes)
	logger.Log.Debug("call handleTransTestRes. Head: %s, Value: %d", head.String(), subMsg.Value)
}
