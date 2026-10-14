package config

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Network struct {
	ServerService string `yaml:"ServerService"`
}

type Common struct {
	GameId          string `yaml:"GameId"`
	EnvName         string `yaml:"EnvName"`
	ConnTimeOut     int64  `yaml:"ConnTimeOut"`
	ServiceId       int32  `yaml:"ServiceId"`
	StartService    string `yaml:"StartService"`
	HeartTimeOut    int64  `yaml:"HeartTimeOut"`
	MaxHeartNum     int32  `yaml:"MaxHeartNum"`
	ReConnTm        int64  `yaml:"ReConnTm"`
	ReConnNum       int32  `yaml:"ReConnNum"`
	NetLog          bool   `yaml:"NetLog"`
	RoomHeartTime   int32  `yaml:"RoomHeartTime"`
	RoomOfflineTime int32  `yaml:"RoomOfflineTime"`
	RoomClearTime   int32  `yaml:"RoomClearTime"`
	MaxMatchNum     int    `yaml:"MaxMatchNum"`
	DiscCancelTime  int64  `yaml:"DiscCancelTime"`
}

type DivRule struct {
	InitDivRange    uint32   `yaml:"InitDivRange"`
	Timeout         int64    `yaml:"Timeout"`
	TimeoutDivRange []uint32 `yaml:"TimeoutDivRange"`
	OpenFixMatch    bool     `yaml:"OpenFixMatch"`
}

type Control struct {
	Service string `yaml:"Service"`
	ScrtKey string `yaml:"ScrtKey"`
}

type Config struct {
	Network Network `yaml:"Network"`
	Common  Common  `yaml:"Common"`
	Control Control `yaml:"Control"`
	DivRule DivRule `yaml:"DivRule"`
}

var Conf Config

var configFilePath string

func InitConfig(yamlPath string) {
	configFilePath = yamlPath
	LoadYaml(true)
}

func LoadYaml(start bool) {
	configFile := configFilePath
	if configFile == "" {
		configFile = cmdflag.GetConfigFilePath()
	}
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		if start {
			panic(fmt.Sprintf("ReadFile %s failed. Err: %#v", configFile, err))
		} else {
			logger.Log.Error("ReadFile %s failed. Err: %#v", configFile, err)
			return
		}
	}

	conf := Config{}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		if start {
			panic(fmt.Sprintf("yaml.Unmarshal %s failed. Err: %#v", configFile, err))
		} else {
			logger.Log.Error("yaml.Unmarshal %s failed. Err: %#v", configFile, err)
			return
		}
	}

	Conf = conf
	if !snowflake.Init(Conf.Common.StartService, int(Conf.Common.ServiceId)) {
		if start {
			panic(fmt.Sprintf("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId))
		} else {
			logger.Log.Error("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId)
			return
		}
	}

	logger.Log.Error("reload %s success, config: %#v", configFile, Conf)
}

func GetServiceType() ss.ServiceType {
	return ss.ServiceType_MatchSvr
}

func GetServiceId() int32 {
	return Conf.Common.ServiceId
}

func GetSelfConnIdx() uint64 {
	iType := ss.ServiceType_MatchSvr
	idx := GetServiceId()
	return (((uint64(iType) & 0xff) << 16) | (uint64(idx) & 0xffff))
}
