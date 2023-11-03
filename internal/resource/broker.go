package resource

import (
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/vearne/chat/internal/config"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"time"
)

// ------broker-------
var (
	LogicClient pb.LogicDealerClient
	Hub         *model.BizHub
)

// ################## init #################################
func InitBrokerResource() {
	// init Hub
	Hub = model.NewBizHub()
	var err error
	var conn *grpc.ClientConn

	// logicClient
	addr := config.GetBrokerOpts().LogicDealer.Address
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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(
			fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 10,
			PermitWithoutStream: true}),
		//grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(interceptors...)),
	}
	return grpc.Dial(addr, dialOpts...)
}
