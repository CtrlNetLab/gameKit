package apps

import (
	cs "gamekit/pkg/cs"
	"gamekit/pkg/game_third"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	utils "gamekit/src/base/utils"
	"gamesvr/scripts/model"

	gnet "github.com/walkon/wsgnet"
)

type HandleService func(conn gnet.Conn, head *ss.SSHead, iMsg interface{})
type HandleMisc func(conn gnet.Conn, head *ss.SSHead, miscReq *game_third.SSMiscReq, iMsg interface{})

type HandleClient func(player *model.Player, head *cs.CSHead, iMsg interface{})
type HandleTransmit func(head *ss.SSHead, subHead interface{}, subMsg interface{})

var (
	handleClientMap   map[cs.EnmCmdValue]HandleClient
	handleServiceMap  map[ss.EnmCmdValue]HandleService
	handleTransmitMap map[uint32]HandleTransmit
	handleMiscMap     map[string]HandleMisc
)

func init() {
	handleClientMap = make(map[cs.EnmCmdValue]HandleClient)
	handleServiceMap = make(map[ss.EnmCmdValue]HandleService)
	handleTransmitMap = make(map[uint32]HandleTransmit)
	handleMiscMap = make(map[string]HandleMisc)
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
}

func RegistMiscHandle(msgName string, handle HandleMisc) int {
	if _, ok := handleMiscMap[msgName]; ok {
		logger.Log.Error("duplicate register misc cmd. MsgName: %s", msgName)
		return -1
	}
	if utils.IsNetLog() {
		logger.Log.Debug("register misc cmd success. MsgName: %s", msgName)
	}
	handleMiscMap[msgName] = handle
	return 0
}

func GetMiscHandle(msgName string) HandleMisc {
	if v, ok := handleMiscMap[msgName]; !ok {
		return nil
	} else {
		return v
	}
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
