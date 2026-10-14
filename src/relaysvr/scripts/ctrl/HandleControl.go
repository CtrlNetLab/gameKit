package ctrl

import (
	"gamekit/src/base/control"
	"gamekit/src/base/utils"
	"relaysvr/scripts/service"
)

func Init() {
	ctrls := map[string]*control.CtrlHandleData{
		"showAllService": &control.CtrlHandleData{Usage: "showAllService", Desc: "显示所有在线服务",
			CmdType: utils.CmdType_All, GmType: utils.GmType_Root, Tags: []string{control.TagServer, control.TagNetwork, control.TagQuery}, Handle: showAllService},
	}

	control.AddControls(ctrls)
}

func showAllService(args []string, CmdType int, param interface{}) (string, error) {
	return service.ServiceMgr.String(), nil
}
