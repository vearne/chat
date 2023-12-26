package resource

import (
	config2 "github.com/vearne/chat/internal/config"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	BrokerHub *model.BrokerHub
)

// ################## init #################################
func InitLogicResource() {
	ServiceDebug = config2.GetLogicOpts().ServiceDebug

	initLogicChan()
	initMySQL()
	initBrokerHub()

}

func initBrokerHub() {
	zlog.Info("initBrokerHub")
	BrokerHub = model.NewBrokerHub()
}

func initLogicChan() {
	zlog.Info("initLogicChan")
	WaitToBrokerDialogueChan = make(chan *pb.PushDialogue, 100)
	WaitToBrokerSignalChan = make(chan *pb.PushSignal, 100)
}

func initMySQL() {
	zlog.Info("initMySQL")
	var err error
	MySQLClient, err = initMySQLClientError(config2.GetLogicOpts().MySQLConf)
	if err != nil {
		zlog.Fatal("Init MySQL error", zap.Error(err))
	}
}

func initMySQLClientError(cf config2.MySQLConf) (*gorm.DB, error) {
	mysqldb, err := gorm.Open(mysql.Open(cf.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if cf.Debug {
		mysqldb = mysqldb.Debug()
	}

	sqlDB, err := mysqldb.DB()
	if err != nil {
		zlog.Fatal("initialize MySQL error", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(cf.MaxIdleConn)
	sqlDB.SetMaxOpenConns(cf.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Duration(cf.ConnMaxLifeSecs) * time.Second)
	// 赋值给全局变量
	return mysqldb, nil
}
