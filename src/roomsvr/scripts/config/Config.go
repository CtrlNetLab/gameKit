package config

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	rdao "gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Network struct {
	ServerService string `yaml:"ServerService"`
}

type Locationing struct {
	DefaultLocation string `yaml:"DefaultLocation"`
	LocaltionEnum   string
}

type Common struct {
	GameId           string   `yaml:"GameId"`
	EnvName          string   `yaml:"EnvName"`
	ConnTimeOut      int64    `yaml:"ConnTimeOut"`
	ServiceType      string   `yaml:"ServiceType"`
	ServiceId        int32    `yaml:"ServiceId"`
	GmOpenFlag       int      `yaml:"GmOpenFlag"`
	SendBalanceTm    int64    `yaml:"SendBalanceTm"`
	ReConnTm         int64    `yaml:"ReConnTm"`
	ReConnNum        int32    `yaml:"ReConnNum"`
	MaxRoomNum       int32    `yaml:"MaxRoomNum"`
	XXTeaOpen        bool     `yaml:"XXTeaOpen"`
	XXTeaKey         string   `yaml:"XXTeaKey"`
	StartService     string   `yaml:"StartService"`
	NetLog           bool     `yaml:"NetLog"`
	HeartTimeOut     int      `yaml:"HeartTimeOut"`
	RoomHeartTimeOut int64    `yaml:"RoomHeartTimeOut"`
	ArenaList        []uint32 `yaml:"ArenaList"`
}

type Control struct {
	Service string `yaml:"Service"`
	ScrtKey string `yaml:"ScrtKey"`
}

type Config struct {
	Network     Network                `yaml:"Network"`
	Common      Common                 `yaml:"Common"`
	Control     Control                `yaml:"Control"`
	RedisMap    map[string]*rdao.Redis `yaml:"Redis"`
	Locationing Locationing            `yaml:"Locationing"`
}

var Conf Config

var configFilePath string

func InitConfig(yamlPath string) error {
	configFilePath = yamlPath
	return LoadService(true)
}

func LoadService(init bool) error {
	configFile := configFilePath
	if configFile == "" {
		configFile = cmdflag.GetConfigFilePath()
	}
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		if init {
			panic(fmt.Sprintf("ReadFile %s failed. Err: %#v", configFile, err))
		} else {
			fmt.Printf("ReadFile %s failed. Err: %#v", configFile, err)
		}
		return err
	}

	conf := Config{}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		if init {
			panic(fmt.Sprintf("yaml.Unmarshal %s failed. Err: %#v", configFile, err))
		} else {
			fmt.Printf("yaml.Unmarshal %s failed. Err: %#v", configFile, err)
		}
		return err
	}
	conf.Locationing.LocaltionEnum = fmt.Sprintf("EL_%s", conf.Locationing.DefaultLocation)

	Conf = conf
	if init && !snowflake.Init(Conf.Common.StartService, int(Conf.Common.ServiceId)) {
		if init {
			panic(fmt.Sprintf("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId))
		} else {
			fmt.Printf("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId)
		}
	}

	logger.Log.Error("init %s success, config: %#v", configFile, Conf)

	return nil
}

func GetServiceType() ss.ServiceType {
	return ss.ServiceType(ss.ServiceType_value[Conf.Common.ServiceType])
}

func GetServiceId() int32 {
	return Conf.Common.ServiceId
}

func GetSelfConnIdx() uint64 {
	iType := GetServiceType()
	idx := GetServiceId()
	return (((uint64(iType) & 0xff) << 16) | (uint64(idx) & 0xffff))
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

func GetAreaId() int32 {
	return ss.EnmLocation_value[Conf.Locationing.LocaltionEnum]
}
