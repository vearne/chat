package resource

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"time"
)

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

var (
	// accountId -> chan
	// 信令
	CientSigChanMap map[uint64]chan *pb.PushSignal
	// 对话
	CientDiaChanMap map[uint64]chan *pb.PushDialogue
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
	InitBrokerMap()
}

func InitBrokerMap() {
	CientSigChanMap = make(map[uint64]chan *pb.PushSignal, 10)
	CientDiaChanMap = make(map[uint64]chan *pb.PushDialogue, 10)
}
