package apps

import (
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	relay "gamekit/src/base/relay"
	"gamekit/src/base/utils"
	"gamesvr/scripts/model"
	"math"
	"time"

	cs "gamekit/pkg/cs"
	"gamekit/pkg/game_third"
	"gamekit/pkg/oss"
	ss "gamekit/pkg/ss"

	"google.golang.org/protobuf/proto"
)

func init() {
	// 测试代码
	RegistTransmitHandle(uint32(cs.EnmCmdValue_ECV_CSChatReq), HandleTransChatReq)
	RegistTransmitHandle(uint32(ss.EnmCmdValue_ECV_SSHeartReq), HandleTransHeartReq)

	RegistTransmitHandle(uint32(ss.EnmCmdValue_ECV_SSTransTestRes), handleTransTestRes)
	RegistTransmitHandle(uint32(ss.EnmCmdValue_ECV_SSTransTestReq), handleTransTestReq)
}

func HandleTransReq(head *ss.SSHead, buffer []byte) {
	subBitMarkOffset := 4
	subHeadLenOffset := subBitMarkOffset + 1
	subHeadOffset := subHeadLenOffset + 1
	subHeadLen := int(buffer[subHeadLenOffset])

	// 直接透传到客户端的消息
	if head.Flag == uint32(ss.EnmMsgFlag_EMF_ToClient) {
		player := model.PlayerMgr.GetPlayerByAccId(head.AccId)
		if player == nil {
			logger.Log.Error("trans client message, but not found player. Head: %s", head.String())
			return
		}

		if player.IsControlActor {
			return
		}

		// 转发类型消息，如果需要携带extend，需要反序列化，并序列化，有不必要的性能消耗
		conn := player.GetConn()
		if utils.IsNetLog() {
			logger.Log.Debug("%s|trans client message success. Head: %s", player.String(), head.String())
		}

		sendBuffer, err := model.PressAndEncrypt(player, buffer[subHeadOffset:], subHeadLen)
		if err != nil {
			logger.Log.Error("%s|press and encrypt message failed. Error: %s, Head: %s", player, err.Error(), head)
			return
		}
		// 发送数据包大小超限检测
		if float64(len(sendBuffer))/float64(model.MaxWsMessageSize) > model.OverLoadAlarmRate {
			osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, player.AccId,
				"The size of the data packet reaches %2.0f%% of the maximum carrying capacity. cmdId: %d, size: %d, maxSize: %d",
				model.OverLoadAlarmRate*100, head.CmdId, len(sendBuffer), model.MaxWsMessageSize)
		}

		utils.StatMgr.StatCmdCall(head.CmdId, len(sendBuffer), time.Duration(0))
		if err := utils.ConnAsyncWrite(conn, sendBuffer); err != nil {
			logger.Log.Error("%s|trans client message, conn.Write failed. Head: %s, Err: %#v, len(buffer): %d",
				player.String(), head.String(), err, len(buffer))
			return
		}

		return
	}

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

		if utils.IsNetLog() {
			logger.Log.Debug("recv transmit message. Head: %s, SubHead: %s, SubMsg: %s", head.String(), subHead.String(), subMsg.String())
		}
		handler(head, subHead, subMsg)
	} else {
		// 服务器消息转发
		subHead := &ss.SSHead{}

		var (
			subMsg utils.IMessage
			ok     bool
		)
		if head.SubCmdId == uint32(ss.EnmCmdValue_ECV_SSMiscReq) {
			subMsg = &game_third.SSMiscReq{}
			ok = true
		} else {
			subMsg, ok = ss.NewServiceMessage(int32(head.SubCmdId))
		}
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

		if utils.IsNetLog() {
			logger.Log.Debug("recv transmit message. Head: %s, SubHead: %s, SubMsg: %s", head.String(), subHead.String(), subMsg.String())
		}
		handler(head, subHead, subMsg)
	}
}

func HandleTransChatReq(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subHead := iSubHead.(*cs.CSHead)
	subMsg := iSubMsg.(*cs.CSChatReq)

	if utils.IsNetLog() {
		logger.Log.Debug("call HandleTransChatReq. Head: %s, SubHead: %s, SubMsg: %s", head.String(), subHead.String(), subMsg.String())
	}
}

func HandleTransHeartReq(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subHead := iSubHead.(*ss.SSHead)
	subMsg := iSubMsg.(*ss.SSHeartReq)

	if utils.IsNetLog() {
		logger.Log.Debug("call HandleTransHeartReq. Head: %s, SubHead: %s, SubMsg: %s", head.String(), subHead.String(), subMsg.String())
	}
}

func handleTransTestReq(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subHead := iSubHead.(*ss.SSHead)
	subMsg := iSubMsg.(*ss.SSTransTestReq)

	logger.Log.Debug("call handleTransTestReq. Head: %s, SubHead: %s, SubMsg: %s", head.String(), subHead.String(), subMsg.String())
	testRes := &ss.SSTransTestReq{
		Value: subMsg.Value,
	}
	ssHead := &ss.SSHead{CmdId: uint32(ss.EnmCmdValue_ECV_SSTransTestRes), SeqId: uint32(math.MaxInt32)}
	relay.TransMsg(0, head.Src, 0, ssHead, testRes)
}

func handleTransTestRes(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subHead := iSubHead.(*ss.SSHead)
	subMsg := iSubMsg.(*ss.SSTransTestRes)

	logger.Log.Debug("call handleTransTestRes. Head: %s, SubHead: %s, SubMsg: %s", head.String(), subHead.String(), subMsg.String())
}
