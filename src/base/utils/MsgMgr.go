package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cs "gamekit/pkg/cs"
	ss "gamekit/pkg/ss"

	gnet "github.com/walkon/wsgnet"

	"google.golang.org/protobuf/proto"
)

var (
	maxBufferSize int = 512 * 1024

	overloadSize error = fmt.Errorf("overload buffer size: %d", maxBufferSize)
	nilConn      error = errors.New("conn is nil or not int64")
)

func GetCSHead(cmdId cs.EnmCmdValue, seqId uint32, retCode cs.EnmRetCode, retMessage string) *cs.CSHead {
	if IsNetLog() {
		return &cs.CSHead{CmdId: uint32(cmdId), SeqId: seqId,
			RetCode: uint32(retCode), SendTm: time.Now().String()[:26], StrCmdId: cmdId.String(), RetMessage: retMessage}
	}
	return &cs.CSHead{CmdId: uint32(cmdId), SeqId: seqId, RetCode: uint32(retCode), RetMessage: retMessage}
}

func ConnAsyncWrite(conn gnet.Conn, buffer []byte) error {
	in, err := codec.Encode(conn, buffer)
	if err != nil {
		return err
	}
	if !conn.IsWebSock() {
		return conn.AsyncWrite(in, nil)
	}
	wsIn, err := WsServEncode(in)
	if err != nil {
		return err
	}
	return conn.AsyncWrite(wsIn, nil)
}

func NewSSHead(cmdId ss.EnmCmdValue, seqId uint32, retCode ss.EnmRetCode) *ss.SSHead {
	return &ss.SSHead{CmdId: uint32(cmdId), SeqId: seqId, RetCode: uint32(retCode)}
}

func PackMsg(head IHeadMsg, message IMessage) ([]byte, error) {
	// headlen := head.Size()
	headlen := proto.Size(head)

	// totallen := 2 + headlen + message.Size()
	totallen := 2 + headlen + proto.Size(message)

	buffer := make([]byte, totallen, totallen)

	buffer[0] = byte(0)
	buffer[1] = byte(headlen)

	headBytes, err := proto.Marshal(head)
	// if _, err := head.MarshalTo(buffer[2:]); err != nil {
	if err != nil {
		return buffer, err
	}
	copy(buffer[2:], headBytes)

	boffset := 2 + headlen
	messageBytes, err := proto.Marshal(message)
	// if _, err := message.MarshalTo(buffer[boffset:]); err != nil {
	if err != nil {
		return buffer, err
	}
	copy(buffer[boffset:], messageBytes)

	return buffer, nil
}

func PackJsonMsg(head IHeadMsg, message IMessage) ([]byte, error) {
	headBytes, err := json.Marshal(head)
	if err != nil {
		return nil, err
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	headlen := len(headBytes)
	totallen := 2 + headlen + len(messageBytes)
	buffer := make([]byte, totallen, totallen)

	buffer[0] = byte(0)
	buffer[1] = byte(headlen)

	copy(buffer[2:], headBytes)
	boffset := 2 + headlen
	copy(buffer[boffset:], messageBytes)

	return buffer, nil
}

func SendJSMsg(conn gnet.Conn, head IHeadMsg, message IMessage) error {
	connIdx, ok := conn.Context().(uint64)
	if !ok {
		return nilConn
	}

	buffer, err := PackJsonMsg(head, message)
	if err != nil {
		Log.Error("%s|PackMsg failed. Head: %s, Msg: %s, Err: %#v", connIdx, head.String(), message.String(), err)
		return err
	}

	if err := ConnAsyncWrite(conn, buffer); err != nil {
		Log.Error("%s|conn.Write failed. Head: %s, Message: %s, Err: %#v, len(buffer): %d",
			connIdx, head.String(), message.String(), err, len(buffer))
		return err
	}

	if netLog && LogFlag(head.GetCmdId()) {
		Log.Debug("%d|SendMsg success. Head: %s, Message: %s", connIdx, head.String(), message.String())
	}

	return nil
}

func SendMsg(conn gnet.Conn, head IHeadMsg, message IMessage) error {
	connIdx, ok := conn.Context().(uint64)
	if !ok {
		Log.Error("conn Context to uint64 failed. Context: %#v, head: %s, msg: %s", conn.Context(), head.String(), message.String())
		return nilConn
	}

	buffer, err := PackMsg(head, message)
	if err != nil {
		Log.Error("%d|PackMsg failed. Head: %s, Msg: %s, Err: %#v", connIdx, head.String(), message.String(), err)
		return err
	}

	if err := ConnAsyncWrite(conn, buffer); err != nil {
		Log.Error("%d|conn.Write failed. Head: %s, Message: %s, Err: %#v, len(buffer): %d",
			connIdx, head.String(), message.String(), err, len(buffer))
		return err
	}

	if netLog && LogFlag(head.GetCmdId()) {
		Log.Debug("%d|SendMsg success. PackageSize: %d, Head: %s, Message: %s", connIdx, len(buffer), head.String(), message.String())
	}

	return nil
}

// @TODO ryder: 不输出日志
var noLogMap = map[ss.EnmCmdValue]bool{
	ss.EnmCmdValue_ECV_SSHeartReq:         true,
	ss.EnmCmdValue_ECV_SSHeartRes:         true,
	ss.EnmCmdValue_ECV_SSGetAllServiceReq: true,
	ss.EnmCmdValue_ECV_SSGetAllServiceRes: true,
}

func LogFlag(cmdId uint32) bool {
	_, ok := noLogMap[ss.EnmCmdValue(cmdId)]
	return !ok
}
