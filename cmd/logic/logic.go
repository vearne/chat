package main

import (
	"flag"
	"fmt"

	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/internal/biz"
	config2 "github.com/vearne/chat/internal/config"
	logic2 "github.com/vearne/chat/internal/engine/logic"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/internal/utils"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
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

	regService()

	app := wm.NewApp()
	app.AddWorker(logic2.NewLogicGrpcWorker())
	app.AddWorker(logic2.NewPumpSignalLoopWorker())
	app.AddWorker(logic2.NewPumpDialogueLoopWorker(1, 5))
	app.AddWorker(logic2.NewBrokerChecker())
	app.Run()
}

func regService() {
	ec := config2.GetLogicOpts().Ectd

	if ec.Register {
		ip, _ := utils.GetIP()
		addr := ip + config2.GetLogicOpts().LogicDealer.ListenAddress
		ser, err := biz.NewServiceRegister(
			ec.Endpoints,
			ec.Username,
			ec.Password,
			"services/logic",
			addr)
		if err != nil {
			zlog.Fatal("regService", zap.Error(err))
		}
		//监听续租相应chan
		go ser.ListenLeaseRespChan()
	}
}
