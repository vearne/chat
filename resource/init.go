package resource

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
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
	InitBrokerHub()
}

func InitBrokerHub() {
	BrokerHub = model.NewBrokerHub()
}

func InitBrokerResource() {
	// init Hub
	Hub = model.NewBizHub()
	var err error

	// logicClient
	addr := config.GetOpts().LogicDealer.ListenAddress

	Conn, err = CreateGrpcClientConn(addr, 3, time.Microsecond*100)
	if err != nil {
		zlog.Fatal("con't connect to logic")
	}
	LogicClient = pb.NewLogicDealerClient(Conn)
}

func CreateGrpcClientConn(addr string, maxRetryCount uint, timeout time.Duration) (*grpc.ClientConn, error) {
	interceptors := make([]grpc.UnaryClientInterceptor, 0)
	r := grpc_retry.UnaryClientInterceptor(
		grpc_retry.WithCodes(
			codes.ResourceExhausted,
			codes.Unavailable,
			codes.Aborted,
			codes.Canceled,
			codes.DeadlineExceeded,
		),
		grpc_retry.WithMax(maxRetryCount),
		grpc_retry.WithPerRetryTimeout(timeout),
	)

	interceptors = append(interceptors, r)
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		//grpc.WithBalancerName(balancerName),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			PermitWithoutStream: true}),
		//grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...)),
	}
	return grpc.Dial(addr, dialOpts...)
}
