package apps

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"relaysvr/scripts/config"
	"relaysvr/scripts/service"

	gnet "github.com/walkon/wsgnet"
)

func init() {
	RegistServiceHandle(ss.EnmCmdValue_ECV_SSRegistReq, HandleRegistReq)
	RegistServiceHandle(ss.EnmCmdValue_ECV_SSHeartReq, HandleHeartReq)
	RegistServiceHandle(ss.EnmCmdValue_ECV_SSGetAllServiceReq, HandleGetAllServiceReq)
}

func HandleRegistReq(conn gnet.Conn, head *ss.SSHead, iMsg interface{}) {
	registReq := iMsg.(*ss.SSRegistReq)

	registRes := &ss.SSRegistRes{RelayAreaId: config.GetAreaId()}
	headRes := utils.NewSSHead(ss.EnmCmdValue_ECV_SSRegistRes, head.SeqId, ss.EnmRetCode_SUCCESS)

	connIdx := utils.ServiceConnIdx(ss.ServiceType(registReq.Type), registReq.Id)
	conn.SetContext(connIdx)
	if s := service.ServiceMgr.GetService(connIdx); s != nil {
		logger.Log.Error("%s service already regist, don't duplicate regist", s.LogInfo())
		headRes.RetCode = uint32(ss.EnmRetCode_DUPLICATE_REGIST)
	} else {
		serv := service.NewService(registReq.Type, registReq.Id, conn)
		serv.AreaId = registReq.AreaId
		service.ServiceMgr.AddService(serv)
		logger.Log.Error("%s|connIdx: %d, remote: %s. regist service success.", serv.LogInfo(), serv.GetConnIdx(), conn.RemoteAddr().String())
	}

	utils.SendMsg(conn, headRes, registRes)
}

func HandleHeartReq(conn gnet.Conn, head *ss.SSHead, iMsg interface{}) {
	headRes := utils.NewSSHead(ss.EnmCmdValue_ECV_SSHeartRes, head.SeqId, ss.EnmRetCode_SUCCESS)
	utils.SendMsg(conn, headRes, &ss.SSHeartRes{})
}

func HandleTransReq(head *ss.SSHead, buffer []byte) {
	if head == nil {
		return
	}

	if head.Flag == uint32(ss.EnmMsgFlag_EMF_ToAllGameSvr) {
		err := service.BroadcastBuffer(head, buffer)
		if utils.IsNetLog() {
			if err == nil {
				logger.Log.Debug("trans all gamsvr success. size: %d, head: %s", len(buffer), head)
			} else {
				logger.Log.Debug("trans failed. size: %d, head: %s, err: %s", len(buffer), head, err.Error())
			}
		}
	} else {
		if head.Dst == 0 && head.AccId != 0 {
			// 传给玩家
			service.TransBufferDst(head, buffer)
		} else {
			err := service.TransBuffer(head, buffer)
			if utils.IsNetLog() {
				if err == nil {
					logger.Log.Debug("trans success. size: %d, head: %s", len(buffer), head)
				} else {
					logger.Log.Debug("trans failed. size: %d, head: %s, err: %s", len(buffer), head, err)
				}
			}
		}
	}
}

func HandleGetAllServiceReq(conn gnet.Conn, head *ss.SSHead, iMsg interface{}) {
	headRes := utils.NewSSHead(ss.EnmCmdValue_ECV_SSGetAllServiceRes, head.SeqId, ss.EnmRetCode_SUCCESS)
	getServiceRes := &ss.SSGetAllServiceRes{}
	for connIdx := range service.ServiceMgr.GetServiceMap() {
		getServiceRes.ConnIdxList = append(getServiceRes.ConnIdxList, connIdx)
	}
	utils.SendMsg(conn, headRes, getServiceRes)
}
