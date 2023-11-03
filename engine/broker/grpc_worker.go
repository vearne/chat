package broker

import (
	"context"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
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
	lis, err := net.Listen("tcp", config.GetBrokerOpts().Broker.GrpcAddress)
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

func (w *GrpcWorker) HealthCheck(ctx context.Context, req *pb.HealthCheckReq) (*pb.HealthCheckResp, error) {
	ans := &pb.HealthCheckResp{Code: pb.CodeEnum_C000}
	return ans, nil
}

func (w *GrpcWorker) ReceiveMsgDialogue(ctx context.Context, in *pb.PushDialogue) (*pb.PushResp, error) {
	zlog.Debug("ReceiveMsgDialogue", zap.Uint64("senderId", in.SenderId),
		zap.Uint64("sessionId", in.SessionId), zap.String("content", in.Content))

	client, ok := resource.Hub.GetClient(in.ReceiverId)
	if ok {
		req := model.NewCmdPushDialogueReq()
		req.SenderId = in.SenderId
		req.MsgId = in.MsgId
		req.SessionId = in.SessionId
		req.Content = in.Content
		clientWrite(client, req)
	} else {
		zlog.Info("Receiver offline", zap.Uint64("receiverId", in.ReceiverId))
		req := pb.LogoutRequest{
			AccountId: in.ReceiverId,
			Broker:    config.GetBrokerOpts().BrokerGrpcAddr,
		}
		_, err := resource.LogicClient.Logout(context.Background(), &req)
		if err != nil {
			zlog.Error("LogicClient.Logout", zap.Error(err))
		}
	}

	// result
	resp := pb.PushResp{Code: pb.CodeEnum_C000}
	return &resp, nil
}

func (w *GrpcWorker) ReceiveMsgSignal(ctx context.Context, in *pb.PushSignal) (*pb.PushResp, error) {
	zlog.Info("ReceiveMsgSignal", zap.Uint64("senderId", in.SenderId),
		zap.Uint64("sessionId", in.SessionId), zap.String("signalType",
			pb.SignalTypeEnum_name[int32(in.SignalType)]))
	// result
	resp := &pb.PushResp{Code: pb.CodeEnum_C000}

	switch in.SignalType {
	case pb.SignalTypeEnum_NewSession:
		zlog.Debug("NewSession")

		req := model.NewCmdPushSignalReq()
		req.SenderId = in.SenderId
		req.SignalType = pb.SignalTypeEnum_name[int32(in.SignalType)]
		req.SessionId = in.SessionId
		req.ReceiverId = in.ReceiverId
		req.Data = in.GetPartner()

		client, ok := resource.Hub.GetClient(in.ReceiverId)
		if ok {
			clientWrite(client, req)
		}

	case pb.SignalTypeEnum_PartnerExit:
		zlog.Debug("PartnerExit")
		/*
		   {
		   	"cmd": "PUSH_SIGNAL_REQ",
		   	"signalType: "PartnerExit"
		       "senderId": 1111,
		       "sessionId": 10000,
		       "receiverId": 12000,
		       "data":{
		           "accountId":  1000,
		       }
		   }
		*/

		req := model.NewCmdPushSignalReq()
		req.SenderId = in.SenderId
		req.SignalType = pb.SignalTypeEnum_name[int32(in.SignalType)]
		req.SessionId = in.SessionId
		req.ReceiverId = in.ReceiverId
		req.Data = map[string]uint64{"accountId": in.GetAccountId()}

		client, ok := resource.Hub.GetClient(in.ReceiverId)
		if ok {
			clientWrite(client, req)
		}

	case pb.SignalTypeEnum_DeleteMsg:
		zlog.Debug("DeleteMsg")

	case pb.SignalTypeEnum_ViewedAck:
		zlog.Debug("ViewedAck")

		req := model.NewCmdPushViewedAckReq()
		req.SessionId = in.SessionId
		req.AccountId = in.SenderId
		req.MsgId = in.GetMsgId()

		client, ok := resource.Hub.GetClient(in.ReceiverId)
		if ok {
			clientWrite(client, req)
		}
	default:
		zlog.Error("unknow SignalType", zap.Int32("SignalType", int32(in.SignalType)))
	}

	// result
	return resp, nil
}

func clientWrite(client *model.Client, obj any) {
	err := client.Write(obj)
	if err != nil {
		zlog.Error("clientWrite", zap.Error(err))
	}
}
