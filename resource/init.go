package resource

import (
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"

	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	_ "github.com/vearne/chat/resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	var conn *grpc.ClientConn

	// logicClient
	addr := config.GetOpts().LogicDealer.ListenAddress
	zlog.Info("logic addr", zap.String("addr", addr))

	conn, err = CreateGrpcClientConn(addr, 3, time.Second*3)
	if err != nil {
		zlog.Fatal("con't connect to logic", zap.Error(err))
	}

	LogicClient = pb.NewLogicDealerClient(conn)
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
		// 负载均衡策略
		grpc.WithBalancerName(roundrobin.Name),
		//grpc.WithBalancerName(balancerName),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			PermitWithoutStream: true}),
		//grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...)),
	}
	return grpc.Dial(addr, dialOpts...)
}
