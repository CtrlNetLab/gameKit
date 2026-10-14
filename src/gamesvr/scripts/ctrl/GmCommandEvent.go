package ctrl

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/control"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"gamesvr/scripts/model"
)

func init() {
	utils.AddEventHandler(ss.EventType_EId_CtrlCommand, HandleGmCommandEvent)
}

func HandleGmCommandEvent(event *utils.Event) {
	if event == nil || event.EventId != ss.EventType_EId_CtrlCommand {
		return
	}

	player, ok := event.IObj.(*model.Player)
	if !ok || player == nil || !player.IsOnlineOK() {
		return
	}

	command, ok := event.ParamI.([]string)
	if !ok || len(command) == 0 {
		logger.Log.Error("%s|invailed command: %v", player, event.ParamI)
		return
	}
	logger.Log.Info("%s|exec offline gm command: %v", player, command)
	control.Handler(command, utils.CmdType_Gm, player)
}
