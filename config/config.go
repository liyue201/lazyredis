package config

import "github.com/jinzhu/configor"

const DefaultConfigFile = "conf.yml"

type (
	RedisConfig struct {
		Addr     []string `yaml:"addr"`
		Password string   `yaml:"password"`
		Db       int      `yaml:"db"`
	}

	AppConfig struct {
		Redis RedisConfig `yaml:"redis"`
	}
)

func Load(path string) (*AppConfig, error) {
	conf := &AppConfig{}
	err := configor.Load(conf, path)
	return  conf, err
}
