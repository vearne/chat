package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	lengine "github.com/vearne/chat/logic_engine"
	"github.com/vearne/chat/resource"
	wm "github.com/vearne/worker_manager"
)

var logicCmd = &cobra.Command{
	Use:   "logic",
	Short: "logic dealer",
	Long:  "logic dealer",
	Run:   RunLogic,
}

func init() {
	rootCmd.AddCommand(logicCmd)
}

func RunLogic(cmd *cobra.Command, args []string) {
	// 1. init resource
	initConfig("logic")
	config.InitLogicConfig()

	logConfig := config.GetLogicOpts().Logger
	zlog.InitLogger(&logConfig)

	resource.InitLogicResource()

	fmt.Println("logic starting ... ")

	app := wm.NewApp()
	app.AddWorker(lengine.NewLogicGrpcWorker())
	app.AddWorker(lengine.NewPumpSignalLoopWorker())
	app.AddWorker(lengine.NewPumpDialogueLoopWorker(1, 5))
	app.AddWorker(lengine.NewBrokerChecker())
	app.Run()
}
