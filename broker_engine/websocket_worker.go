package broker_engine

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	"go.uber.org/zap"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"time"
)

type WebsocketWorker struct {
	Server *http.Server
}

func NewWebsocketWorker() *WebsocketWorker {
	zlog.Info("[init]WebServer")
	worker := &WebsocketWorker{}
	worker.Server = &http.Server{
		Addr:           config.GetBrokerOpts().Broker.WebSocketAddress,
		Handler:        createGinEngine(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return worker
}

func (worker *WebsocketWorker) Start() {
	zlog.Info("[start]WebsocketWorker")
	// 将之前连接在此broker上的用户，都置为离线
	zlog.Info("WebsocketWorker-ClearUserStatus")

	worker.Server.ListenAndServe()
}

func createGinEngine() *gin.Engine {
	r := gin.Default()
	m := melody.New()
	m.Config.MaxMessageSize = 1024 * 10
	m.Config.MessageBufferSize = 4 * 1024

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	//m.HandlePong(handlePong)
	m.HandleConnect(func(s *melody.Session) {
		zlog.Debug("HandleConnect")
	})
	m.HandleDisconnect(HandleDisconnect)
	m.HandleMessage(handlerMessage)
	return r
}

func (worker *WebsocketWorker) Stop() {
	//defer Conn.Close()
	// 将之前连接在此broker上的用户，都置为离线
	zlog.Info("WebsocketWorker-ClearUserStatus")
	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := worker.Server.Shutdown(cxt)
	if err != nil {
		zlog.Error("shutdown error", zap.Error(err))
	}
	zlog.Info("[end]WebsocketWorker exit")
}

func HandleDisconnect(s *melody.Session) {
	accountId, _ := s.Get("accountId")
	zlog.Info("HandleDisconnect", zap.Uint64("accountId", accountId.(uint64)))

	ExecuteLogout(accountId.(uint64))
	s.Close()
}

func HandlePing(s *melody.Session, data []byte) {
	zlog.Debug("CmdPing")
	var cmd model.CmdPingReq
	json.Unmarshal(data, &cmd)
	resource.Hub.SetLastPong(cmd.AccountId, time.Now())

	// 返回给客户端
	var result model.CmdPingResp
	result.Cmd = consts.CmdPong
	result.AccountId = cmd.AccountId
	data, _ = json.Marshal(&result)
	s.Write(data)
}

func HandlePong(s *melody.Session, data []byte) {
	zlog.Debug("CmdPong")
	var cmd model.CmdPingResp
	json.Unmarshal(data, &cmd)
	resource.Hub.SetLastPong(cmd.AccountId, time.Now())
}

func handlerMessage(s *melody.Session, data []byte) {
	var cmd model.CommonCmd
	json.Unmarshal(data, &cmd)
	switch cmd.Cmd {
	case consts.CmdCreateAccount:
		zlog.Info("handlerMessage", zap.String("msg", string(data)))
		HandleCrtAccount(s, data)
	case consts.CmdMatch:
		zlog.Info("handlerMessage", zap.String("msg", string(data)))
		HandleMatch(s, data)
	case consts.CmdDialogue:
		zlog.Info("handlerMessage", zap.String("msg", string(data)))
		HandleDialogue(s, data)
	case consts.CmdPing:
		HandlePing(s, data)
	case consts.CmdPong:
		HandlePong(s, data)
	case consts.CmdViewedAck:
		HandleViewedAck(s, data)
	case consts.CmdReConnect:
		HandleReConnect(s, data)
	default:
		zlog.Debug("unknow cmd", zap.String("cmd", cmd.Cmd))
	}

}

func HandleReConnect(s *melody.Session, data []byte) {
	zlog.Debug("CmdReConnect")
	var cmd model.CmdReConnectReq
	json.Unmarshal(data, &cmd)
	ctx := context.Background()

	req := pb.ReConnectRequest{
		AccountId: cmd.AccountId,
		Token:     cmd.Token,
		Broker:    config.GetBrokerOpts().BrokerGrpcAddr,
	}
	resp, err := resource.LogicClient.Reconnect(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.Reconnect", zap.Error(err))
	}
	// 重新登录成功
	if resp.Code == pb.CodeEnum_C000 {
		zlog.Debug("LogicClient.Reconnect", zap.Any("resp", resp),
			zap.Uint64("accountId", resp.AccountId))
		// 2 记录accountId和session的对应关系
		resource.Hub.SetClient(resp.Nickname, resp.AccountId, s)
		s.Set("accountId", resp.AccountId)
	}
	var result model.CmdReConnectResp
	result.Cmd = cmd.Cmd
	result.Code = int32(resp.Code)
	data, _ = json.Marshal(&result)
	s.Write(data)
}

func HandleViewedAck(s *melody.Session, data []byte) {
	zlog.Debug("CmdViewedAck")
	var cmd model.CmdViewedAckReq
	json.Unmarshal(data, &cmd)

	ctx := context.Background()
	req := pb.ViewedAckRequest{
		SessionId: cmd.SessionId,
		AccountId: cmd.AccountId,
		MsgId:     cmd.MsgId,
	}
	resp, err := resource.LogicClient.ViewedAck(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.ViewedAck", zap.Error(err))
	}
	var result model.CmdViewedAckResp
	result.Cmd = cmd.Cmd
	result.Code = int32(resp.Code)
	data, _ = json.Marshal(&result)
	s.Write(data)

}

func HandleDialogue(s *melody.Session, data []byte) {
	zlog.Debug("CmdDialogue")
	var cmd model.CmdDialogueReq
	json.Unmarshal(data, &cmd)

	ctx := context.Background()
	req := pb.SendMsgRequest{
		SenderId: cmd.SenderId, SessionId: cmd.SessionId,
		Msgtype: pb.MsgTypeEnum_Dialogue, Content: cmd.Content}

	resp, err := resource.LogicClient.SendMsg(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.HandleDialogue", zap.Error(err))
	}
	var result model.CmdDialogueResp
	result.Cmd = cmd.Cmd
	result.Code = int32(resp.Code)
	result.MsgId = resp.MsgId
	result.RequestId = cmd.RequestId
	data, _ = json.Marshal(&result)
	s.Write(data)
}

func HandleMatch(s *melody.Session, data []byte) {
	zlog.Debug("CmdMatch")
	var cmd model.CmdMatchReq
	json.Unmarshal(data, &cmd)

	// 1. 请求
	ctx := context.Background()
	req := pb.MatchRequest{AccountId: cmd.AccountId}
	resp, err := resource.LogicClient.Match(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.Match", zap.Error(err))
	}
	var result model.CmdMatchResp
	result.Code = int32(resp.Code)
	result.Cmd = cmd.Cmd
	if resp.Code == pb.CodeEnum_C000 {
		result.Cmd = consts.CmdMatch
		result.PartnerId = resp.PartnerId
		result.PartnerName = resp.PartnerName
		result.SessionId = resp.SessionId
	}
	data, _ = json.Marshal(&result)
	s.Write(data)
}

func HandleCrtAccount(s *melody.Session, data []byte) {
	zlog.Debug("CmdCreateAccount")
	var cmd model.CmdCreateAccountReq
	json.Unmarshal(data, &cmd)

	// 1. 请求
	ctx := context.Background()
	req := pb.CreateAccountRequest{Nickname: cmd.NickName, Broker: config.GetBrokerOpts().BrokerGrpcAddr}
	resp, err := resource.LogicClient.CreateAccount(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.CreateAccount", zap.Error(err))
	}

	zlog.Debug("LogicClient.CreateAccount", zap.Any("resp", resp),
		zap.Uint64("accountId", resp.AccountId))
	// 2 记录accountId和session的对应关系
	resource.Hub.SetClient(req.Nickname, resp.AccountId, s)
	s.Set("accountId", resp.AccountId)

	// 3. 返回给客户端
	var result model.CmdCreateAccountResp
	result.AccountId = resp.AccountId
	result.NickName = req.Nickname
	result.Cmd = cmd.Cmd
	result.Token = resp.Token
	data, _ = json.Marshal(&result)
	s.Write(data)
}

func ExecuteLogout(accountId uint64) {
	// 1. 修改本地状态
	// melody 会清理它的hub
	// 我们只需要清理我们自己的
	resource.Hub.RemoveClient(accountId)

	// 2. 通知其他人
	req := pb.LogoutRequest{
		AccountId: accountId,
		Broker:    config.GetBrokerOpts().BrokerGrpcAddr,
	}
	resp, err := resource.LogicClient.Logout(context.Background(), &req)
	if err != nil {
		zlog.Error("LogicClient.Logout", zap.Error(err))
		return
	}
	zlog.Info("LogicClient.Logout", zap.Uint64("accountId", accountId),
		zap.Int32("code", int32(resp.Code)))

}
