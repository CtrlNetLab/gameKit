package framework

import (
	"fmt"
	"net"

	"gamekit/pkg/game_third"
	ss "gamekit/pkg/ss"

	"github.com/panjf2000/gnet"
	"github.com/xxtea/xxtea-go/xxtea"
	"google.golang.org/protobuf/proto"
)

var (
	maxBufferSize int    = 512 * 1024
	SeqId         uint32 = 1
	Codec         gnet.ICodec

	xxteaOpen bool   = false
	xxteaKey  []byte = []byte("AA9zg1CZMnYLpA5H")
)

func SendMsg(conn *net.TCPConn, message *game_third.SSMiscReq) error {
	head := &ss.SSHead{CmdId: uint32(ss.EnmCmdValue_ECV_SSMiscReq), External: 1}

	var err error
	headlen := proto.Size(head)
	bodylen := proto.Size(message)

	msgBuffer := make([]byte, maxBufferSize)
	headBytes, err := proto.Marshal(head)
	// if headlen, err = head.MarshalTo(msgBuffer); err != nil {
	if err != nil {
		fmt.Printf("proto.Marshal failed. head: %s, Err: %#v\n", head.String(), err)
		return err
	}
	copy(msgBuffer, headBytes)

	fmt.Printf("message.size: %d, totallen: %d\n", bodylen, len(msgBuffer))

	boffset := headlen
	messageBytes, err := proto.Marshal(message)
	// if bodylen, err = message.MarshalTo(msgBuffer[boffset:]); err != nil {
	if err != nil {
		fmt.Printf("proto.Marshal failed. Message: %d, Err: %v\n", bodylen, err)
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

	fmt.Printf("buffer.len: %d, msgBuffer.len: %d\n", len(buffer), len(xxteaBuffer))

	// @TODO ryder 多了一次拷贝，可以考虑不用进行拷贝，原地处理即可
	raw, err := Codec.Encode(nil, buffer)
	if err != nil {
		fmt.Printf("Codec.Encode failed. Message: %s, Err: %#v\n", message.String(), err)
		return err
	}

	n, err := conn.Write(raw)
	if n != len(raw) || err != nil {
		fmt.Printf("conn.Write failed. Message: %s, Err: %#v, n: %d, len(raw): %d\n", message.String(), err, n, len(raw))
		return err
	}

	fmt.Printf("SendMsg success. head: %s, Message: %s\n", head.String(), message.String())

	return nil
}
