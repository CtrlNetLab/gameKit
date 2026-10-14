package apps

import (
	cs "gamekit/pkg/cs"
	"client/scripts/framework"
	"log"
	"time"

	"golang.org/x/net/websocket"
	// "time"
)

func init() {
	RegistClientHandle(cs.EnmCmdValue_ECV_CSLoginRes, handleLoginRes)
}

func handleLoginRes(conn *websocket.Conn, Head *cs.CSHead, iMsg cs.IMessage) uint32 {
	loginRes := iMsg.(*cs.CSLoginRes)

	log.Printf("Head: %s, handleLoginRes: %s\n", Head.String(), loginRes.String())
	if Head.RetCode != 0 || loginRes.RetCode != 0 {
		log.Printf("login failed. RetCode: %d, %d", Head.RetCode, loginRes.RetCode)
		return loginRes.RetCode
	}

	// AccId = loginRes.PlayerData.AccId
	// Name = loginRes.PlayerData.Name

	time.Sleep(time.Millisecond * 20)

	// 定时发心跳包
	testHeart(conn)

	return loginRes.RetCode
}

func DefaultHandler(conn *websocket.Conn, head *cs.CSHead, iMsg cs.IMessage) uint32 {
	//log.Printf("Head: %s, Msg: %s\n", proto.MarshalTextString(head), proto.MarshalTextString(iMsg))
	return 0
}

func testHeart(conn *websocket.Conn) {
	go func() {
		for {
			// send heart req
			framework.SendMsgWeb(conn, cs.EnmCmdValue_ECV_CSHeartReq, &cs.CSHeartReq{})
			time.Sleep(5 * time.Second)
		}
	}()
}
