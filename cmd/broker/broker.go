package broker

import (
	"flag"
	"fmt"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/engine/broker"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/resource"
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

	config.ReadConfig("broker", cfgFile)
	config.InitBrokerConfig()

	zlog.InitLogger(&config.GetBrokerOpts().Logger)
	resource.InitBrokerResource()

	app := wm.NewApp()
	app.AddWorker(broker.NewWebsocketWorker())
	app.AddWorker(broker.NewGrpcWorker())
	app.AddWorker(broker.NewPingWorker())
	app.Run()
}
