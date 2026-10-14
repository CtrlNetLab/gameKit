package apps

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/logger"

	gnet "github.com/walkon/wsgnet"
)

// type HandleService func(conn gnet.Conn, head *ss.SSHead, iMsg interface{})
type HandleService func(conn gnet.Conn, head *ss.SSHead, iMsg interface{})

var (
	handleServiceMap map[ss.EnmCmdValue]HandleService
)

func init() {
	handleServiceMap = make(map[ss.EnmCmdValue]HandleService)
}

func RegistServiceHandle(cmdId ss.EnmCmdValue, handle HandleService) int {
	if _, ok := handleServiceMap[cmdId]; ok {
		logger.Log.Error("duplicate register cmd. CmdId: %d", cmdId)
		return -1
	}
	logger.Log.Debug("register cmd success. CmdId: %d", cmdId)
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
