package apps

import (
	cs "gamekit/pkg/cs"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"gamesvr/scripts/model"
)

// 游戏基础 App: 心跳等基础协议示例
func init() {
	// 原心跳包
	RegistClientHandle(cs.EnmCmdValue_ECV_CSHeartReq, HandleHeartReq)
}

// HandleHeartReq 处理心跳包请求
func HandleHeartReq(player *model.Player, head *cs.CSHead, iMsg interface{}) {
	if utils.IsNetLog() {
		logger.Log.Debug("%s|recv heartbeat request", player.String())
	}

	// 收到心跳包，不需要做任何处理，在 React 已经更新了最后一次收包时间戳
	player.SendClientMsg(cs.EnmCmdValue_ECV_CSHeartRes, head.SeqId, cs.EnmRetCode_SUCCESS, "", &cs.CSHeartRes{ServerTime: utils.GetTimeMS()})
}
