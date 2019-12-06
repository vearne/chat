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
	zlog.Debug("ReceiveMsgDialogue", zap.Uint64("senderId", in.SenderId),
		zap.Uint64("sessionId", in.SessionId), zap.String("content", in.Content))

	session := Hub.GetSession(in.ReceiverId)
	req := model.CmdPushDialogueReq{Cmd: consts.CmdPushDialogue, SenderId: in.SenderId,
		SessionId: in.SessionId, Content: in.Content}
	data, _ := json.Marshal(&req)
	session.Write(data)

	// result
	resp := pb.PushResp{Code: pb.CodeEnum_C000}
	return &resp, nil
}

func (w *GrpcWorker) ReceiveMsgSignal(ctx context.Context, in *pb.PushSignal) (*pb.PushResp, error) {
	zlog.Info("ReceiveMsgSignal", zap.Uint64("senderId", in.SenderId),
		zap.Uint64("sessionId", in.SessionId), zap.String("signalType",
			pb.SignalTypeEnum_name[int32(in.SignalType)]))
	var resp *pb.PushResp
	switch in.SignalType {
	case pb.SignalTypeEnum_NewSession:
		zlog.Debug("NewSession")
		req := model.CmdPushSignalReq{Cmd: consts.CmdPushSignal, SenderId: in.SenderId,
			SignalType: pb.SignalTypeEnum_name[int32(in.SignalType)],
			SessionId:  in.SessionId, ReceiverId: in.ReceiverId, Data: in.GetPartner()}

		session := Hub.GetSession(in.ReceiverId)
		data, _ := json.Marshal(&req)
		session.Write(data)

		// result
		// result
		resp = &pb.PushResp{Code: pb.CodeEnum_C000}

	case pb.SignalTypeEnum_PartnerExit:
		zlog.Debug("PartnerExit")
	case pb.SignalTypeEnum_DeleteMsg:
		zlog.Debug("DeleteMsg")
	default:
		zlog.Error("unknow SignalType", zap.Int32("SignalType", int32(in.SignalType)))
	}

	// result
	return resp, nil
}
