package ctrl

import (
	"fmt"

	"gamekit/src/base/control"
	"gamekit/src/base/utils"
	"gamesvr/scripts/config"
	"gamesvr/scripts/model"
)

// Init 注册 GM/控制命令(脚手架最小示例)
// 完整 GM 命令请在 ctrls map 中按同一模式扩展
func Init() {
	ctrls := map[string]*control.CtrlHandleData{
		"getSelfConnIdx": {
			Usage: "getSelfConnIdx#", Desc: "获取自身连接索引",
			CmdType: utils.CmdType_All, GmType: utils.GmType_Root,
			Tags:   []string{control.TagServer, control.TagDebug, control.TagNetwork},
			Handle: getSelfConnIdx,
		},

		"reloadConfig": {
			Usage: "reloadConfig#", Desc: "热重载服务配置(service.yaml)",
			CmdType: utils.CmdType_All, GmType: utils.GmType_Root,
			Tags:   []string{control.TagConfig, control.TagDebug},
			Handle: reloadConfig,
		},

		"getOnlineNum": {
			Usage: "getOnlineNum#", Desc: "获取当前在线人数",
			CmdType: utils.CmdType_All, GmType: utils.GmType_Root,
			Tags:   []string{control.TagServer, control.TagDebug},
			Handle: getOnlineNum,
		},
	}

	control.AddControls(ctrls)
}

func getSelfConnIdx(args []string, cmdType int, param interface{}) (string, error) {
	return fmt.Sprintf("%d", config.GetSelfConnIdx()), nil
}

func reloadConfig(args []string, cmdType int, param interface{}) (string, error) {
	config.LoadYaml(false)
	return "reload config success", nil
}

func getOnlineNum(args []string, cmdType int, param interface{}) (string, error) {
	return fmt.Sprintf("online num: %d", model.PlayerMgr.GetOnlineNum()), nil
}
