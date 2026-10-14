package misc

import (
	"gamekit/pkg/game_third"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	utils "gamekit/src/base/utils"
)

func HandleMiscReq(head *ss.SSHead, iSubHead interface{}, iSubMsg interface{}) {
	subHead, hok := iSubHead.(*ss.SSHead)
	subMsg, mok := iSubMsg.(*game_third.SSMiscReq)

	if head == nil || !hok || !mok || subHead == nil || subMsg == nil {
		return
	}

	if utils.IsNetLog() {
		logger.Log.Error("recv misc message, subHead: %s, Msg: %s", subHead.String(), subMsg.String())
	}

	MiscMgr.Handle(head, subHead, subMsg)
}
