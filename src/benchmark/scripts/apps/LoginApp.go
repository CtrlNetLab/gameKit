package apps

import (
	"benchmark/scripts/model"
	"gamekit/pkg/cs"
	"gamekit/src/base/logger"
)

func init() {
	// 登陆响应: 压测最小示例, 业务模块按此模式注册
	model.RegistClientHandle(cs.EnmCmdValue_ECV_CS_AC_LoginRes, HandleLoginRes)
	model.RegistClientHandle(cs.EnmCmdValue_ECV_CSHeartRes, HandleHeartRes)
}

// HandleLoginRes 处理登陆响应
func HandleLoginRes(player *model.Player, head *cs.CSHead, iMsg interface{}) {
	loginRes, ok := iMsg.(*cs.CS_AC_LoginRes)
	if !ok {
		logger.Log.Error("login res type error. Msg: %#v", iMsg)
		return
	}

	if loginRes.RetCode != uint32(cs.EnmRetCode_SUCCESS) {
		logger.Log.Error("%s|login failed. RetCode: %d", player, loginRes.RetCode)
		return
	}

	player.LoginSuccess(loginRes)
}

// HandleHeartRes 处理心跳响应
func HandleHeartRes(player *model.Player, head *cs.CSHead, iMsg interface{}) {
}
