package main

import (
	"control/scripts/framework"
	"fmt"
	"os"
	"strings"
)

func main() {
	argsCount := len(os.Args)

	if argsCount <= 1 {
		fmt.Printf("Usage: %s ip:port [cmd args]", os.Args[0])
	} else if argsCount == 2 {
		framework.Talk(os.Args[1])
	} else {
		// 执行一条命令
		cmd := []string{}
		for i := 2; i < argsCount; i++ {
			cmd = append(cmd, os.Args[i])
		}

		framework.ExecCmd(os.Args[1], strings.Join(cmd, " "))
	}
}
