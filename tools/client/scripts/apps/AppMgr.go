package apps

import (
	"log"
	cs "gamekit/pkg/cs"

	"golang.org/x/net/websocket"
)

type HandleClient func(conn *websocket.Conn, Head *cs.CSHead, iMsg cs.IMessage) uint32

var (
	handleClientMap map[cs.EnmCmdValue]HandleClient
	WSConn          *websocket.Conn
	AccId           uint64
	Name            string
	PlayerData      *cs.CSPlayerData
)

func init() {
	handleClientMap = make(map[cs.EnmCmdValue]HandleClient)
}

func RegistClientHandle(CmdId cs.EnmCmdValue, handler HandleClient) int {
	if _, ok := handleClientMap[CmdId]; ok {
		log.Printf("duplicate register cmd. CmdId: %d\n", CmdId)
		return -1
	}
	log.Printf("register cmd success. CmdId: %d\n", CmdId)
	handleClientMap[CmdId] = handler
	return 0
}

func GetClientHandle(CmdId cs.EnmCmdValue) HandleClient {
	if v, ok := handleClientMap[CmdId]; !ok {
		return nil
	} else {
		return v
	}
}
