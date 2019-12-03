package config

import (
	"github.com/spf13/viper"
	"log"
	"sync"
	"sync/atomic"
)

var initOnce sync.Once
var gcf atomic.Value

type AppConfig struct {
	LogicDealer struct {
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"logic_dealer"`

	Logger struct {
		Level    string `mapstructure:"level"`
		FilePath string `mapstructure:"filepath"`
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
		ListenAddress string `mapstructure:"listen_address"`
	} `mapstructure:"broker"`
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
