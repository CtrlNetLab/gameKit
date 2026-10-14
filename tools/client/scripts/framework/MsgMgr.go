package framework

import (
	cs "gamekit/pkg/cs"
	"log"
	"net"
	"time"

	"github.com/panjf2000/gnet"
	"github.com/xxtea/xxtea-go/xxtea"

	"golang.org/x/net/websocket"
	// "github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var (
	maxBufferSize int    = 512 * 1024
	SeqId         uint32 = 1
	Codec         gnet.ICodec

	xxteaOpen bool   = false
	xxteaKey  []byte = []byte("AA9zg1CZMnYLpA5H")
)

func SendMsg(conn *net.TCPConn, CmdId cs.EnmCmdValue, message cs.IMessage) error {
	// func SendMsg(conn *websocket.Conn, CmdId cs.EnmCmdValue, message cs.IMessage) error {
	Head := &cs.CSHead{CmdId: uint32(CmdId), SeqId: SeqId, StrCmdId: CmdId.String()}
	SeqId++

	var err error
	headlen := proto.Size(Head)
	bodylen := proto.Size(message)

	msgBuffer := make([]byte, maxBufferSize)
	headBytes, err := proto.Marshal(Head)
	// if headlen, err = Head.MarshalTo(msgBuffer); err != nil {
	if err != nil {
		log.Printf("proto.Marshal failed. Head: %s, Err: %#v\n", Head.String(), err)
		return err
	}
	copy(msgBuffer, headBytes)

	boffset := headlen
	messageBytes, err := proto.Marshal(message)
	// if bodylen, err = message.MarshalTo(msgBuffer[boffset:]); err != nil {
	if err != nil {
		log.Printf("proto.Marshal failed. Message: %s, Err: %#v\n", message.String(), err)
		return err
	}
	copy(msgBuffer[boffset:], messageBytes)

	var xxteaBuffer []byte
	if xxteaOpen {
		xxteaBuffer = xxtea.Encrypt(msgBuffer[:bodylen+headlen], xxteaKey)
	} else {
		xxteaBuffer = msgBuffer[:bodylen+headlen]
	}

	// 暂时不加密、压缩
	totallen := 2 + len(xxteaBuffer)

	buffer := make([]byte, 0, totallen)

	if xxteaOpen {
		buffer = append(buffer, byte(1))
	} else {
		buffer = append(buffer, byte(0))
	}

	buffer = append(buffer, byte(headlen))
	buffer = append(buffer, xxteaBuffer...)

	//log.Printf("buffer.len: %d, msgBuffer.len: %d\n", len(buffer), len(xxteaBuffer))

	// @TODO ryder 多了一次拷贝，可以考虑不用进行拷贝，原地处理即可
	raw, err := Codec.Encode(nil, buffer)
	if err != nil {
		log.Printf("Codec.Encode failed. Message: %s, Err: %#v\n", message.String(), err)
		return err
	}

	n, err := conn.Write(raw)
	if n != len(raw) || err != nil {
		log.Printf("conn.Write failed. Message: %s, Err: %#v, n: %d, len(raw): %d\n", message.String(), err, n, len(raw))
		return err
	}

	log.Printf("SendMessage<<Head: %s, Message: %s\n", Head.String(), message.String())

	return nil
}

