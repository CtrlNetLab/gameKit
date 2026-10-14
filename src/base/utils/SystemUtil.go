package utils

import (
	"gamekit/src/base/logger"
	"net"
	"os/user"
	"runtime"
)

/**
* @author: Tavin
* @date: 2024/5/21
 */

// GetSysUser 获得系统用户
func GetSysUser() string {
	user, _ := user.Current()
	return user.Username
}

// GetLocalIp 获取当前第一个获取的局域网IP
func GetLocalIp() string {
	addrList, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrList {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil && ipNet.IP.IsGlobalUnicast() {
				return ipNet.IP.String()
			}
		}
	}
	return ""
}

func PrintStack() {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	logger.Log.Error("DebugStack: %s", buf[:n])
}
