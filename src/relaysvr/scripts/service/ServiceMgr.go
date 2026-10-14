package service

import (
	"gamekit/pkg/oss"
	"gamekit/pkg/ss"
	"gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/utils"
	"encoding/binary"
	"errors"
	"fmt"
	"relaysvr/scripts/config"
	"time"

	"google.golang.org/protobuf/proto"
)

var (
	ServiceMgr *serviceMgr = &serviceMgr{}
)

type serviceMgr struct {
	services map[uint64]*Service
}

func (this *serviceMgr) Init() {
	this.services = make(map[uint64]*Service)

	utils.AddTimer("ServiceMgr::tick1s", time.Second, this.tick1s)
}

func (this *serviceMgr) String() string {
	typeMap := make(map[ss.ServiceType][]string)
	for _, serv := range this.services {
		if _, ok := typeMap[serv.stype]; !ok {
			typeMap[serv.stype] = []string{serv.String()}
		} else {
			typeMap[serv.stype] = append(typeMap[serv.stype], serv.String())
		}
	}
	var logstr string = ""
	for _, types := range typeMap {
		for _, stype := range types {
			logstr += fmt.Sprintf("%s\n", stype)
		}
	}
	return logstr
}

func (this *serviceMgr) GetServiceMap() map[uint64]*Service {
	return this.services
}

func (sm *serviceMgr) AddService(s *Service) bool {
	connIdx := utils.ServiceConnIdx(s.stype, s.id)
	if _, ok := sm.services[connIdx]; ok {
		logger.Log.Error("%s|add service failed, duplicate id.", s.LogInfo())
		return false
	}
	sm.services[connIdx] = s
	logger.Log.Error("%s|add service success.", s.LogInfo())
	return true
}

func (sm *serviceMgr) GetService(connIdx uint64) *Service {
	if v, ok := sm.services[connIdx]; ok {
		return v
	}
	return nil
}

func (sm *serviceMgr) GetRandService(itype int32) *Service {
	var serv *Service
	for connIdx, s := range sm.services {
		it, _ := utils.ServiceTypeIdx(connIdx)
		if it == itype && (serv == nil || serv.lastRecvMsgTm < s.lastRecvMsgTm) {
			serv = s
		}
	}
	return serv
}

func (sm *serviceMgr) GetRandGameService() *Service {
	return sm.GetRandService(int32(ss.ServiceType_GameSvr))
}

func (sm *serviceMgr) GetRandMatchService() *Service {
	return sm.GetRandService(int32(ss.ServiceType_MatchSvr))
}

func (sm *serviceMgr) DelService(connIdx uint64) bool {
	if _, ok := sm.services[connIdx]; !ok {
		logger.Log.Error("delete service failed, not found service. ConnIdx: %d", connIdx)
		return false
	} else {
		// 这里不能再调用 serv.GetConn().Close() 了
		delete(sm.services, connIdx)
		logger.Log.Error("delete service success. ConnIdx: %d", connIdx)
	}
	return true
}

func (sm *serviceMgr) tick1s(now time.Time) {
	nowMS := now.UnixNano() / 1e6
	timeOut := config.Conf.Common.ConnTimeOut

	var delServices []uint64
	for connIdx, s := range sm.services {
		// 服务长时间没有心跳包
		if s.GetLastRecvMsgTm()+timeOut < nowMS {
			logger.Log.Error("%s|service timeout. LastTm: %d, Now: %d, TimeOut: %d", s.LogInfo(), s.GetLastRecvMsgTm(), nowMS, timeOut)
			delServices = append(delServices, connIdx)
		}
	}

	for _, v := range delServices {
		if serv, ok := sm.services[v]; ok {
			serv.GetConn().SetContext(nil)
			delete(sm.services, v)
			// 关闭连接
			serv.GetConn().Close()
		}
	}
}

