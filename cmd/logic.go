package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	zlog "github.com/vearne/chat/log"
	lengine "github.com/vearne/chat/logic_engine"
	"github.com/vearne/chat/resource"
	manager "github.com/vearne/worker_manager"
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
	zlog.InitLogger()
	resource.InitLogicResource()

	fmt.Println("logic starting ... ")
	// 1. init some worker
	wm := prepareLogicWorker()

	// 2. start
	wm.Start()

	// 3. register grace exit
	GracefulExit(wm)

	// 4. block and wait
	wm.Wait()
}

func prepareLogicWorker() *manager.WorkerManager {
	wm := manager.NewWorkerManager()

	wm.AddWorker(lengine.NewLogicGrpcWorker())
	wm.AddWorker(lengine.NewPumpSignalLoopWorker())
	wm.AddWorker(lengine.NewPumpDialogueLoopWorker(1, 5))

	return wm
}
