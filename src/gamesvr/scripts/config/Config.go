package config

import (
	ss "gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	rdao "gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"gamekit/src/base/utils"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Network struct {
	ClientService string `yaml:"ClientService"`
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

type RedisKey struct {
	OnlineKey       string `yaml:"OnlineKey"`
	RoleNameKey     string `yaml:"RoleNameKey"`
	RoleInfoKey     string `yaml:"RoleInfoKey"`
	OfflineDataKey  string `yaml:"OfflineDataKey"`
	PlayerDBDataKey string `yaml:"PlayerDBDataKey"`
}

type Common struct {
	GameId              string  `yaml:"GameId"`
	EnvName             string  `yaml:"EnvName"`
	ConnTimeOut         int64   `yaml:"ConnTimeOut"`
	UpdateDBTm          []int64 `yaml:"UpdateDBTm"`
	GmOpenFlag          int     `yaml:"GmOpenFlag"`
	ServiceId           int32   `yaml:"ServiceId"`
	XXTeaOpen           bool    `yaml:"XXTeaOpen"`
	XXTeaKey            string  `yaml:"XXTeaKey"`
	PressOpen           bool    `yaml:"PressOpen"`
	StartService        string  `yaml:"StartService"`
	HealthPort          int     `yaml:"HealthPort"`
	HealthMaxNum        int     `yaml:"HealthMaxNum"`
	HealthCheckInterval int64   `yaml:"HealthCheckInterval"`
	NetLog              bool    `yaml:"NetLog"`
	RoleNameExTm        int32   `yaml:"RoleNameExTm"`
	LimitIpOpen         bool    `yaml:"LimitIpOpen"`
	LimitIpMax          int32   `yaml:"LimitIpMax"`
	LimitLoginOpen      bool    `yaml:"LimitLoginOpen"`
	LimitTimeInterval   []int64 `yaml:"LimitTimeInterval"`
	LimitTimeStage      []int64 `yaml:"LimitTimeStage"`
	UseRedisCache       bool    `yaml:"UseRedisCache"`
	WorkerPoolSize      int     `yaml:"WorkerPoolSize"`
	DelayCloseConnTime  int64   `yaml:"DelayCloseConnTime"`
}

type Control struct {
	Service     string `yaml:"Service"`
	ScrtKey     string `yaml:"ScrtKey"`
	CtrlDocPort int    `yaml:"CtrlDocPort"`
}

type Config struct {
	Network    Network                `yaml:"Network"`
	MySQLMap   map[string]MySQL       `yaml:"MySQL"`
	RedisMap   map[string]*rdao.Redis `yaml:"Redis"`
	RedisKey   RedisKey               `yaml:"RedisKey"`
	Common     Common                 `yaml:"Common"`
	Control    Control                `yaml:"Control"`
	HttpClient utils.HttpClient       `yaml:"HttpClient"`
	Locationing Locationing           `yaml:"Locationing"`
}

type Locationing struct {
	DefaultLocation string `yaml:"DefaultLocation"`
	LocaltionEnum   string
}

var (
	Conf           Config
	UseGm          bool
	LimitLoginTime = [3]int64{0, 0, 0}
)

var configFilePath string

func InitConfig(yamlPath string) {
	configFilePath = yamlPath
	LoadYaml(true)
}

// LoadYaml 重新加载配置文件
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
	conf.Locationing.LocaltionEnum = fmt.Sprintf("EL_%s", conf.Locationing.DefaultLocation)

	Conf = conf
	if !snowflake.Init(Conf.Common.StartService, int(Conf.Common.ServiceId)) {
		if start {
			panic(fmt.Sprintf("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId))
		} else {
			logger.Log.Error("init snowflake failed. StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId)
			return
		}
	}

	logger.Log.Error("init %s success, config: %#v", configFile, Conf)
}

func GetGmOpenFlag() int {
	return Conf.Common.GmOpenFlag
}

func IsProductGm() bool {
	return Conf.Common.GmOpenFlag == 1
}

func IsDevelopGm() bool {
	return Conf.Common.GmOpenFlag == 2
}

func GetSelfConnIdx() uint64 {
	iType := ss.ServiceType_GameSvr
	idx := GetServiceId()
	return (((uint64(iType) & 0xff) << 16) | (uint64(idx) & 0xffff))
}

func GetServiceType() ss.ServiceType {
	return ss.ServiceType_GameSvr
}

func GetServiceId() int32 {
	return Conf.Common.ServiceId
}

func IsAdminService() bool {
	return Conf.Common.ServiceId == 1
}

func GetOnlineKey() string {
	return Conf.RedisKey.OnlineKey
}

func GetRoleNameKey() string {
	return Conf.RedisKey.RoleNameKey
}

func GetRoleInfoKey() string {
	return Conf.RedisKey.RoleInfoKey
}

func GetOfflineDataKey() string {
	return Conf.RedisKey.OfflineDataKey
}

func GetPlayerDBDataKey() string {
	return Conf.RedisKey.PlayerDBDataKey
}

func IsLimitIpOpen() bool {
	return Conf.Common.LimitIpOpen
}

func GetLimitIpMax() int32 {
	return Conf.Common.LimitIpMax
}

func GetAreaId() int32 {
	return ss.EnmLocation_value[Conf.Locationing.LocaltionEnum]
}

func GetNextGameOpenTime(tm int64) int64 {
	if UseGm {
		return LimitLoginTime[2]
	}

	if !Conf.Common.LimitLoginOpen {
		return 0
	}

	if len(Conf.Common.LimitTimeInterval) != 2 {
		return 0
	}

	if tm < Conf.Common.LimitTimeInterval[0] {
		return Conf.Common.LimitTimeInterval[0]
	}

	if tm >= Conf.Common.LimitTimeInterval[1] {
		return 0
	}

	length := len(Conf.Common.LimitTimeStage)
	for i := 0; i+1 < length; i += 2 {
		ltm := Conf.Common.LimitTimeStage[i]
		if tm < ltm {
			return ltm
		}
	}
	return 0
}

func IsLimitLoginTime(tm int64) bool {
	if UseGm {
		if LimitLoginTime[0] <= tm && LimitLoginTime[1] > tm {
			return false
		}

		return true
	}

	if !Conf.Common.LimitLoginOpen {
		return false
	}

	if len(Conf.Common.LimitTimeInterval) != 2 {
		return false
	}

	if tm < Conf.Common.LimitTimeInterval[0] {
		return true
	}

	if tm >= Conf.Common.LimitTimeInterval[1] {
		return false
	}

	// 必须为偶数
	length := len(Conf.Common.LimitTimeStage)
	if (length % 2) != 0 {
		return false
	}

	for i := 0; i+1 < length; i += 2 {
		ltm := Conf.Common.LimitTimeStage[i]
		offset := Conf.Common.LimitTimeStage[i+1]
		rtm := ltm + offset*int64(60)
		if ltm <= tm && rtm > tm {
			return false
		}
	}

	return true
}
