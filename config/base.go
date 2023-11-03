package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"sync"
	"sync/atomic"
)

var initOnce sync.Once
var gcf atomic.Value

type LogConfig struct {
	Level    string `mapstructure:"level"`
	FilePath string `mapstructure:"filepath"`
}

type MySQLConf struct {
	DSN             string `mapstructure:"dsn"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	ConnMaxLifeSecs int    `mapstructure:"conn_max_life_secs"`
	// 是否启动debug模式
	// 若开启则会打印具体的执行SQL
	Debug bool `mapstructure:"debug"`
}

func ReadConfig(role string, cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)

	} else {
		viper.AddConfigPath("config")
		fname := fmt.Sprintf("config.%s", role)
		viper.SetConfigName(fname)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Println("can't find config file", err)
	}
}
