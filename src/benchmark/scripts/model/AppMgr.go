package model

import (
	"gamekit/pkg/cs"
	"gamekit/src/base/logger"
)

type HandleClient func(player *Player, head *cs.CSHead, iMsg interface{})

var (
	handleClientMap map[cs.EnmCmdValue]HandleClient
)

func init() {
	handleClientMap = make(map[cs.EnmCmdValue]HandleClient)
}

func RegistClientHandle(cmdId cs.EnmCmdValue, handle HandleClient) int {
	if _, ok := handleClientMap[cmdId]; ok {
		logger.Log.Error("duplicate register cmd. CmdId: %d", cmdId)
		return -1
	}
	logger.Log.Debug("register cmd success. CmdId: %d", cmdId)
	handleClientMap[cmdId] = handle
	return 0
}

func GetClientHandle(cmdId uint32) HandleClient {
	if v, ok := handleClientMap[cs.EnmCmdValue(cmdId)]; !ok {
		return nil
	} else {
		return v
	}
}