func BroadcastBuffer(tHead *ss.SSHead, subBuffer []byte) error {
	if tHead == nil {
		return errors.New("TransBuffer head nil")
	}

	if tHead.Flag != uint32(ss.EnmMsgFlag_EMF_ToAllGameSvr) {
		// @TODO ryder 从 redis 获取 accid 所在的 gamesvr
		return fmt.Errorf("BroadcastBuffer need dst address. subHead: %s", tHead.String())
	}

	subMsgLen := len(subBuffer)

	// 外层包装协议
	// tHeadLen := tHead.Size()
	tHeadLen := proto.Size(tHead)
	totallen := 2 + tHeadLen + subMsgLen

	buffer := make([]byte, totallen, totallen)
	buffer[0] = byte(0)
	buffer[1] = byte(tHeadLen)

	headOffset := 2
	headBytes, err := proto.Marshal(tHead)
	if err != nil {
		logger.Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", tHead.String(), err)
		return err
	}
	copy(buffer[headOffset:], headBytes)

	subOffset := headOffset + tHeadLen
	copy(buffer[subOffset:], subBuffer)

	for _, serv := range ServiceMgr.services {
		if serv == nil {
			logger.Log.Error("trans not found service. tHead: %s", tHead.String())
			continue
		}

		if serv.stype != ss.ServiceType_GameSvr {
			continue
		}

		if err := utils.ConnAsyncWrite(serv.conn, buffer); err != nil {
			logger.Log.Error("%s|trans conn.Write failed. tHead: %s, Err: %v, len(buffer): %d",
				serv.LogInfo(), tHead.String(), err, len(buffer))
		}
	}

	return nil
}

func TransBufferDst(tHead *ss.SSHead, subBuffer []byte) {
	if tHead == nil {
		logger.Log.Debug("trans failed. head: %s, buflen: %d", tHead.String(), len(subBuffer))
		return
	}

	callback := func() {
		if tHead.Dst == 0 {
			tHead.Dst = uint64(ss.ServiceType_GameSvr)
		}
		subMsgLen := len(subBuffer)
		// 外层包装协议
		tHeadLen := proto.Size(tHead)
		totallen := 2 + tHeadLen + subMsgLen

		buffer := make([]byte, totallen, totallen)
		buffer[0] = byte(0)
		buffer[1] = byte(tHeadLen)

		headOffset := 2
		headBytes, err := proto.Marshal(tHead)
		if err != nil {
			logger.Log.Error("pack proto.Marshal failed. Head: %s, Err: %v", tHead.String(), err)
			return
		}
		copy(buffer[headOffset:], headBytes)

		subOffset := headOffset + tHeadLen
		copy(buffer[subOffset:], subBuffer)

		var serv *Service
		if tHead.Dst > uint64(ss.ServiceType_MaxType) {
			// 转发到指定服务器
			serv = ServiceMgr.GetService(tHead.Dst)
		} else {
			serv = ServiceMgr.GetRandService(int32(tHead.Dst))
		}

		if serv == nil {
			logger.Log.Error("trans not found service. tHead: %s", tHead.String())
			osslog.OSS.ElkAlarm(oss.EnmElk_Network_Error, tHead.AccId, "trans not found service. tHead: %s", tHead.String())
			return
		}

		if err := utils.ConnAsyncWrite(serv.conn, buffer); err != nil {
			logger.Log.Error("%s|trans conn.Write failed. tHead: %s, Err: %v, len(buffer): %d",
				serv.LogInfo(), tHead.String(), err, len(buffer))
			return
		}

		if utils.IsNetLog() {
			logger.Log.Debug("trans success. connIdx: %d. head: %s, buflen: %d", serv.GetConnIdx(), tHead.String(), len(buffer))
		}
	}

	utils.WorkerPool.AddTask(0, "TransBuffer", func() {
		tHead.Dst = dao.HGetGameId(config.GetOnlineKey(), tHead.AccId)
		utils.WorkerPool.AddCallback("TransBuffer", callback)
	})
}

