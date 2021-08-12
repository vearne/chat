package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

var initOnce sync.Once
var gcf atomic.Value

type BrokerConfig struct {
	Logger struct {
		Level         string `mapstructure:"level"`
		FilePath      string `mapstructure:"filepath"`
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logger"`

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
}

type LogicConfig struct {
	Logger struct {
		Level         string `mapstructure:"level"`
		FilePath      string `mapstructure:"filepath"`
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logger"`

	LogicDealer struct {
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logic_dealer"`

	MySQLConf struct {
		DSN             string `mapstructure:"dsn"`
		MaxIdleConn     int    `mapstructure:"max_idle_conn"`
		MaxOpenConn     int    `mapstructure:"max_open_conn"`
		ConnMaxLifeSecs int    `mapstructure:"conn_max_life_secs"`
		// 是否启动debug模式
		// 若开启则会打印具体的执行SQL
		Debug bool `mapstructure:"debug"`
	} `mapstructure:"mysql"`
}

func InitBrokerConfig() error {
	log.Println("---InitBrokerConfig---")
	initOnce.Do(func() {
		var cf = BrokerConfig{}
		viper.Unmarshal(&cf)
		gcf.Store(&cf)
	})
	return nil
}

func GetBrokerOpts() *BrokerConfig {
	return gcf.Load().(*BrokerConfig)
}

func InitLogicConfig() error {
	log.Println("---InitBrokerConfig---")
	initOnce.Do(func() {
		var cf = LogicConfig{}
		viper.Unmarshal(&cf)
		gcf.Store(&cf)
	})
	return nil
}

func GetLogicOpts() *LogicConfig {
	return gcf.Load().(*LogicConfig)
}
