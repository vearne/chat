package main

import (
	"flag"
	"fmt"
	"github.com/vearne/chat/consts"
	config2 "github.com/vearne/chat/internal/config"
	logic2 "github.com/vearne/chat/internal/engine/logic"
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

// nolint: all
func main() {
	flag.Parse()

	if versionFlag {
		fmt.Println("service: chat-broker")
		fmt.Println("Version", consts.Version)
		fmt.Println("BuildTime", consts.BuildTime)
		fmt.Println("GitTag", consts.GitTag)
		return
	}

	config2.ReadConfig("logic", cfgFile)
	config2.InitLogicConfig()

	zlog.InitLogger(&config2.GetLogicOpts().Logger)
	resource.InitLogicResource()

	fmt.Println("logic starting ... ")

	app := wm.NewApp()
	app.AddWorker(logic2.NewLogicGrpcWorker())
	app.AddWorker(logic2.NewPumpSignalLoopWorker())
	app.AddWorker(logic2.NewPumpDialogueLoopWorker(1, 5))
	app.AddWorker(logic2.NewBrokerChecker())
	app.Run()
}
