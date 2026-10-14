package relay

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type RelayCommon struct {
	RelayHeartTm int64 `yaml:"RelayHeartTm"`
	HeartTimeOut int64 `yaml:"HeartTimeOut"`
	MaxHeartNum  int32 `yaml:"MaxHeartNum"`
	ReConnTm     int64 `yaml:"ReConnTm"`
	ReConnNum    int32 `yaml:"ReConnNum"`
	DelayTimeOut int64 `yaml:"DelayTimeOut"`
}

type relayConf struct {
	RelaySvr map[string][]string `yaml:"RelaySvr"`
	Common   RelayCommon         `yaml:"Common"`
}

var RelayConf relayConf

func InitRelayConfig() {
	configFile := "../../binconf/common/relay.yaml"

	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("ReadFile %s failed. Err: %#v", configFile, err))
	}

	RelayConf = relayConf{}

	err = yaml.Unmarshal(yamlFile, &RelayConf)
	if err != nil {
		panic(fmt.Sprintf("yaml.Unmarshal %s failed. Err: %#v", configFile, err))
	}

	Log.Error("init %s success, config: %#v", configFile, RelayConf)
}
