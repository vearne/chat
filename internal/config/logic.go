package config

import (
	"github.com/spf13/viper"
	"log"
)

type LogicConfig struct {
	Logger LogConfig `mapstructure:"logger"`

	LogicDealer struct {
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logic_dealer"`

	MySQLConf MySQLConf `mapstructure:"mysql"`

	Ectd struct {
		Register  bool     `mapstructure:"register"`
		Endpoints []string `mapstructure:"endpoints"`
		Username  string   `mapstructure:"username"`
		Password  string   `mapstructure:"password"`
	} `mapstructure:"etcd"`

	ServiceDebug bool `mapstructure:"service-debug"`
}

func InitLogicConfig() {
	log.Println("---InitLogicConfig---")
	initOnce.Do(func() {
		var cf = LogicConfig{}
		err := viper.Unmarshal(&cf)
		if err != nil {
			log.Fatalf("InitLogicConfig:%v \n", err)
		}
		gcf.Store(&cf)
	})
}

func GetLogicOpts() *LogicConfig {
	return gcf.Load().(*LogicConfig)
}
