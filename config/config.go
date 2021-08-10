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

type AppConfig struct {
	LogicDealer struct {
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logic_dealer"`

	Logger struct {
		Level         string `mapstructure:"level"`
		FilePath      string `mapstructure:"filepath"`
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logger"`

	MySQLConf struct {
		DSN             string `mapstructure:"dsn"`
		MaxIdleConn     int    `mapstructure:"max_idle_conn"`
		MaxOpenConn     int    `mapstructure:"max_open_conn"`
		ConnMaxLifeSecs int    `mapstructure:"conn_max_life_secs"`
		// 是否启动debug模式
		// 若开启则会打印具体的执行SQL
		Debug bool `mapstructure:"debug"`
	} `mapstructure:"mysql"`

	Broker struct {
		WebSocketAddress string `mapstructure:"ws_address"`
		GrpcAddress      string `mapstructure:"grpc_address"`
	} `mapstructure:"broker"`

	Ping struct {
		Interval time.Duration `mapstructure:"interval"`
		MaxWait  time.Duration `mapstructure:"maxWait"`
	} `mapstructure:"ping"`
}

func InitConfig() error {
	log.Println("---InitConfig---")
	initOnce.Do(func() {
		var cf = AppConfig{}
		viper.Unmarshal(&cf)
		gcf.Store(&cf)
	})
	return nil
}

func GetOpts() *AppConfig {
	return gcf.Load().(*AppConfig)
}