func TransBuffer(tHead *ss.SSHead, subBuffer []byte) error {
	if tHead == nil {
		return errors.New("TransBuffer head nil")
	}

	if tHead.Dst == 0 {
		return fmt.Errorf("TransMsg need dst address. subHead: %s", tHead.String())
	}

	subMsgLen := len(subBuffer)

	// 外层包装协议
	// tHeadLen := tHead.Size()
	tHeadLen := proto.Size(tHead)
	totallen := 2 + tHeadLen + subMsgLen

	buffer := make([]byte, totallen, totallen)
	buffer[0] = byte(0)
	buffer[1] = byte(tHeadLen)

	headOffset := 2
	headBytes, err := proto.Marshal(tHead)
	if err != nil {
		logger.Log.Error("pack proto.Marshal failed. Head: %s, Err: %v", tHead.String(), err)
		return err
	}
	copy(buffer[headOffset:], headBytes)

	subOffset := headOffset + tHeadLen
	copy(buffer[subOffset:], subBuffer)

	var serv *Service
	if tHead.Dst > uint64(ss.ServiceType_MaxType) {
		// 转发到指定服务器
		serv = ServiceMgr.GetService(tHead.Dst)
	} else {
		serv = ServiceMgr.GetRandService(int32(tHead.Dst))
	}

	if serv == nil {
		logger.Log.Error("trans not found service. tHead: %s", tHead.String())
		// osslog.OSS.ElkAlarm(oss.EnmElk_Network_Error, tHead.AccId, "trans not found service. tHead: %s", tHead.String())
		return fmt.Errorf("not found service")
	}

	if err := utils.ConnAsyncWrite(serv.conn, buffer); err != nil {
		logger.Log.Error("%s|trans conn.Write failed. tHead: %s, Err: %v, len(buffer): %d",
			serv.LogInfo(), tHead.String(), err, len(buffer))
		return err
	}

	return nil
}

func TransMsg(tHead *ss.SSHead, subHead, subMsg utils.IMessage) error {
	if tHead == nil {
		return errors.New("TransBuffer head nil")
	}

	if tHead.Dst == 0 {
		return fmt.Errorf("TransMsg need dst address. subHead: %s", subHead.String())
	}

	// 子消息也是按统一方式进行封装 |length|bitmark|head-length|message-buffer|
	subHeadLen := proto.Size(subHead)
	subMsgLen := 4 + 2 + subHeadLen + proto.Size(subMsg)

	// 外层包装协议
	tHeadLen := proto.Size(tHead)
	totallen := 2 + tHeadLen + subMsgLen

	buffer := make([]byte, totallen, totallen)
	buffer[0] = byte(0)
	buffer[1] = byte(tHeadLen)

	headOffset := 2
	headBytes, err := proto.Marshal(tHead)
	if err != nil {
		logger.Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", tHead.String(), err)
		return err
	}
	copy(buffer[headOffset:], headBytes)

	subLenOffset := headOffset + tHeadLen
	binary.BigEndian.PutUint32(buffer[subLenOffset:], uint32(subMsgLen))

	subBitMarkOffset := subLenOffset + 4
	buffer[subBitMarkOffset] = byte(0)

	subHeadLenOffset := subBitMarkOffset + 1
	buffer[subHeadLenOffset] = byte(subHeadLen)

	subHeadOffset := subHeadLenOffset + 1
	subHeadBytes, err := proto.Marshal(subHead)
	if err != nil {
		logger.Log.Error("pack proto.Marshal failed. Head: %s, Err: %#v", subHead.String(), err)
		return err
	}
	copy(buffer[subHeadOffset:], subHeadBytes)

	subMsgOffset := subHeadOffset + subHeadLen
	subMsgBytes, err := proto.Marshal(subMsg)
	if err != nil {
		logger.Log.Error("trans pack proto.Marshal failed. Head: %s, Message: %s, Err: %#v", subHead.String(), subMsg.String(), err)
		return err
	}
	copy(buffer[subMsgOffset:], subMsgBytes)

	serv := ServiceMgr.GetService(tHead.Dst)

	if serv == nil {
		logger.Log.Error("trans not found service. tHead: %s, subHead: %s, subMsg: %s", tHead.String(), subHead.String(), subMsg.String())
		osslog.OSS.ElkAlarm(oss.EnmElk_Network_Error, tHead.AccId, "trans not found service. tHead: %s", tHead.String())
		return fmt.Errorf("not found service")
	}

	if err := utils.ConnAsyncWrite(serv.conn, buffer); err != nil {
		logger.Log.Error("%s|trans conn.Write failed. subHead: %s, subMsg: %s, Err: %v, len(buffer): %d",
			serv.LogInfo(), subHead.String(), subMsg.String(), err, len(buffer))
		return err
	}

	return nil
}
