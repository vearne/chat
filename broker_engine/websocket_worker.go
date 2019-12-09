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
	"github.com/vearne/chat/utils"
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
		Addr:           config.GetOpts().Broker.WebSocketAddress,
		Handler:        createGinEngine(),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return worker
}

func (worker *WebsocketWorker) Start() {
	zlog.Info("[start]WebsocketWorker")
	worker.Server.ListenAndServe()
}

func createGinEngine() *gin.Engine {
	r := gin.Default()
	m := melody.New()

	//r.GET("/", func(c *gin.Context) {
	//	http.ServeFile(c.Writer, c.Request, "index.html")
	//})

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
	ExecuteLogout(accountId.(uint64))
	s.Close()
}

func HandlePong(s *melody.Session, data []byte) {
	zlog.Debug("CmdPong")
	var cmd model.CmdPingResp
	json.Unmarshal(data, &cmd)
	resource.Hub.SetLastPong(cmd.AccountId, time.Now())
}

func handlerMessage(s *melody.Session, data []byte) {
	zlog.Info("handlerMessage", zap.String("msg", string(data)))
	var cmd model.CommonCmd
	json.Unmarshal(data, &cmd)
	switch cmd.Cmd {
	case consts.CmdCreateAccount:
		HandleCrtAccount(s, data)
	case consts.CmdMatch:
		HandleMatch(s, data)
	case consts.CmdDialogue:
		HandleDialogue(s, data)
	case consts.CmdPong:
		HandlePong(s, data)
	default:
		zlog.Debug("unknow cmd", zap.String("cmd", cmd.Cmd))
	}

}

func HandleDialogue(s *melody.Session, data []byte) {
	zlog.Debug("CmdMatch")
	var cmd model.CmdDialogueReq
	json.Unmarshal(data, &cmd)

	ctx := context.Background()
	req := pb.SendMsgRequest{SenderId: cmd.SenderId, SessionId: cmd.SessionId,
		Msgtype: pb.MsgTypeEnum_Dialogue, Content: cmd.Content}

	resp, err := resource.LogicClient.SendMsg(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.HandleDialogue", zap.Error(err))
	}
	var result model.CmdDialogueResp
	result.Cmd = cmd.Cmd
	result.Code = int32(resp.Code)
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
	ip, _ := utils.GetIP()
	broker := ip + config.GetOpts().Broker.GrpcAddress
	ctx := context.Background()
	req := pb.CreateAccountRequest{Nickname: cmd.NickName, Broker: broker}
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
	data, _ = json.Marshal(&result)
	s.Write(data)
}

func ExecuteLogout(accountId uint64) {
	// 1. 修改本地状态
	// melody 会清理它的hub
	// 我们只需要清理我们自己的
	resource.Hub.RemoveClient(accountId)

	// 2. 通知其他人
	req := pb.LogoutRequest{AccountId: accountId}
	resp, err := resource.LogicClient.Logout(context.Background(), &req)
	if err != nil {
		zlog.Error("LogicClient.Logout", zap.Error(err))
		return
	}
	zlog.Info("LogicClient.Logout", zap.Uint64("accountId", accountId),
		zap.Int32("code", int32(resp.Code)))

}
