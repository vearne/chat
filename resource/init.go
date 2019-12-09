package resource

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

// -------logic-------
var (
	MySQLClient *gorm.DB
)

var (
	WaitToBrokerDialogueChan chan *pb.PushDialogue
	WaitToBrokerSignalChan   chan *pb.PushSignal
)

var (
	BrokerMap map[string]pb.BrokerClient
)

// ------broker-------
var (
	LogicClient pb.LogicDealerClient
	Hub         *model.BizHub
	Conn        *grpc.ClientConn
)

func InitMySQL() {
	zlog.Info("Init MySQL")
	mysqlConf := config.GetOpts().MySQLConf
	mysqldb, err := gorm.Open("mysql", mysqlConf.DSN)
	if err != nil {
		zlog.Error("initialize_db error", zap.Error(err))
		panic(err)
	}
	if mysqlConf.Debug {
		mysqldb = mysqldb.Debug()
	}
	mysqldb.DB().SetMaxIdleConns(mysqlConf.MaxIdleConn)
	mysqldb.DB().SetMaxOpenConns(mysqlConf.MaxOpenConn)
	mysqldb.DB().SetConnMaxLifetime(time.Duration(mysqlConf.ConnMaxLifeSecs) * time.Second)
	// 赋值给全局变量
	MySQLClient = mysqldb
}

func InitLogicChan() {
	WaitToBrokerDialogueChan = make(chan *pb.PushDialogue, 100)
	WaitToBrokerSignalChan = make(chan *pb.PushSignal, 100)
}

func InitLogicResource() {
	InitLogicChan()
	InitMySQL()
	InitLogicrMap()
}

func InitLogicrMap() {
	BrokerMap = make(map[string]pb.BrokerClient, 10)
}

func InitBrokerResource() {
	// init Hub
	Hub = model.NewBizHub()
	var err error

	// logicClient
	Conn, err = grpc.Dial(config.GetOpts().LogicDealer.ListenAddress, grpc.WithInsecure())
	if err != nil {
		zlog.Fatal("con't connect to logic")
	}
	LogicClient = pb.NewLogicDealerClient(Conn)
}
