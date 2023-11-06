package main

import (
	"flag"
	"fmt"
	"github.com/vearne/chat/consts"
	config2 "github.com/vearne/chat/internal/config"
	broker2 "github.com/vearne/chat/internal/engine/broker"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	wm "github.com/vearne/worker_manager"
)

var (
	// config file path
	cfgFile     string
	versionFlag bool
)

func init() {
	flag.StringVar(&cfgFile, "config", "", "config file")
	flag.BoolVar(&versionFlag, "version", false, "Show version")
}

// nolint: unused
func main() {
	flag.Parse()

	if versionFlag {
		fmt.Println("service: chat-broker")
		fmt.Println("Version", consts.Version)
		fmt.Println("BuildTime", consts.BuildTime)
		fmt.Println("GitTag", consts.GitTag)
		return
	}

	config2.ReadConfig("broker", cfgFile)
	config2.InitBrokerConfig()

	zlog.InitLogger(&config2.GetBrokerOpts().Logger)
	resource.InitBrokerResource()

	app := wm.NewApp()
	app.AddWorker(broker2.NewWebsocketWorker())
	app.AddWorker(broker2.NewGrpcWorker())
	app.AddWorker(broker2.NewPingWorker())
	app.Run()
}
