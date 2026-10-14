package framework

import (
	"fmt"
	"net"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	oldTermState *terminal.State
)

const (
	Prompt string = "> "
)

type TalkData struct {
	Addr   *net.UDPAddr
	Args   []string
	Result string
}

// 删除首位空格、以及中间多于 2 个的空格或者\t
func DeleteExtraSpace(s string) string {
	var out []byte
	var space int = 0
	for i := 0; i < len(s); i++ {
		v := s[i]
		if v != ' ' && v != '\t' {
			if len(out) > 0 && space > 0 {
				out = append(out, ' ')
			}
			out = append(out, v)
			space = 0
		} else {
			space++
		}
	}
	return string(out)
}

func ExecCmd(service string, cmd string) {
	conn, err := net.Dial("udp", service)

	if err != nil {
		fmt.Printf("connect udp failed. Service: %s, err: %#v\n", service, err)
		return
	}

	defer conn.Close()

	cmd = DeleteExtraSpace(cmd)

	_, err = conn.Write([]byte(cmd))
	if err != nil {
		fmt.Printf("send message failed. Err: %#v\n", err)
		return
	}

	buffer := make([]byte, 32*1024)

	n, err := conn.Read(buffer)

	if err != nil {
		fmt.Printf("recv message failed. Err: %v\n", err)
		return
	}

	out := string(buffer[:n])

	fmt.Printf("%s", out)
}

func Talk(service string) {
	conn, err := net.Dial("udp", service)

	if err != nil {
		fmt.Printf("connect udp failed. Service: %s, err: %#v\n", service, err)
		return
	}

	defer conn.Close()

	buffer := make([]byte, 32*1024)

	oldTermState, err = terminal.MakeRaw(syscall.Stdin)
	if err != nil {
		fmt.Println(err)
		return
	}

	term := terminal.NewTerminal(os.Stdin, Prompt)
	term.SetSize(1024, 768)
	rawState, err := terminal.GetState(syscall.Stdin)
	if err != nil {
		fmt.Println("termainal getstat failed:  %v", err)
		return
	}

	for {
		cmdline, err := term.ReadLine()
		if err != nil {
			break
		}

		// 去掉首位空格
		cmdline = DeleteExtraSpace(cmdline)
		if cmdline == "" {
			continue
		}

		if cmdline == "quit" {
			break
		}

		terminal.Restore(syscall.Stdin, oldTermState)
		_, err = conn.Write([]byte(cmdline))
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
		terminal.Restore(syscall.Stdin, rawState)
	}
	terminal.Restore(syscall.Stdin, oldTermState)
}
