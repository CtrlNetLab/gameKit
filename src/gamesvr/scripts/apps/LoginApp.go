package apps

import (
	cs "gamekit/pkg/cs"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"gamesvr/scripts/model"
)

// 登陆示例 App: 演示客户端协议处理的标准写法
// 完整业务(三方认证、账号数据加载、创角等)请在此基础上扩展
func init() {
	// 玩家登陆
	RegistClientHandle(cs.EnmCmdValue_ECV_CS_AC_LoginReq, HandleLoginReq)
}

// HandleLoginReq 处理玩家登陆请求
func HandleLoginReq(player *model.Player, head *cs.CSHead, iMsg interface{}) {
	loginReq, ok := iMsg.(*cs.CS_AC_LoginReq)
	if player == nil || head == nil || !ok || loginReq == nil {
		return
	}

	connIdx := player.GetConnIdx()
	logger.Log.Error("%d|0|loginStep 1|HandleLoginReq, LoginReq: %s", connIdx, loginReq.String())

	// 1. 平台认证(脚手架示例: 游客/平台登录均直接通过, 业务请替换为真实的三方校验)
	if !checkAccess(player, loginReq) {
		player.SendClientMsg(cs.EnmCmdValue_ECV_CS_AC_LoginRes, head.SeqId,
			cs.EnmRetCode_LOGIN_AUTH_ACCOUNT, "auth failed", &cs.CS_AC_LoginRes{RetCode: uint32(cs.EnmRetCode_LOGIN_AUTH_ACCOUNT)})
		return
	}

	// 2. 分配账号ID并完成登陆(脚手架示例: 每次登录生成新账号, 业务请改为账号体系查询/创建)
	player.AccId = model.GenAccId()
	player.Sid = utils.Uint64ToStr(player.AccId)
	player.SetLoginState(model.LoginState_Success)
	model.PlayerMgr.AddPlayer(player)

	// 3. 返回登陆成功
	player.SendClientMsg(cs.EnmCmdValue_ECV_CS_AC_LoginRes, head.SeqId,
		cs.EnmRetCode_SUCCESS, "", &cs.CS_AC_LoginRes{
			RetCode:    uint32(cs.EnmRetCode_SUCCESS),
			ServerTime: utils.GetTimeMS(),
			AccId:      player.AccId,
		})

	logger.Log.Error("%d|0|loginStep 2|login success. accId: %d", connIdx, player.AccId)
}

// checkAccess 平台认证(脚手架示例实现)
func checkAccess(player *model.Player, loginReq *cs.CS_AC_LoginReq) bool {
	if loginReq.LoginType == 0 {
		// 游客登录: 必须有 TouristId
		return loginReq.TouristId != ""
	}
	// 平台登录: 必须有 PlatformId
	return loginReq.PlatformId != ""
}
