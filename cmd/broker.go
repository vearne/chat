package cmd

import (
	"github.com/spf13/cobra"
	bengine "github.com/vearne/chat/broker_engine"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/resource"
	wm "github.com/vearne/worker_manager"
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
	config.InitBrokerConfig()

	logConfig := config.GetBrokerOpts().Logger
	zlog.InitLogger(&logConfig)
	resource.InitBrokerResource()

	app := wm.NewApp()
	app.AddWorker(bengine.NewWebsocketWorker())
	app.AddWorker(bengine.NewGrpcWorker())
	app.AddWorker(bengine.NewPingWorker())
	app.Run()
}
