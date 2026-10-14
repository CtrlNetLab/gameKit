package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"time"

	"gamekit/pkg/game_third"
	ss "gamekit/pkg/ss"
	"misc/scripts/framework"

	"github.com/panjf2000/gnet"
	"github.com/xxtea/xxtea-go/xxtea"
	"google.golang.org/protobuf/proto"
)

var userId = flag.String("u", "10000", "userid")

var xxteaKey []byte = []byte("AA9zg1CZMnYLpA5H")

type mockConn struct {
	gnet.Conn
	buf []byte
}

func (c *mockConn) Read() []byte {
	return c.buf
}

func (c *mockConn) ShiftN(n int) int {
	c.buf = c.buf[n:]
	// fmt.Printf("ShiftN: %d, buflen: %d\n", n, len(c.buf))
	return 0
}

func RecvMsg(conn *net.TCPConn, frame []byte) error {
	var length uint32 = uint32(len(frame))
	if length <= 2 || length < (2+uint32(frame[1])) {
		fmt.Printf("[ERROR] length too small. length: %d, headlen: %d\n", length, uint32(frame[1]))
		// @TODO ryder error log
		return errors.New("invalid message format")
	}

	// 加密 && 压缩
	bitmask := uint32(frame[0])
	var buffer []byte
	if (bitmask & 0xf) != 0 {
		buffer = xxtea.Decrypt(frame[2:], xxteaKey)
	} else {
		buffer = frame[2:]
	}

	fmt.Printf("buffer.len: %d, frame.len: %d, bitmask: %d\n", len(buffer), len(frame), bitmask)

	// 头部长度
	headlen := uint32(frame[1])
	Head := &ss.SSHead{}

	err := proto.Unmarshal(buffer[:headlen], Head)
	// if err := Head.Unmarshal(buffer[:headlen]); err != nil {
	if err != nil {
		fmt.Printf("[ERROR] unpack head failed. length: %d, headlen: %d\n", length, headlen)
		return errors.New("unpack head failed")
	}

	imessage, _ := ss.NewServiceMessage(int32(Head.GetCmdId()))
	if imessage == nil {
		fmt.Printf("[ERROR] not found message. CmdId: %d\n", Head.GetCmdId())
		return fmt.Errorf("not found message, CmdId: %d", Head.GetCmdId())
	}

	moffset := headlen

	err = proto.Unmarshal(buffer[moffset:], imessage)
	// if merr := imessage.Unmarshal(buffer[moffset:]); merr != nil {
	if err != nil {
		fmt.Printf("[ERROR] unpack message failed. Head: %s, Err: %#v\n", Head.String(), err)
		return fmt.Errorf("unpack message failed, CmdId: %d", Head.GetCmdId())
	}

	fmt.Printf("Recv Message. Head: %s, Message: %s\n", Head.String(), imessage.String())

	return nil
}

func startCodecMisc(network, addr string, multicore, async bool) {
	rand.Seed(time.Now().UnixNano())

	taddr, _ := net.ResolveTCPAddr("tcp4", addr)
	conn, err := net.DialTCP("tcp", nil, taddr)

	if err != nil {
		panic(err)
	}

	defer func() {
		conn.Close()
		time.Sleep(10 * time.Second)
	}()

	/*
		getPlayerReq := &game_third.SPGetPlayerDataReq{}
		reqData, rerr := json.Marshal(getPlayerReq)
		if rerr != nil {
			fmt.Printf("json marshal failed. Err: %v\n", rerr)
			return
		}
		miscReq := &game_third.SSMiscReq{
			AccList: []uint64{648994},
			// @NOTE 修改这条协议名字
			MsgName: "SPGetPlayerDataReq",
			MsgData: string(reqData),
		}
	*/

	var reqData string = "{}"
	miscReq := &game_third.SSMiscReq{
		AccList: []uint64{648994},
		// @NOTE 修改这条协议名字
		MsgName: "SPGetPlayerDataReq",
		MsgData: string(reqData),
	}
	framework.SendMsg(conn, miscReq)

	buffer := make([]byte, 128*1024)

	for {
		buflen, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("recv data failed. err: %#v\n", err)
			break
		}

		fmt.Printf("recv buffer len: %d\n", buflen)

		gc := &mockConn{buf: buffer[:buflen]}

		for {
			frame, ferr := framework.Codec.Decode(gc)
			if frame == nil || ferr != nil {
				// fmt.Printf("decode frame failed. err: %#v\n", ferr)
				break
			}
			rerr := RecvMsg(conn, frame)
			if rerr != nil {
				fmt.Printf("recv message failed. Err: %v\n", rerr)
			}
			return
		}
	}
}

func main() {
	flag.Parse()

	encoderConfig := gnet.EncoderConfig{
		ByteOrder:                       binary.BigEndian,
		LengthFieldLength:               4,
		LengthAdjustment:                0,
		LengthIncludesLengthFieldLength: false,
	}

	decoderConfig := gnet.DecoderConfig{
		ByteOrder:           binary.BigEndian,
		LengthFieldOffset:   0,
		LengthFieldLength:   4,
		LengthAdjustment:    0,
		InitialBytesToStrip: 4,
	}

	framework.Codec = gnet.NewLengthFieldBasedFrameCodec(encoderConfig, decoderConfig)

	startCodecMisc("tcp", "127.0.0.1:7302", true, true)
}
