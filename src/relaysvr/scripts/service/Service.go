package service

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/utils"
	"fmt"

	gnet "github.com/walkon/wsgnet"
)

type Service struct {
	// @see ss_type.proto:ServerType
	stype ss.ServiceType
	// 服务ID，每个服务器 <Type, ID> 都是唯一的
	id int32
	// 地区Id，非地区共享的服务为0
	AreaId int32

	conn gnet.Conn

	// 单位: 毫秒
	lastRecvMsgTm int64
}

func (this *Service) String() string {
	return fmt.Sprintf("{type: %s, id: %d, cidx: %d, areaId: %d, recvtime: %s}",
		this.stype.String(), this.id, utils.ServiceConnIdx(this.stype, this.id), this.AreaId, utils.FormatUnixTime(this.lastRecvMsgTm/1000))
}

func NewService(stype, id int32, conn gnet.Conn) *Service {
	serv := &Service{stype: ss.ServiceType(stype), id: id, conn: conn, lastRecvMsgTm: utils.GetTimeMS()}
	return serv
}

func (s *Service) SetLastRecvMsgTm(now int64) {
	s.lastRecvMsgTm = now
}

func (s *Service) GetLastRecvMsgTm() int64 {
	return s.lastRecvMsgTm
}

func (s *Service) SetConn(c gnet.Conn) {
	s.conn = c
}

func (s *Service) GetConn() gnet.Conn {
	return s.conn
}

func (s *Service) LogInfo() string {
	return fmt.Sprintf("%d-%d-%d", s.stype, s.id, s.AreaId)
}

func (s *Service) GetConnIdx() uint64 {
	return utils.ServiceConnIdx(s.stype, s.id)
}
