package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

func InitConfig(path, conf string) error {
	filename := path + conf
	_, e := os.Stat(filename)
	if e != nil {
		log.Fatalf("file:%s, error:%v\n", filename, e.Error())
		return e
	}

	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(filename)

	e = viper.ReadInConfig()
	if e != nil {
		log.Fatalf("read config file %s fail, error: %v\n", filename, e)
		return e
	}
	return nil
}
