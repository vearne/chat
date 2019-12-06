package broker_engine

import (
	"context"
	"encoding/json"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

type GrpcWorker struct {
	server *grpc.Server
}

func NewGrpcWorker() *GrpcWorker {
	worker := GrpcWorker{}

	worker.server = grpc.NewServer()
	pb.RegisterBrokerServer(worker.server, &worker)
	// Register reflection service on gRPC server.
	reflection.Register(worker.server)

	return &worker
}

func (w *GrpcWorker) Start() {
	lis, err := net.Listen("tcp", config.GetOpts().Broker.GrpcAddress)
	if err != nil {
		zlog.Fatal("failed to listen", zap.Error(err))
	}
	if err := w.server.Serve(lis); err != nil {
		zlog.Fatal("failed to serve", zap.Error(err))
	}
}

func (w *GrpcWorker) Stop() {
	w.server.Stop()
}

func (w *GrpcWorker) ReceiveMsgDialogue(ctx context.Context, in *pb.PushDialogue) (*pb.PushResp, error) {
	zlog.Debug("ReceiveMsgDialogue",  zap.Uint64("senderId", in.SenderId),
		zap.Uint64("sessionId", in.SessionId), zap.String("content", in.Content))

	session := Hub.GetSession(in.ReceiverId)
	req := model.CmdPushDialogueReq{Cmd:consts.CmdPushDialogue, SenderId:in.SenderId,
		SessionId:in.SessionId, Content:in.Content}
	data, _ := json.Marshal(&req)
	session.Write(data)
	return nil, nil
}

func (w *GrpcWorker) ReceiveMsgSignal(context.Context, *pb.PushSignal) (*pb.PushResp, error) {

	return nil, nil
}
