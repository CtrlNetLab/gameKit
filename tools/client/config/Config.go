package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host   string `yaml:"Host"`
	UserId string `yaml:"UserId"`
}

var Conf Config

func init() {
	configFile := "conf.yaml"

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("ReadFile %s failed. Err: %#v", configFile, err))
	}

	Conf = Config{}

	err = yaml.Unmarshal(yamlFile, &Conf)
	if err != nil {
		panic(fmt.Sprintf("yaml.Unmarshal %s failed. Err: %#v", configFile, err))
	}

	fmt.Printf("init %s success, config: %#v\n", configFile, Conf)
}
