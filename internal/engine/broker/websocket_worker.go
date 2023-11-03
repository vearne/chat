package broker

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/vearne/chat/consts"
	"github.com/vearne/chat/internal/config"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/internal/utils"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
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

	zlog.Error(worker.Server.ListenAndServe().Error())
}

func createGinEngine() *gin.Engine {
	r := gin.Default()
	m := melody.New()
	m.Config.MaxMessageSize = 1024 * 10
	m.Config.MessageBufferSize = 4 * 1024

	r.GET("/ws", func(c *gin.Context) {
		err := m.HandleRequest(c.Writer, c.Request)
		if err != nil {
			zlog.Error("websocket", zap.Error(err))
		}
	})

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
	err := s.Close()
	if err != nil {
		zlog.Error("Session close", zap.Error(err))
	}
}

func HandlePing(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdPing")
	var cmd model.CmdPingReq
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandlePing", zap.Error(err))
		return
	}
	resource.Hub.SetLastPong(cmd.AccountId, time.Now())

	// 返回给客户端
	result := model.NewCmdPingResp()
	result.AccountId = cmd.AccountId
	wrapperWrite(wrapper, result)
}

func HandlePong(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdPong")
	var cmd model.CmdPingResp
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandlePong", zap.Error(err))
		return
	}
	resource.Hub.SetLastPong(cmd.AccountId, time.Now())
}

func handlerMessage(s *melody.Session, data []byte) {
	var cmd model.CommonCmd
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("handlerMessage", zap.Error(err))
		return
	}

	wrapper := model.NewSessionWrapper(s)
	switch cmd.Cmd {
	case utils.AssembleCmdReq(consts.CmdCreateAccount):
		zlog.Info("handlerMessage", zap.String("msg", string(data)))
		HandleCrtAccount(wrapper, data)
	case utils.AssembleCmdReq(consts.CmdMatch):
		zlog.Info("handlerMessage", zap.String("msg", string(data)))
		HandleMatch(wrapper, data)
	case utils.AssembleCmdReq(consts.CmdDialogue):
		zlog.Info("handlerMessage", zap.String("msg", string(data)))
		HandleDialogue(wrapper, data)
	case utils.AssembleCmdReq(consts.CmdPing):
		HandlePing(wrapper, data)
	case utils.AssembleCmdResp(consts.CmdPing):
		HandlePong(wrapper, data)
	case utils.AssembleCmdReq(consts.CmdViewedAck):
		HandleViewedAck(wrapper, data)
	case utils.AssembleCmdReq(consts.CmdReConnect):
		HandleReConnect(wrapper, data)
	default:
		zlog.Debug("unknow cmd", zap.String("cmd", cmd.Cmd))
	}

}

func HandleReConnect(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdReConnect")
	var cmd model.CmdReConnectReq
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandleReConnect", zap.Error(err))
		return
	}

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
		resource.Hub.SetClient(resp.Nickname, resp.AccountId, wrapper.Session)
		wrapper.Session.Set("accountId", resp.AccountId)
	}

	result := model.NewCmdReConnectResp()
	result.Code = int32(resp.Code)
	wrapperWrite(wrapper, result)
}

func HandleViewedAck(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdViewedAck")
	var cmd model.CmdViewedAckReq
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandleViewedAck", zap.Error(err))
		return
	}

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

	result := model.NewCmdViewedAckResp()
	result.Code = int32(resp.Code)
	wrapperWrite(wrapper, result)
}

func HandleDialogue(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdDialogue")
	var cmd model.CmdDialogueReq
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandleDialogue", zap.Error(err))
		return
	}

	ctx := context.Background()
	req := pb.SendMsgRequest{
		SenderId: cmd.SenderId, SessionId: cmd.SessionId,
		Msgtype: pb.MsgTypeEnum_Dialogue, Content: cmd.Content}

	resp, err := resource.LogicClient.SendMsg(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.HandleDialogue", zap.Error(err))
	}
	result := model.NewCmdDialogueResp()
	result.Code = int32(resp.Code)
	result.MsgId = resp.MsgId
	result.RequestId = cmd.RequestId
	wrapperWrite(wrapper, result)
}

func HandleMatch(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdMatch")
	var cmd model.CmdMatchReq
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandleMatch", zap.Error(err))
		return
	}

	// 1. 请求
	ctx := context.Background()
	req := pb.MatchRequest{AccountId: cmd.AccountId}
	resp, err := resource.LogicClient.Match(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.Match", zap.Error(err))
	}
	result := model.NewCmdMatchResp()
	result.Code = int32(resp.Code)
	if resp.Code == pb.CodeEnum_C000 {
		result.PartnerId = resp.PartnerId
		result.PartnerName = resp.PartnerName
		result.SessionId = resp.SessionId
	}
	wrapperWrite(wrapper, result)
}

func HandleCrtAccount(wrapper *model.SessionWrapper, data []byte) {
	zlog.Debug("CmdCreateAccount")
	var cmd model.CmdCreateAccountReq
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		zlog.Error("HandleCrtAccount", zap.Error(err))
		return
	}

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
	resource.Hub.SetClient(req.Nickname, resp.AccountId, wrapper.Session)
	wrapper.Session.Set("accountId", resp.AccountId)

	// 3. 返回给客户端
	result := model.NewCmdCreateAccountResp()
	result.AccountId = resp.AccountId
	result.NickName = req.Nickname
	result.Token = resp.Token
	wrapperWrite(wrapper, result)
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

func wrapperWrite(wrapper *model.SessionWrapper, obj any) {
	err := wrapper.Write(obj)
	if err != nil {
		zlog.Error("wrapperWrite", zap.Error(err))
	}
}
