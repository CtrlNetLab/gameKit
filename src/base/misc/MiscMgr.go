package misc

import (
	"gamekit/pkg/game_third"
	"gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/relay"
	utils "gamekit/src/base/utils"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var (
	MiscMgr = &miscMgr{}
)

func init() {
	MiscMgr.Init()
}

type MiscData struct {
	Head       *ss.SSHead
	SubHead    *ss.SSHead
	MiscReq    *game_third.SSMiscReq
	Request    interface{}
	ResMsgName string
}

type HandleMisc func(MiscData *MiscData)

type HandleData struct {
	Handle   HandleMisc
	Template proto.Message
}

type miscMgr struct {
	handleMap map[string]*HandleData
}

func (this *miscMgr) Init() {
	// "SPTestServiceReq": &handleData{Handle: handleTestServiceReq, msg: &game_third.SPTestServiceReq{}},
	this.handleMap = make(map[string]*HandleData)
}

func (this *miscMgr) AddMsg(handle HandleMisc, template proto.Message) {
	fullName := proto.MessageName(template)
	fields := strings.Split(string(fullName), ".")
	name := fields[len(fields)-1]

	this.handleMap[name] = &HandleData{Handle: handle, Template: template}
}

func (this *miscMgr) Handle(head, subHead *ss.SSHead, miscReq *game_third.SSMiscReq) {
	hd, ok := this.handleMap[miscReq.MsgName]
	if !ok || hd == nil {
		SendMiscRetCode(head, subHead, miscReq, ss.EnmRetCode_MISC_NOT_SUPPORT_MSG)
		logger.Log.Error("not found message, subHead: %s, miscReq: %s", subHead.String(), miscReq.String())
		return
	}
	request := proto.Clone(hd.Template)

	err := protojson.Unmarshal([]byte(miscReq.MsgData), request)
	if err != nil {
		SendMiscRetCode(head, subHead, miscReq, ss.EnmRetCode_MISC_JSON_UNMARSHAL)
		logger.Log.Error("protojson.Unmarshal message, subHead: %s, miscReq: %s, err: %v", subHead.String(), miscReq.String(), err)
		return
	}

	md := &MiscData{
		Head:       head,
		SubHead:    subHead,
		MiscReq:    miscReq,
		Request:    request,
		ResMsgName: miscReq.MsgName[:len(miscReq.MsgName)-1] + "s",
	}

	hd.Handle(md)
}

func SendMiscRetCode(head, subHead *ss.SSHead, miscReq *game_third.SSMiscReq, code ss.EnmRetCode) {
	resHead := &ss.SSHead{
		CmdId: uint32(ss.EnmCmdValue_ECV_SSMiscRes),
		SeqId: subHead.SeqId,
		AccId: head.AccId,
	}

	resMsgName := miscReq.MsgName[:len(miscReq.MsgName)-1] + "s"
	miscRes := &game_third.SSMiscRes{
		MsgName: resMsgName,
	}

	if len(miscReq.AccList) > 0 {
		for _, accId := range miscReq.AccList {
			oneRes := &game_third.SSOneMiscRes{
				Code:  int32(code),
				AccId: accId,
			}
			miscRes.MiscResList = append(miscRes.MiscResList, oneRes)
		}
	} else {
		oneRes := &game_third.SSOneMiscRes{
			Code:    int32(code),
			ConnIdx: relay.RelayMgr.GetSelfConnIdx(),
		}
		miscRes.MiscResList = append(miscRes.MiscResList, oneRes)
	}
	relay.TransMsg(0, head.Src, head.AccId, resHead, miscRes)
}

func SendMisc(miscData *MiscData, resHead utils.IHeadMsg, miscRes utils.IMessage) {
	relay.TransAreaMsg(0, miscData.Head.Src, miscData.Head.AccId, 0, resHead, miscRes)
}
