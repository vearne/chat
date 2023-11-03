package logic

import (
	"flag"
	"fmt"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/engine/logic"
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

	config.ReadConfig("logic", cfgFile)
	config.InitLogicConfig()

	zlog.InitLogger(&config.GetLogicOpts().Logger)
	resource.InitLogicResource()

	fmt.Println("logic starting ... ")

	app := wm.NewApp()
	app.AddWorker(logic.NewLogicGrpcWorker())
	app.AddWorker(logic.NewPumpSignalLoopWorker())
	app.AddWorker(logic.NewPumpDialogueLoopWorker(1, 5))
	app.AddWorker(logic.NewBrokerChecker())
	app.Run()
}
