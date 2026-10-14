package main

import (
	cs "gamekit/pkg/cs"
	"client/scripts/apps"
	"client/scripts/framework"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"client/config"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/panjf2000/gnet"
	"github.com/xxtea/xxtea-go/xxtea"

	"golang.org/x/net/websocket"
	"google.golang.org/protobuf/proto"
	// "github.com/gorilla/websocket"
)

var (
	defaultUserId string = "10000"

	userId = flag.String("u", defaultUserId, "userid")

	xxteaKey []byte = []byte("AA9zg1CZMnYLpA5H")
)

func Login(conn *websocket.Conn) {
	userId := *userId
	if userId == defaultUserId && config.Conf.UserId != defaultUserId {
		// 命令行是默认，配置了其他，使用配置的 UserId
		userId = config.Conf.UserId
	}
	UserId := fmt.Sprintf("%s", userId)
	Token := fmt.Sprintf("%s", userId)
	// Name := fmt.Sprintf("n-%s", userId)
	PlatType := "gp"
	LoginReq := &cs.CSLoginReq{UserId: UserId, Token: Token, PlatType: PlatType, ReLogin: false}
	log.Printf("LoginReq: %s\n", LoginReq.String())
	framework.SendMsgWeb(conn, cs.EnmCmdValue_ECV_CSLoginReq, LoginReq)
}

type mockConn struct {
	gnet.Conn
	buf []byte
}

func (c *mockConn) Read() []byte {
	return c.buf
}

func (c *mockConn) ShiftN(n int) int {
	c.buf = c.buf[n:]
	// log.Printf("ShiftN: %d, buflen: %d\n", n, len(c.buf))
	return 0
}

func RecvMsg(conn *websocket.Conn, frame []byte) error {
	var length uint32 = uint32(len(frame))
	if length <= 2 || length < (2+uint32(frame[1])) {
		log.Printf("[ERROR] length too small. length: %d, headlen: %d\n", length, uint32(frame[1]))
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

	//log.Printf("buffer.len: %d, frame.len: %d, bitmask: %d\n", len(buffer), len(frame), bitmask)

	// 头部长度
	headlen := uint32(frame[1])
	Head := &cs.CSHead{}
	err := proto.Unmarshal(buffer[:headlen], Head)
	// if err := Head.Unmarshal(buffer[:headlen]); err != nil {
	if err != nil {
		log.Printf("[ERROR] unpack head failed. length: %d, headlen: %d\n", length, headlen)
		return errors.New("unpack head failed")
	}

	imessage, ok := cs.NewClientMessage(int32(Head.GetCmdId()))
	if !ok || imessage == nil {
		log.Printf("[ERROR] not found message. CmdId: %d\n", Head.GetCmdId())
		return errors.New(fmt.Sprintf("not found message, CmdId: %d", Head.GetCmdId()))
	}

	moffset := headlen
	err = proto.Unmarshal(buffer[moffset:], imessage)
	// if merr := imessage.Unmarshal(buffer[moffset:]); merr != nil {
	if err != nil {
		log.Printf("[ERROR] unpack message failed. Head: %s, Err: %#v\n", Head.String(), err)
		return errors.New(fmt.Sprintf("unpack message failed, CmdId: %d", Head.GetCmdId()))
	}

	if Head.CmdId != uint32(cs.EnmCmdValue_ECV_CSHeartRes) {
		js, _ := new(jsonpb.Marshaler).MarshalToString(imessage)
		log.Printf("RecvMessage>>Head: %s, Message: %s\n", Head.String(), js)
	}

	handle := apps.GetClientHandle(cs.EnmCmdValue(Head.GetCmdId()))
	if handle == nil {
		handle = apps.DefaultHandler
	}

	handle(conn, Head, imessage)

	return nil
}

func startCodecClient(network, addr string, multicore, async bool) {
	rand.Seed(time.Now().UnixNano())

	taddr, _ := net.ResolveTCPAddr("tcp4", addr)
	var (
		err error
	)

	// apps.CurConn, err = net.DialTCP("tcp", nil, taddr)
	origin := "http://localhost/"
	// u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	apps.WSConn, err = websocket.Dial(fmt.Sprintf("ws://%s/ws", taddr), "", origin)
	// apps.WSConn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		panic(err)
	}

	Login(apps.WSConn)

	buffer := make([]byte, 128*1024)

	for {
		if apps.WSConn == nil {
			break
		}
		n, err := apps.WSConn.Read(buffer)
		if err != nil {
			log.Printf("recv websock data failed. err: %#v\n", err)
			break
		}
		// log.Printf("recv message. wsb: %v\n", buffer[:n])

		gc := &mockConn{buf: buffer[:n]}

		for {
			frame, ferr := framework.Codec.Decode(gc)
			if frame == nil || ferr != nil {
				// log.Printf("decode frame failed. err: %#v\n", ferr)
				break
			}

			if rerr := RecvMsg(apps.WSConn, frame); rerr != nil {
				log.Printf("RecvMsg err: %#v\n", rerr)
				break
			}
		}
	}
}

func main() {
	flag.Parse()
	fmt.Println(flag.Args())
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
	host := config.Conf.Host

	startCodecClient("tcp", host, true, true)
}
