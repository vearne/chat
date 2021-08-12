package logic_engine

import (
	"context"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/dao"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"time"
)

const TokenLen = 30

type LogicGrpcWorker struct {
	server *grpc.Server
}

func NewLogicGrpcWorker() *LogicGrpcWorker {
	worker := LogicGrpcWorker{}

	worker.server = grpc.NewServer()
	pb.RegisterLogicDealerServer(worker.server, &LogicServer{})
	// Register reflection service on gRPC server.
	reflection.Register(worker.server)

	return &worker
}

func (w *LogicGrpcWorker) Start() {
	listenAddr := config.GetLogicOpts().LogicDealer.ListenAddress
	zlog.Info("[start]LogicGrpcWorker", zap.String("LogicDealer", listenAddr))
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		zlog.Fatal("failed to listen", zap.Error(err))
	}
	if err := w.server.Serve(lis); err != nil {
		zlog.Fatal("failed to serve", zap.Error(err))
	}
}

func (w *LogicGrpcWorker) Stop() {
	w.server.Stop()
	zlog.Info("[end]LogicGrpcWorker")
}

type LogicServer struct{}

func (s *LogicServer) Reconnect(ctx context.Context, in *pb.ReConnectRequest) (*pb.ReConnectResponse, error) {
	var size int
	resource.MySQLClient.Model(&model.Account{}).Where("id = ? and token = ?",
		in.AccountId, in.Token).Count(&size)

	out := &pb.ReConnectResponse{}
	if size >= 0 {
		// 确实存在的用户
		resource.MySQLClient.Model(&model.Account{}).Update("status", consts.AccountStatusInUse)
		out.Code = pb.CodeEnum_C000
	} else {
		out.Code = pb.CodeEnum_C004
	}
	return out, nil
}

func (s *LogicServer) BrokerOnline(cxt context.Context, in *pb.OnlineRequest) (*pb.OnlineResponse, error) {
	//in.Broker
	ClearUserStatus(in.Broker)
	var resp pb.OnlineResponse
	resp.Code = pb.CodeEnum_C000
	return &resp, nil
}

func (s *LogicServer) ViewedAck(ctx context.Context, req *pb.ViewedAckRequest) (*pb.ViewedAckResponse, error) {
	// 在数据库中做记录
	dao.CreatOrUpdateViewedAck(req.SessionId, req.AccountId, req.MsgId)

	partner := model.SessionAccount{}
	resource.MySQLClient.Where("session_id = ? and account_id != ?",
		req.SessionId, req.AccountId).First(&partner)

	notifyViewedAck(req.AccountId, partner.AccountId, req.SessionId, req.MsgId)

	resp := pb.ViewedAckResponse{Code: pb.CodeEnum_C000}
	return &resp, nil
}

func (s *LogicServer) CreateAccount(ctx context.Context,
	req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	// Broker
	// 192.168.10.100:18223
	token := RandStringBytes(TokenLen)
	var account model.Account
	account.NickName = req.Nickname
	account.Broker = req.Broker
	account.Status = consts.AccountStatusInUse
	account.Token = token
	account.CreatedAt = time.Now()
	account.ModifiedAt = account.CreatedAt
	resource.MySQLClient.Create(&account)

	var resp pb.CreateAccountResponse
	resp.Code = pb.CodeEnum_C000
	resp.AccountId = account.ID
	resp.Token = token
	return &resp, nil
}
func (s *LogicServer) Match(ctx context.Context, req *pb.MatchRequest) (*pb.MatchResponse, error) {
	var partner model.Account
	var session model.Session
	var resp pb.MatchResponse
	sql := "select * from account where status = 1 and id != ? order by rand() limit 1"
	resource.MySQLClient.Raw(sql, req.AccountId).Scan(&partner)
	if partner.ID <= 0 {
		// 找不到合适目标
		resp.Code = pb.CodeEnum_C004
		return &resp, nil
	}
	// 1. 创建会话
	session.Status = consts.SessionStatusInUse
	session.CreatedAt = time.Now()
	session.ModifiedAt = session.CreatedAt
	resource.MySQLClient.Create(&session)
	// 2. 创建会话中的对象 session-account
	s1 := model.SessionAccount{SessionId: session.ID, AccountId: partner.ID}
	resource.MySQLClient.Create(&s1)
	s2 := model.SessionAccount{SessionId: session.ID, AccountId: req.AccountId}
	resource.MySQLClient.Create(&s2)

	// 3. 给被匹配的account发送一个信令，通知他有新的会话建立
	notifyPartnerNewSession(req.AccountId, partner.ID, session.ID)

	resp.PartnerId = partner.ID
	resp.PartnerName = partner.NickName
	resp.SessionId = session.ID
	resp.Code = pb.CodeEnum_C000

	return &resp, nil
}

