package cmdflag

import (
	"flag"
	"os"
)

var configFileOverride string
var configFileFlagProvided bool

func SetConfigFilePath(path string) {
	configFileOverride = path
}

func GetConfigFilePath() string {
	if configFileOverride != "" {
		return configFileOverride
	}
	// 保持常见命令行约定：显式传入的 -c 优先于环境变量。
	if configFileFlagProvided {
		return *ConfigFilePath
	}
	if v, ok := os.LookupEnv("AC_SERVICE_YAML"); ok && v != "" {
		return v
	}
	return *ConfigFilePath
}

var StartType = flag.String("s", "start", "启动方式，默认直接启动\n\t-s start: 启动\n\t-s stop: 退出\n\t-s control: 向当前进程发送消息")
var PidFile = flag.String("p", "./pid.pid", "进程 pid 文件")
var ConfigFilePath = flag.String("c", "../conf/service.yaml", "配置文件路径")

func ParseFlag() {
	flag.Parse()
	flag.CommandLine.Visit(func(f *flag.Flag) {
		if f.Name == "c" {
			configFileFlagProvided = true
		}
	})
}
