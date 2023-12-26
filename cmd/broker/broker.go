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
	clientv3 "go.etcd.io/etcd/client/v3"
	etcdresolver "go.etcd.io/etcd/client/v3/naming/resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
	"log"
	"os"
	"strings"
	"time"
)

var (
	// config file path
	cfgFile     string
	versionFlag bool
)

func init() {
	flag.StringVar(&cfgFile, "config", "", "config file")
	flag.BoolVar(&versionFlag, "version", false, "Show version")

	// register gRPC resolver
	regEtcdResolver()
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

func regEtcdResolver() {
	// ETCD_ENDPOINTS="192.168.2.101:2379;192.168.2.102:2379;192.168.2.103:2379"
	// ETCD_USERNAME=root
	// ETCD_PASSWORD=8323-01AmA004A509
	endpoints, ok := os.LookupEnv("ETCD_ENDPOINTS")
	username := os.Getenv("ETCD_USERNAME")
	password := os.Getenv("ETCD_PASSWORD")
	if !ok {
		return
	}

	log.Println("regEtcdResolver, endpoints:", endpoints)
	cf := clientv3.Config{
		Endpoints:   strings.Split(endpoints, ";"),
		DialTimeout: 5 * time.Second,
		Logger:      zlog.DefaultLogger,
	}
	if len(username) > 0 {
		cf.Username = username
		cf.Password = password
	}
	cli, err := clientv3.New(cf)
	if err != nil {
		log.Fatal("regEtcdResolver", zap.Error(err))
	}

	etcdResolver, err := etcdresolver.NewBuilder(cli)
	if err != nil {
		log.Fatal("regEtcdResolver", zap.Error(err))
	}
	resolver.Register(etcdResolver)
}
