package config

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Network struct {
	ServerService string `yaml:"ServerService"`
}

type Locationing struct {
	DefaultLocation string `yaml:"DefaultLocation"`
	LocationEnum    string
}

type Common struct {
	GameId       string `yaml:"GameId"`
	EnvName      string `yaml:"EnvName"`
	StartService string `yaml:"StartService"`
	ServiceId    int32  `yaml:"ServiceId"`
	NetLog       bool   `yaml:"NetLog"`
	HealthPort   int    `yaml:"HealthPort"`
}

type RedisKey struct {
	OnlineKey string `yaml:"OnlineKey"`
}

type MiscConf struct {
	MaxBatCount    int           `yaml:"MaxBatCount"`
	TimeOut        time.Duration `yaml:"TimeOut"`
	GetServiceTime time.Duration `yaml:"GetServiceTime"`
}

type Control struct {
	Service string `yaml:"Service"`
	ScrtKey string `yaml:"ScrtKey"`
}

type Redis struct {
	AddrList []string          `yaml:"AddrList"`
	Passwd   string            `yaml:"Passwd"`
	DB       int               `yaml:"DB"`
	Scripts  map[string]string `yaml:"Scripts"`
}

type Config struct {
	Network     Network               `yaml:"Network"`
	Common      Common                `yaml:"Common"`
	Control     Control               `yaml:"Control"`
	RedisMap    map[string]*dao.Redis `yaml:"Redis"`
	RedisKey    RedisKey              `yaml:"RedisKey"`
	Misc        MiscConf              `yaml:"Misc"`
	Locationing Locationing           `yaml:"Locationing"`
}

var Conf Config

var configFilePath string

func InitConfig(yamlPath string) {
	configFilePath = yamlPath
	loadConfig()
}

func loadConfig() {
	configFile := configFilePath
	if configFile == "" {
		configFile = cmdflag.GetConfigFilePath()
	}
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("ReadFile %s failed. Err: %#v", configFile, err))
	}

	Conf = Config{}

	err = yaml.Unmarshal(yamlFile, &Conf)
	if err != nil {
		panic(fmt.Sprintf("yaml.Unmarshal %s failed. Err: %#v", configFile, err))
	}

	Conf.Locationing.LocationEnum = fmt.Sprintf("EL_%s", Conf.Locationing.DefaultLocation)
	if !snowflake.Init(Conf.Common.StartService, int(Conf.Common.ServiceId)) {
		panic(fmt.Sprintf("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId))
	}

	logger.Log.Error("init %s success, config: %#v", configFile, Conf)
}

// LoadYaml 热重载配置(GM reloadConfig 命令调用)
func LoadYaml(start bool) {
	loadConfig()
}

func GetOnlineKey() string {
	return Conf.RedisKey.OnlineKey
}

func GetServiceId() int32 {
	return Conf.Common.ServiceId
}

func GetAreaId() int32 {
	return ss.EnmLocation_value[Conf.Locationing.LocationEnum]
}
