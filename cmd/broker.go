package cmd

import (
	"github.com/spf13/cobra"
	bengine "github.com/vearne/chat/broker_engine"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/resource"
	manager "github.com/vearne/worker_manager"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "broker",
	Long:  "broker",
	Run:   RunBroker,
}

func init() {
	rootCmd.AddCommand(brokerCmd)

}

func RunBroker(cmd *cobra.Command, args []string) {
	// init resource
	initConfig("broker")
	zlog.InitLogger()
	resource.InitBrokerResource()

	// 1. init some worker
	wm := prepareBrokerWorker()

	// 2. start
	wm.Start()

	// 3. register grace exit
	GracefulExit(wm)

	// 4. block and wait
	wm.Wait()
}

func GracefulExit(wm *manager.WorkerManager) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch)
	for sig := range ch {
		switch sig {
		case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
			zlog.Info("got a signal, execute stop", zap.Reflect("signal", sig))
			close(ch)
			wm.Stop()
		case syscall.SIGPIPE:
			zlog.Info("got a signal, ignore SIGPIPE", zap.Reflect("signal", sig))
		default:
			zlog.Info("got a signal, default", zap.Reflect("signal", sig))
		}
	}
}

func prepareBrokerWorker() *manager.WorkerManager {
	wm := manager.NewWorkerManager()

	wm.AddWorker(bengine.NewWebsocketWorker())
	wm.AddWorker(bengine.NewGrpcWorker())
	wm.AddWorker(bengine.NewPingWorker())

	return wm
}
