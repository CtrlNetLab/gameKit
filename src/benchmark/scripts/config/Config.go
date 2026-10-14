package config

import (
	"gamekit/src/base/logger"
	"gamekit/src/base/snowflake"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Network struct {
	GameSvr []string `yaml:"GameSvr"`
	Proto   string   `yaml:"Proto"`
}

type Control struct {
	Service string `yaml:"Service"`
	ScrtKey string `yaml:"ScrtKey"`
}

type LoginConfig struct {
	QPS          int `yaml:"QPS"`
	StartUserIdx int `yaml:"StartUserIdx"`
}

type Common struct {
	BenchType        string        `yaml:"BenchType"`
	KeepOnline       int           `yaml:"KeepOnline"`
	XXTeaOpen        bool          `yaml:"XXTeaOpen"`
	XXTeaKey         string        `yaml:"XXTeaKey"`
	StartService     string        `yaml:"StartService"`
	ServiceId        int           `yaml:"ServiceId"`
	NetLog           bool          `yaml:"NetLog"`
	DialerTimeout    time.Duration `yaml:"DialerTimeout"`
	HandshakeTimeout time.Duration `yaml:"HandshakeTimeout"`
}

type Config struct {
	Network Network     `yaml:"Network"`
	Control Control     `yaml:"Control"`
	Common  Common      `yaml:"Common"`
	Login   LoginConfig `yaml:"Login"`
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
		configFile = "../conf/service.yaml"
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

	if !snowflake.Init(Conf.Common.StartService, Conf.Common.ServiceId) {
		panic(fmt.Sprintf("init snowflake failed! StartService: %s, ServiceId: %d", Conf.Common.StartService, Conf.Common.ServiceId))
	}

	logger.Log.Error("init %s success, config: %#v", configFile, Conf)
}

func IsBenchLogin() bool {
	return Conf.Common.BenchType == "Login"
}
