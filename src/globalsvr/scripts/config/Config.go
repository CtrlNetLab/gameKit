package config

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	rdao "gamekit/src/base/dao"
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

type MySQL struct {
	HostList        []string `yaml:"HostList"`
	DBName          string   `yaml:"DBName"`
	SplitNum        int      `yaml:"SplitNum"`
	TablePrefix     string   `yaml:"TablePrefix"`
	Port            int      `yaml:"Port"`
	User            string   `yaml:"User"`
	Passwd          string   `yaml:"Passwd"`
	CharSet         string   `yaml:"CharSet"`
	Debug           bool     `yaml:"Debug"`
	MaxIdleConns    int      `yaml:"MaxIdleConns"`
	MaxOpenConns    int      `yaml:"MaxOpenConns"`
	ConnMaxLifetime int      `yaml:"ConnMaxLifetime"`
}

type Redis struct {
	AddrList []string          `yaml:"AddrList"`
	Passwd   string            `yaml:"Passwd"`
	DB       int               `yaml:"DB"`
	Scripts  map[string]string `yaml:"Scripts"`
}

type RedisKey struct {
	OnlineKey string `yaml:"OnlineKey"`
}

type Common struct {
	GameId          string        `yaml:"GameId"`
	EnvName         string        `yaml:"EnvName"`
	ServiceId       int32         `yaml:"ServiceId"`
	StartService    string        `yaml:"StartService"`
	NetLog          bool          `yaml:"NetLog"`
	SelectCount     int           `yaml:"SelectCount"`
	ThreadCount     int           `yaml:"ThreadCount"`
	PageCount       int           `yaml:"PageCount"`
	OnlineKey       string        `yaml:"OnlineKey"`
	DBUpdateTime    time.Duration `yaml:"DBUpdateTime"`
	FriendCrossFlag bool          `yaml:"FriendCrossFlag"`
}

type Control struct {
	Service string `yaml:"Service"`
	ScrtKey string `yaml:"ScrtKey"`
}

type Config struct {
	Network  Network                `yaml:"Network"`
	Control  Control                `yaml:"Control"`
	MySQLMap map[string]MySQL       `yaml:"MySQL"`
	RedisMap map[string]*rdao.Redis `yaml:"Redis"`
	RedisKey RedisKey               `yaml:"RedisKey"`
	Common   Common                 `yaml:"Common"`
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
		panic(fmt.Sprintf("ReadFile %s failed. Err: %#v", configFile, err))
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
	logger.Log.Error("init %s success, config: %+v", configFile, Conf)
}

func GetServiceType() ss.ServiceType {
	return ss.ServiceType_GlobalSvr
}

func GetServiceId() int32 {
	return Conf.Common.ServiceId
}

func GetSelfConnIdx() uint64 {
	return uint64(ss.ServiceType_GlobalSvr)
}

func GetTableCount(name string) int {
	my, ok := Conf.MySQLMap[name]
	if !ok {
		return 64
	}
	return my.SplitNum
}

func GetOnlineKey() string {
	return Conf.RedisKey.OnlineKey
}
