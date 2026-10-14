package logger

import (
	"fmt"
	log "github.com/walkon/log4go"
	"net"
	"os"
	"os/user"
	"testing"
)

var (
	Log      *log.Filter
	OSS      *log.Filter
	Monitor  *log.Filter
	Play     *log.Filter
	LocalOSS *log.Filter
)

func init() {
	infile := "../conf/log4go.json"
	_, err := os.Stat(infile)
	if err != nil {
		infile = "./conf/log4go.json"
	}
	_, err = os.Stat(infile)
	if err != nil {
		if testing.Testing() {
			// 单元测试环境：降级为 stdout，不退出进程
			// 必须用业务 category 名注册，否则 LOGGER("MainLog") 找不到，会返回 CRITICAL 级别的空 filter
			fmt.Fprintf(os.Stderr, "logger: config file not found, fallback to stdout for testing\n")
			writer := log.NewConsoleLogWriter()
			for _, category := range []string{"MainLog", "MonitorLog", "OSSLog", "LocalOSS", "PlayLog"} {
				log.AddFilter(category, log.DEBUG, writer)
			}
		} else {
			// 生产环境：配置文件缺失属于致命错误，交由 log4go 退出
			log.LoadConfiguration(infile)
		}
	} else {
		log.LoadConfiguration(infile)
	}

	curSysUser := GetSysUser()
	curLocalIp := GetLocalIp()

	Log = initLoggerFilter("MainLog", curSysUser, curLocalIp)
	Monitor = initLoggerFilter("MonitorLog", curSysUser, curLocalIp)
	OSS = initLoggerFilter("OSSLog", curSysUser, curLocalIp)
	LocalOSS = initLoggerFilter("LocalOSS", curSysUser, curLocalIp)
	Play = initLoggerFilter("PlayLog", curSysUser, curLocalIp)

}

type GormWriter struct {
}

func (w GormWriter) Printf(fmt string, args ...interface{}) {
	Log.Warn(fmt, args...)
}

func initLoggerFilter(category string, curSysUser string, curLocalIp string) *log.Filter {
	logFilter := log.LOGGER(category)
	logFilter.Category = fmt.Sprintf("%s|%s@%s", logFilter.Category, curSysUser, curLocalIp)
	return logFilter
}

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