func (s *LogicServer) SendMsg(ctx context.Context, req *pb.SendMsgRequest) (*pb.SendMsgResponse, error) {
	// 这个的消息可能是 正常的消息 或者 某种信号
	// 比如 1) 用户主动退出会话 2)用户掉线退出会话 3)删除某条消息

	// 1. 存储在发件箱
	outMsg := dao.CreateOutMsg(req.Msgtype, req.SenderId, req.SessionId, req.Content)

	// 判断一下会话的状态，收件人是否退出等情况
	session := dao.GetSession(req.SessionId)
	// 2. 存储在收件箱
	if session.Status == consts.SessionStatusInUse {
		partner := model.SessionAccount{}
		resource.MySQLClient.Where("session_id = ? and account_id != ?",
			outMsg.SessionId, req.SenderId).First(&partner)
		dao.CreateInMsg(req.SenderId, outMsg.ID, partner.AccountId)
		SendPartnerMsg(outMsg.ID, req.SenderId, partner.AccountId, req.SessionId, req.Content)

	} else {
		// 由系统产生一条消息，来替代用户发出的消息
		// 消息的接收人已经退出了
		partner := model.SessionAccount{}
		resource.MySQLClient.Where("session_id = ? and account_id != ?",
			req.SessionId, req.SenderId).First(&partner)
		notifyPartnerExit(req.SenderId, partner.SessionId, partner.AccountId)
	}

	// push
	resp := pb.SendMsgResponse{Code: pb.CodeEnum_C000, MsgId: outMsg.ID}
	return &resp, nil
}

func (s *LogicServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	handlerLogout(req.AccountId)
	var resp pb.LogoutResponse
	resp.Code = pb.CodeEnum_C000
	return &resp, nil
}

func notifyPartnerExit(receiverId, sessionId uint64, exiterId uint64) {
	resource.WaitToBrokerSignalChan <- &pb.PushSignal{
		SignalType: pb.SignalTypeEnum_PartnerExit,
		SenderId:   consts.SystemSender,
		SessionId:  sessionId,
		ReceiverId: receiverId,
		Data:       &pb.PushSignal_AccountId{AccountId: exiterId},
	}
	zlog.Debug("notifyPartnerExit, 1.send signal to broker")
	// 存入数据库
	// outbox
	outMsg := dao.CreateOutMsg(pb.MsgTypeEnum_Signal, consts.SystemSender, sessionId,
		pb.SignalTypeEnum_name[int32(pb.SignalTypeEnum_PartnerExit)])

	// inbox
	dao.CreateInMsg(consts.SystemSender, outMsg.ID, receiverId)
	zlog.Debug("notifyPartnerExit, 2.save to database")
}

func notifyPartnerNewSession(senderId, receiverId, sessionId uint64) {
	//resource.WaitToBrokerSignalChan <- &
	msg := pb.PushSignal{
		SignalType: pb.SignalTypeEnum_NewSession,
		SenderId:   senderId,
		SessionId:  sessionId,
		ReceiverId: receiverId,
	}

	var sender model.Account
	resource.MySQLClient.Where("id = ?", senderId).First(&sender)

	msg.Data = &pb.PushSignal_Partner{Partner: &pb.AccountInfo{
		AccountId: sender.ID,
		NickName:  sender.NickName,
	}}
	resource.WaitToBrokerSignalChan <- &msg
	zlog.Debug("notifyPartnerNewSession, 1.send signal to broker")

	// 存入数据库
	// outbox
	outMsg := dao.CreateOutMsg(pb.MsgTypeEnum_Signal, senderId, sessionId,
		pb.SignalTypeEnum_name[int32(pb.SignalTypeEnum_NewSession)])

	// inbox
	dao.CreateInMsg(senderId, outMsg.ID, receiverId)
	zlog.Debug("notifyPartnerNewSession, 2.save to database")
}

func notifyViewedAck(senderId, receiverId, sessionId uint64, msgId uint64) {
	msg := pb.PushSignal{
		SignalType: pb.SignalTypeEnum_ViewedAck,
		SenderId:   senderId,
		SessionId:  sessionId,
		ReceiverId: receiverId,
		Data:       &pb.PushSignal_MsgId{MsgId: msgId},
	}
	resource.WaitToBrokerSignalChan <- &msg
	zlog.Debug("notifyViewedAck", zap.Uint64("SenderId", senderId),
		zap.Uint64("SessionId", sessionId),
		zap.Uint64("ReceiverId", receiverId), zap.Uint64("msgId", msgId))

}

func SendPartnerMsg(msgId, senderId, receiverId, sessionId uint64, content string) {
	resource.WaitToBrokerDialogueChan <- &pb.PushDialogue{
		MsgId:      msgId,
		SenderId:   senderId,
		SessionId:  sessionId,
		ReceiverId: receiverId,
		Content:    content,
	}
}

func ClearUserStatus(broker string) {
	accounts := make([]model.Account, 0)
	resource.MySQLClient.Model(&model.Account{}).Where("broker = ?", broker).Find(&accounts)
	for _, item := range accounts {
		zlog.Info("logout", zap.Uint64("AccountId", item.ID))
		handlerLogout(item.ID)
	}
}

func handlerLogout(accountId uint64) {
	// 1, 把账号置为退出
	resource.MySQLClient.Model(&model.Account{}).Where("id = ?", accountId).Updates(map[string]interface{}{
		"status":      consts.AccountStatusDestroyed,
		"modified_at": time.Now()})

	var itemList []model.SessionAccount
	resource.MySQLClient.Where("account_id = ?", accountId).Find(&itemList)
	for _, item := range itemList {
		// update session
		// 2. 将账号关联的所有会话都退出
		resource.MySQLClient.Model(&model.Session{}).Where("id = ?", item.SessionId).Updates(map[string]interface{}{
			"status":      consts.SessionStatusDestroyed,
			"modified_at": time.Now()})

		// notify parnter
		// 通知这些会话的参与者，会话即将销毁
		var sa model.SessionAccount
		resource.MySQLClient.Where("session_id = ? and account_id != ?", item.SessionId,
			accountId).First(&sa)
		notifyPartnerExit(sa.AccountId, item.SessionId, accountId)
	}

}