func SendMsgWeb(conn *websocket.Conn, CmdId cs.EnmCmdValue, message cs.IMessage) error {
	Head := &cs.CSHead{CmdId: uint32(CmdId), SeqId: SeqId, StrCmdId: CmdId.String(), SendTm: time.Now().String()[:26]}
	SeqId++

	var err error
	headlen := proto.Size(Head)
	bodylen := proto.Size(message)

	msgBuffer := make([]byte, maxBufferSize)
	headBytes, err := proto.Marshal(Head)
	// if headlen, err = Head.MarshalTo(msgBuffer); err != nil {
	if err != nil {
		log.Printf("proto.Marshal failed. Head: %s, Err: %#v\n", Head.String(), err)
		return err
	}
	copy(msgBuffer, headBytes)

	//log.Printf("message.size: %d, totallen: %d\n", message.Size(), len(msgBuffer))

	boffset := headlen
	messageBytes, err := proto.Marshal(message)
	// if bodylen, err = message.MarshalTo(msgBuffer[boffset:]); err != nil {
	if err != nil {
		log.Printf("proto.Marshal failed. Message: %s, Err: %#v\n", message.String(), err)
		return err
	}
	copy(msgBuffer[boffset:], messageBytes)

	var xxteaBuffer []byte
	if xxteaOpen {
		xxteaBuffer = xxtea.Encrypt(msgBuffer[:bodylen+headlen], xxteaKey)
	} else {
		xxteaBuffer = msgBuffer[:bodylen+headlen]
	}

	// 暂时不加密、压缩
	totallen := 2 + len(xxteaBuffer)

	buffer := make([]byte, 0, totallen)

	if xxteaOpen {
		buffer = append(buffer, byte(1))
	} else {
		buffer = append(buffer, byte(0))
	}

	buffer = append(buffer, byte(headlen))
	buffer = append(buffer, xxteaBuffer...)

	//log.Printf("buffer.len: %d, msgBuffer.len: %d\n", len(buffer), len(xxteaBuffer))

	// @TODO ryder 多了一次拷贝，可以考虑不用进行拷贝，原地处理即可
	raw, err := Codec.Encode(nil, buffer)
	if err != nil {
		log.Printf("Codec.Encode failed. Message: %s, Err: %#v\n", message.String(), err)
		return err
	}

	// conn.PayloadType = websocket.BinaryFrame
	// _, err = conn.Write(raw)
	err = websocket.Message.Send(conn, raw)
	// err = wsutil.WriteClientBinary(conn, raw)
	// err = conn.WriteMessage(websocket.BinaryMessage, raw)
	if err != nil {
		log.Printf("conn.Write failed. Message: %s, Err: %#v, len(raw): %d\n", message.String(), err, len(raw))
		return err
	}

	if CmdId != cs.EnmCmdValue_ECV_CSHeartReq {
		log.Printf("Head: %s, Message: %s\n", Head.String(), message.String())
	}

	return nil
}

/*
type mockConn struct {
	gnet.Conn
	buf []byte
}

func (c *mockConn) Read() []byte {
	return c.buf
}

func (c *mockConn) ShiftN(_ int) int {
	return 0
}

func RecvMsg(conn *net.TCPConn, Codec gnet.ICodec) {
	buffer := make([]byte, 8*1024)

	buflen, err := conn.Read(buffer)
	if err != nil {
		log.Printf("recv data failed. err: %#v\n", err)
		return
	}

	log.Printf("recv buffer len: %d\n", buflen)

	gc := &mockConn{buf: buffer[:buflen]}

	frame, _ := Codec.Decode(gc)

	var length uint32 = uint32(len(frame))
	if length <= 2 || length < (2+uint32(frame[1])) {
		log.Printf("[ERROR] length too small. length: %d, headlen: %d\n", length, uint32(frame[1]))
		// @TODO ryder error log
		return
	}

	// 加密 && 压缩
	// bit := uint32(frame[0])
	// 头部长度
	headlen := uint32(frame[1])
	Head := &cs.CSHead{}

	if err := Head.Unmarshal(frame[2 : 2+headlen]); err != nil {
		// @TODO ryder error log
		log.Printf("[ERROR] unpack head failed. length: %d, headlen: %d\n", length, headlen)
		return
	}

	imessage := cs.NewClientMessage(int32(Head.GetCmdId()))
	if imessage == nil {
		// @TODO ryder error log
		log.Printf("[ERROR] not found message. CmdId: %d\n", Head.GetCmdId())
		return
	}

	// message := imessage.(*cs.CSLoginReq)
	moffset := 2 + headlen + 1

	if merr := imessage.Unmarshal(frame[moffset:]); merr != nil {
		// @TODO ryder error log
		log.Printf("[ERROR] unpack message failed. Head: %s, Err: %#v\n", Head.String(), merr)
		return
	}

	log.Printf("Recv Message. Head: %s, Message: %s\n", Head.String(), imessage.String())

	handler := apps.GetClientHandle(int32(Head.GetCmdId()))

	if handler == nil {
		log.Printf("[ERROR] not found handler. Head: %s, Message: %s\n", Head.String(), imessage.String())
		return
	}

	handler(conn, Head, imessage)
}
*/
