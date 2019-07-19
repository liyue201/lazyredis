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

var Conf AppConfig

func Load(path string) (*AppConfig, error) {
	err := configor.Load(&Conf, path)
	return &Conf, err
}
