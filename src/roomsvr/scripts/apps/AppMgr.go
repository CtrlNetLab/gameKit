package apps

import (
	cs "gamekit/pkg/cs"
	"gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"roomsvr/scripts/model"

	gnet "github.com/walkon/wsgnet"
)

// type HandleService func(conn gnet.Conn, head *ss.SSHead, iMsg interface{})
type HandleService func(conn gnet.Conn, head *ss.SSHead, iMsg interface{})
type HandleTransmit func(head *ss.SSHead, subHead interface{}, subMsg interface{})
type HandleClient func(player *model.Player, ssHead *ss.SSHead, csHead *cs.CSHead, iMsg interface{})

var (
	handleServiceMap  map[ss.EnmCmdValue]HandleService
	handleTransmitMap map[uint32]HandleTransmit
	handleClientMap   map[cs.EnmCmdValue]HandleClient
)

func init() {
	handleServiceMap = make(map[ss.EnmCmdValue]HandleService)
	handleTransmitMap = make(map[uint32]HandleTransmit)
	handleClientMap = make(map[cs.EnmCmdValue]HandleClient)
}

func RegistServiceHandle(cmdId ss.EnmCmdValue, handle HandleService) int {
	if _, ok := handleServiceMap[cmdId]; ok {
		logger.Log.Error("duplicate register cmd. CmdId: %d", cmdId)
		return -1
	}
	if utils.IsNetLog() {
		logger.Log.Debug("register cmd success. CmdId: %d", cmdId)
	}
	handleServiceMap[cmdId] = handle
	return 0
}

func GetServiceHandle(cmdId uint32) HandleService {
	if v, ok := handleServiceMap[ss.EnmCmdValue(cmdId)]; !ok {
		return nil
	} else {
		return v
	}
	return nil
}

func RegistTransmitHandle(cmdId uint32, handle HandleTransmit) int {
	if _, ok := handleTransmitMap[cmdId]; ok {
		logger.Log.Error("duplicate register transmit cmd. CmdId: %d", cmdId)
		return -1
	}
	if utils.IsNetLog() {
		logger.Log.Debug("register transmit cmd success. CmdId: %d", cmdId)
	}
	handleTransmitMap[cmdId] = handle
	return 0
}

func GetTransmitHandle(cmdId uint32) HandleTransmit {
	if v, ok := handleTransmitMap[cmdId]; !ok {
		return nil
	} else {
		return v
	}
}

func RegistClientHandle(cmdId cs.EnmCmdValue, handle HandleClient) int {
	if _, ok := handleClientMap[cmdId]; ok {
		logger.Log.Error("duplicate register cmd. CmdId: %d", cmdId)
		return -1
	}
	if utils.IsNetLog() {
		logger.Log.Debug("register cmd success. CmdId: %d", cmdId)
	}
	handleClientMap[cmdId] = handle
	return 0
}

func GetClientHandle(cmdId uint32) HandleClient {
	if v, ok := handleClientMap[cs.EnmCmdValue(cmdId)]; !ok {
		return nil
	} else {
		return v
	}
	return nil
}
