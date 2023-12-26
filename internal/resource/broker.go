package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
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
	"google.golang.org/grpc/peer"
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
	ServiceDebug = config.GetBrokerOpts().ServiceDebug

	// logicClient
	addr := config.GetBrokerOpts().LogicDealer.Address
	zlog.Info("logic addr", zap.String("addr", addr))

	conn, err = CreateGrpcClientConn(addr, 3, time.Second*3, ServiceDebug)
	if err != nil {
		zlog.Fatal("con't connect to logic", zap.Error(err))
	}

	LogicClient = pb.NewLogicDealerClient(conn)
}

func CreateGrpcClientConn(addr string, maxRetryCount uint, timeout time.Duration, debug bool) (*grpc.ClientConn, error) {
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

	if debug {
		interceptors = append(interceptors, DebugInterceptor())
	}

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

func DebugInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		p := peer.Peer{}
		opts = append(opts, grpc.Peer(&p))
		err := invoker(ctx, method, req, reply, cc, opts...)
		fmt.Println("addr:", p.Addr)
		color.Red("Call service: %s@%s (%s)", method, p.Addr, time.Since(start))

		data, _ := json.MarshalIndent(req, "", "    ")
		color.Cyan("Request(%s): %s", method, data)

		if err == nil {
			data, _ = json.MarshalIndent(reply, "", "    ")
			color.Cyan("Response(%s): %s", method, data)
		} else {
			color.Cyan("Response(%s): error %v", method, err)
		}

		return err
	}
}
