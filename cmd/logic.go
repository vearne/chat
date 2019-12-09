package cmd

import (
	"context"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/spf13/cobra"
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

var logicCmd = &cobra.Command{
	Use:   "logic",
	Short: "logic dealer",
	Long:  "logic dealer",
	Run:   RunLogic,
}

func init() {
	rootCmd.AddCommand(logicCmd)
}

type LogicServer struct{}

func (s *LogicServer) CreateAccount(ctx context.Context,
	req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	// Broker
	// 192.168.10.100:18223
	var account model.Account
	account.NickName = req.Nickname
	account.Broker = req.Broker
	account.Status = consts.AccountStatusInUse
	account.CreatedAt = time.Now()
	account.ModifiedAt = account.CreatedAt
	resource.MySQLClient.Create(&account)

	var resp pb.CreateAccountResponse
	resp.Code = pb.CodeEnum_C000
	resp.AccountId = account.ID
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
		SendPartnerMsg(req.SenderId, partner.AccountId, req.SessionId, req.Content)

	} else {
		// 由系统产生一条消息，来替代用户发出的消息
		// 消息的接收人已经退出了
		partner := model.SessionAccount{}
		resource.MySQLClient.Where("session_id = ? and account_id != ?",
			req.SessionId, req.SenderId).First(&partner)
		notifyPartnerExit(req.SenderId, partner.SessionId, partner.AccountId)
	}

	// push
	resp := pb.SendMsgResponse{Code: pb.CodeEnum_C000}
	return &resp, nil
}

func (s *LogicServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// 1, 把账号置为退出
	resource.MySQLClient.Model(&model.Account{}).Where("id = ?", req.AccountId).Updates(map[string]interface{}{
		"status":      consts.AccountStatusDestroyed,
		"modified_at": time.Now()})

	var itemList []model.SessionAccount
	resource.MySQLClient.Where("account_id = ?", req.AccountId).Find(&itemList)
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
			req.AccountId).First(&sa)
		notifyPartnerExit(sa.AccountId, item.SessionId, req.AccountId)
	}

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

func SendPartnerMsg(senderId, receiverId, sessionId uint64, content string) {
	resource.WaitToBrokerDialogueChan <- &pb.PushDialogue{
		SenderId:   senderId,
		SessionId:  sessionId,
		ReceiverId: receiverId,
		Content:    content,
	}
}

func RunLogic(cmd *cobra.Command, args []string) {
	// 1. init resource
	initConfig("logic")
	zlog.InitLogger()
	resource.InitLogicResource()

	fmt.Println("logic starting ... ")

	// 2. 负责向broker推送
	go PumpSignalToBroker()
	go PumpDialogueToBroker()

	// 3. starting
	zlog.Info("logic dealer running...", zap.String("port",
		config.GetOpts().LogicDealer.ListenAddress))
	lis, err := net.Listen("tcp", config.GetOpts().LogicDealer.ListenAddress)
	if err != nil {
		zlog.Fatal("failed to listen", zap.Error(err))
	}
	s := grpc.NewServer()
	pb.RegisterLogicDealerServer(s, &LogicServer{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		zlog.Fatal("failed to serve", zap.Error(err))
	}
}

func PumpSignalToBroker() {
	for msg := range resource.WaitToBrokerSignalChan {
		var client pb.BrokerClient
		var err error
		var ok bool
		// 先获取目标所在的broker
		account := dao.GetAccount(msg.ReceiverId)
		if client, ok = resource.BrokerMap[account.Broker]; !ok {
			client, err = CreateBrokerClient(account.Broker)
			if err != nil {
				zlog.Error("CreateBrokerClient fail", zap.Error(err))
				continue
			}
			resource.BrokerMap[account.Broker] = client
		}
		str, _ := jsoniter.MarshalToString(msg)
		zlog.Debug("----2---", zap.String("msg", str))

		resp, err := client.ReceiveMsgSignal(context.Background(), msg)
		if err != nil {
			zlog.Error("PumpSignalToBroker", zap.Error(err))
			return
		}
		zlog.Info("PumpSignalToBroker", zap.Int32("code", int32(resp.Code)),
			zap.Uint64("ReceiverId", msg.ReceiverId),
			zap.String("signalType", pb.SignalTypeEnum_name[int32(msg.SignalType)]))
	}
}

func PumpDialogueToBroker() {
	for msg := range resource.WaitToBrokerDialogueChan {
		var client pb.BrokerClient
		var err error
		var ok bool
		// 先获取目标所在的broker
		account := dao.GetAccount(msg.ReceiverId)
		if client, ok = resource.BrokerMap[account.Broker]; !ok {
			client, err = CreateBrokerClient(account.Broker)
			if err != nil {
				zlog.Error("CreateBrokerClient fail", zap.Error(err))
				continue
			}
			resource.BrokerMap[account.Broker] = client
		}
		resp, err := client.ReceiveMsgDialogue(context.Background(), msg)
		if err != nil {
			zlog.Error("PumpDialogueToBroker", zap.Error(err))
			return
		}
		zlog.Info("PumpDialogueToBroker", zap.Int32("code", int32(resp.Code)),
			zap.Uint64("ReceiverId", msg.ReceiverId),
			zap.String("content", msg.Content))
	}
}

func CreateBrokerClient(broker string) (pb.BrokerClient, error) {
	conn, err := grpc.Dial(broker, grpc.WithInsecure())
	if err != nil {
		zlog.Error("con't connect to logic", zap.String("broker", broker))
		return nil, fmt.Errorf("con't connect to logic:%v", broker)
	}
	//defer conn.Close()
	client := pb.NewBrokerClient(conn)
	return client, nil
}
