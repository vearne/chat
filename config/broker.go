package config

import (
	"github.com/spf13/viper"
	"github.com/vearne/chat/utils"
	"log"
	"time"
)

type BrokerConfig struct {
	Logger LogConfig `mapstructure:"logger"`

	LogicDealer struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"logic_dealer"`

	Broker struct {
		WebSocketAddress string `mapstructure:"ws_address"`
		GrpcAddress      string `mapstructure:"grpc_address"`
	} `mapstructure:"broker"`

	Ping struct {
		Interval time.Duration `mapstructure:"interval"`
		MaxWait  time.Duration `mapstructure:"maxWait"`
	} `mapstructure:"ping"`

	BrokerGrpcAddr string
}

func InitBrokerConfig() {
	log.Println("---InitBrokerConfig---")
	initOnce.Do(func() {
		var cf = BrokerConfig{}
		err := viper.Unmarshal(&cf)
		if err != nil {
			log.Fatalf("InitBrokerConfig:%v \n", err)
		}
		// Grpc地址
		ip, _ := utils.GetIP()
		cf.BrokerGrpcAddr = ip + cf.Broker.GrpcAddress
		gcf.Store(&cf)
	})
}

func GetBrokerOpts() *BrokerConfig {
	return gcf.Load().(*BrokerConfig)
}
