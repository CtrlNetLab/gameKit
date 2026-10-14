package ctrl

import (
	"fmt"

	ss "gamekit/pkg/ss"
	"gamekit/src/base/control"
	"gamekit/src/base/utils"
	"roomsvr/scripts/config"
)

// Init 注册 GM/控制命令(脚手架最小示例)
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
	}

	control.AddControls(ctrls)
}

func getSelfConnIdx(args []string, cmdType int, param interface{}) (string, error) {
	iType := ss.ServiceType_RoomSvr
	idx := config.GetServiceId()
	return fmt.Sprintf("%d", ((uint64(iType)&0xff)<<16)|(uint64(idx)&0xffff)), nil
}

func reloadConfig(args []string, cmdType int, param interface{}) (string, error) {
	config.LoadService(false)
	return "reload config success", nil
}
