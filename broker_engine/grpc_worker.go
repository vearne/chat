package broker_engine

import (
	"context"
	"encoding/json"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
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

	session, ok := resource.Hub.GetSession(in.ReceiverId)
	if ok {
		req := model.CmdPushDialogueReq{Cmd: consts.CmdPushDialogue,
			SenderId: in.SenderId, MsgId: in.MsgId,
			SessionId: in.SessionId, Content: in.Content}
		data, _ := json.Marshal(&req)
		session.Write(data)
	} else {
		zlog.Info("Receiver offline", zap.Uint64("receiverId", in.ReceiverId))
		req := pb.LogoutRequest{AccountId: in.ReceiverId}
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
	var resp *pb.PushResp
	// result
	resp = &pb.PushResp{Code: pb.CodeEnum_C000}

	switch in.SignalType {
	case pb.SignalTypeEnum_NewSession:
		zlog.Debug("NewSession")
		req := model.CmdPushSignalReq{Cmd: consts.CmdPushSignal, SenderId: in.SenderId,
			SignalType: pb.SignalTypeEnum_name[int32(in.SignalType)],
			SessionId:  in.SessionId, ReceiverId: in.ReceiverId, Data: in.GetPartner()}

		session, ok := resource.Hub.GetSession(in.ReceiverId)
		if ok {
			data, _ := json.Marshal(&req)
			session.Write(data)
		}

	case pb.SignalTypeEnum_PartnerExit:
		zlog.Debug("PartnerExit")
		/*
		   {
		   	"cmd": "PUSH_SIGNAL",
		   	"signalType: "PartnerExit"
		       "senderId": 1111,
		       "sessionId": 10000,
		       "receiverId": 12000,
		       "data":{
		           "accountId":  1000,
		       }
		   }
		*/
		req := model.CmdPushSignalReq{Cmd: consts.CmdPushSignal, SenderId: in.SenderId,
			SignalType: pb.SignalTypeEnum_name[int32(in.SignalType)],
			SessionId:  in.SessionId, ReceiverId: in.ReceiverId,
			Data: map[string]uint64{"accountId": in.GetAccountId()}}

		session, ok := resource.Hub.GetSession(in.ReceiverId)
		if ok {
			data, _ := json.Marshal(&req)
			session.Write(data)
		}

	case pb.SignalTypeEnum_DeleteMsg:
		zlog.Debug("DeleteMsg")

	case pb.SignalTypeEnum_ViewedAck:
		zlog.Debug("ViewedAck")
		req := model.CmdPushViewedAckReq{
			Cmd:       consts.CmdPushViewedAck,
			SessionId: in.SessionId,
			AccountId: in.SenderId,
			MsgId:     in.GetMsgId(),
		}
		session, ok := resource.Hub.GetSession(in.ReceiverId)
		if ok {
			data, _ := json.Marshal(&req)
			session.Write(data)
		}
	default:
		zlog.Error("unknow SignalType", zap.Int32("SignalType", int32(in.SignalType)))
	}

	// result
	return resp, nil
}
