package control

import (
	"bufio"
	"fmt"
	"net"
	"os"

	utils "gamekit/src/base/utils"
)

func Talk() {
	conn, err := net.Dial("udp", ControlService.service)

	if err != nil {
		Log.Error("connect udp failed. Service: %s, err: %#v", ControlService.service, err)
		return
	}

	defer conn.Close()

	buffer := make([]byte, 32*1024)

	fmt.Printf("> ")
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Text()
		// 去掉首位空格
		line = utils.DeleteExtraSpace(line)
		if line == "quit" {
			break
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("send message failed. Err: %#v\n", err)
			break
		}

		n, err := conn.Read(buffer)

		if err != nil {
			fmt.Printf("recv message failed. Err: %#v\n", err)
			break
		}

		out := string(buffer[:n])

		fmt.Printf("%s\n", out)

		fmt.Printf("> ")
	}
}

func TalkCmd(args string) {
	args = utils.DeleteExtraSpace(args)

	conn, err := net.Dial("udp", ControlService.service)

	if err != nil {
		Log.Error("connect udp failed. Service: %s, err: %#v", ControlService.service, err)
		return
	}

	defer conn.Close()

	buffer := make([]byte, 32*1024)

	_, err = conn.Write([]byte(args))
	if err != nil {
		fmt.Printf("send message failed. Err: %#v\n", err)
		return
	}

	n, err := conn.Read(buffer)

	if err != nil {
		fmt.Printf("recv message failed. Err: %#v\n", err)
		return
	}

	out := string(buffer[:n])

	fmt.Printf("%s\n", out)
}
