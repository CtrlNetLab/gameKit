package control

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	utils "gamekit/src/base/utils"

	log "github.com/walkon/log4go"
)

var (
	ControlService *controlService = &controlService{}
	Log            *log.Filter
)

type controlService struct {
	service string
	scrtKey string
}

func (this *controlService) Init(service, scrtKey string, logger *log.Filter) {
	this.service = service
	this.scrtKey = scrtKey
	Log = logger
	// this.Serve()
}

func (this *controlService) Serve() {
	ra, err := net.ResolveUDPAddr("udp", this.service)
	if err != nil {
		panic(fmt.Sprintf("init udp address failed. err: %#v\n", err))
	}

	conn, err := net.ListenUDP("udp", ra)
	if err != nil {
		panic(fmt.Sprintf("udp listen failed. err: %#v\n", err))
	}

	go func(c *net.UDPConn) {
		defer c.Close()

		var buffer [4096]byte
		maxOutConsole := 4096
		for {
			n, raddr, err := c.ReadFromUDP(buffer[0:])
			if err != nil {
				Log.Error("read udp failed. raddr: %v, err: %v", raddr, err)
				continue
			}

			in := string(buffer[:n])
			Log.Debug("recv buffer: %s", in)
			// decode64/decrypt

			in = utils.DeleteExtraSpace(in)
			args := strings.Split(in, " ")

			utils.WorkerPool.AddCallback(
				"controlService:Serve",
				func() {
					var (
						result string
						err    error
					)

					result, err = Handler(args, utils.CmdType_Control, nil)
					go func() {
						if err == nil && len(result) >= maxOutConsole {
							outfile := args[0] + "." + time.Now().Format("20060102150405") + ".txt"
							err = os.WriteFile(outfile, []byte(result), 0644)
							result = "result is too big, output file: " + outfile
						}
						_, uerr := c.WriteToUDP([]byte(fmt.Sprintf("result: \n%s\nerr: %v\n", result, err)), raddr)
						if uerr != nil {
							Log.Error("response control failed. uerr: %#v, Args: %#v, Result: %s, err: %v", uerr, args, result, err)
						}
					}()
				},
			)
		}
	}(conn)
}
