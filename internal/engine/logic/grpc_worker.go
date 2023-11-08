package logic

import (
	"context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/internal/config"
	dao2 "github.com/vearne/chat/internal/dao"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/middleware"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
	"net"
	"time"
)

const TokenLen = 30

type LogicGrpcWorker struct {
	server *grpc.Server
}

func NewLogicGrpcWorker() *LogicGrpcWorker {
	worker := LogicGrpcWorker{}

	limiter := middleware.NewTokenBucketLimiter(10, 2)

	worker.server = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			ratelimit.UnaryServerInterceptor(limiter),
		),
		grpc_middleware.WithStreamServerChain(
			ratelimit.StreamServerInterceptor(limiter),
		),
	)

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
	var account model.Account
	err := resource.MySQLClient.Model(&model.Account{}).Where("id = ? and token = ?",
		in.AccountId, in.Token).Take(&account).Error

	out := &pb.ReConnectResponse{}
	if err == gorm.ErrRecordNotFound {
		out.Code = pb.CodeEnum_C004
	} else {
		resource.MySQLClient.Model(&model.Account{}).Where("id = ?",
			in.AccountId).Updates(map[string]interface{}{
			"status": consts.AccountStatusInUse,
			"broker": in.Broker,
		})
		out.Code = pb.CodeEnum_C000
		out.AccountId = in.AccountId
		out.Nickname = account.NickName
	}
	return out, nil
}

func (s *LogicServer) ViewedAck(ctx context.Context, req *pb.ViewedAckRequest) (*pb.ViewedAckResponse, error) {
	// 在数据库中做记录
	err := dao2.CreatOrUpdateViewedAck(req.SessionId, req.AccountId, req.MsgId)
	if err != nil {
		zlog.Error("dao2.CreatOrUpdateViewedAck", zap.Error(err))
		return nil, err
	}

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

	err := resource.MySQLClient.Create(&account).Error
	if err != nil {
		zlog.Error("CreateAccount", zap.Error(err))
		return nil, err
	}

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
	outMsg, err := dao2.CreateOutMsg(req.Msgtype, req.SenderId, req.SessionId, req.Content)
	if err != nil {
		zlog.Error("dao2.CreateOutMsg", zap.Error(err))
		return nil, err
	}

	// 判断一下会话的状态，收件人是否退出等情况
	session, err := dao2.GetSession(req.SessionId)
	if err != nil {
		zlog.Error("dao2.GetSession", zap.Error(err))
		return nil, err
	}
	// 2. 存储在收件箱
	if session.Status == consts.SessionStatusInUse {
		partner, err := dao2.GetSessionPartner(outMsg.SessionId, req.SenderId)
		if err != nil {
			zlog.Error("dao2.GetSessionPartner", zap.Error(err))
			return nil, err
		}

		_, err = dao2.CreateInMsg(req.SenderId, outMsg.ID, partner.AccountId)
		if err != nil {
			zlog.Error("dao2.CreateInMsg", zap.Error(err))
			return nil, err
		}
		SendPartnerMsg(outMsg.ID, req.SenderId, partner.AccountId, req.SessionId, req.Content)

	} else {
		// 由系统产生一条消息，来替代用户发出的消息
		// 消息的接收人已经退出了
		partner, err := dao2.GetSessionPartner(req.SessionId, req.SenderId)
		if err != nil {
			zlog.Error("dao2.GetSessionPartner", zap.Error(err))
			return nil, err
		}
		notifyPartnerExit(req.SenderId, partner.SessionId, partner.AccountId)
	}

	// push
	resp := pb.SendMsgResponse{Code: pb.CodeEnum_C000, MsgId: outMsg.ID}
	return &resp, nil
}

func (s *LogicServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	handlerLogout(req.AccountId, req.Broker)
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
	outMsg, err := dao2.CreateOutMsg(pb.MsgTypeEnum_Signal, consts.SystemSender, sessionId,
		pb.SignalTypeEnum_name[int32(pb.SignalTypeEnum_PartnerExit)])
	if err != nil {
		zlog.Error("dao2.CreateOutMsg", zap.Error(err))
		return
	}

	// inbox
	_, err = dao2.CreateInMsg(consts.SystemSender, outMsg.ID, receiverId)
	if err != nil {
		zlog.Error("dao2.CreateInMsg", zap.Error(err))
		return
	}
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

	sender, err := dao2.GetAccount(senderId)
	if err != nil {
		zlog.Error("dao2.GetAccount", zap.Error(err))
		return
	}

	msg.Data = &pb.PushSignal_Partner{Partner: &pb.AccountInfo{
		AccountId: sender.ID,
		NickName:  sender.NickName,
	}}
	resource.WaitToBrokerSignalChan <- &msg
	zlog.Debug("notifyPartnerNewSession, 1.send signal to broker")

	// 存入数据库
	// outbox
	outMsg, err := dao2.CreateOutMsg(pb.MsgTypeEnum_Signal, senderId, sessionId,
		pb.SignalTypeEnum_name[int32(pb.SignalTypeEnum_NewSession)])
	if err != nil {
		zlog.Error("dao2.CreateOutMsg", zap.Error(err))
		return
	}

	// inbox
	_, err = dao2.CreateInMsg(senderId, outMsg.ID, receiverId)
	if err != nil {
		zlog.Error("dao2.CreateInMsg", zap.Error(err))
		return
	}

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
	// 清理某个broker上的所有账号
	// 让他们都下线(登出)
	accounts := make([]model.Account, 0)
	err := resource.MySQLClient.Model(&model.Account{}).Where("broker = ?", broker).Find(&accounts).Error
	if err != nil {
		zlog.Error("ClearUserStatus", zap.Error(err))
		return
	}

	for _, item := range accounts {
		zlog.Info("logout", zap.Uint64("AccountId", item.ID))
		handlerLogout(item.ID, broker)
	}
}

func handlerLogout(accountId uint64, broker string) bool {
	// 1, 把账号置为退出
	result := resource.MySQLClient.Model(&model.Account{}).Where("id = ? AND broker = ? AND status = ?",
		accountId, broker, consts.AccountStatusInUse).Updates(map[string]interface{}{
		"status":      consts.AccountStatusDestroyed,
		"modified_at": time.Now()})

	if result.RowsAffected <= 0 {
		return false
	}

	var itemList []model.SessionAccount
	err := resource.MySQLClient.Where("account_id = ?", accountId).Find(&itemList).Error
	if err != nil {
		zlog.Error("handlerLogout", zap.Error(err))
		return false
	}

	for _, item := range itemList {
		// update session
		// 2. 将账号关联的所有会话都退出
		resource.MySQLClient.Model(&model.Session{}).Where("id = ?",
			item.SessionId).Updates(map[string]interface{}{
			"status":      consts.SessionStatusDestroyed,
			"modified_at": time.Now(),
		})

		// notify parnter
		// 通知这些会话的参与者，会话即将销毁
		var sa model.SessionAccount
		resource.MySQLClient.Where("session_id = ? and account_id != ?", item.SessionId,
			accountId).First(&sa)
		notifyPartnerExit(sa.AccountId, item.SessionId, accountId)
	}
	return true
}
