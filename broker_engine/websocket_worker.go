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
	"github.com/vearne/chat/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"time"
)

var (
	LogicClient pb.LogicDealerClient
	Hub         *model.BizHub
	Conn        *grpc.ClientConn
)

type WebsocketWorker struct {
	Server *http.Server
}

func NewWebsocketWorker() *WebsocketWorker {
	zlog.Info("[init]WebServer")
	// init Hub
	Hub = model.NewBizHub()
	var err error

	// logicClient
	Conn, err = grpc.Dial(config.GetOpts().LogicDealer.ListenAddress, grpc.WithInsecure())
	if err != nil {
		zlog.Fatal("con't connect to logic")
	}
	LogicClient = pb.NewLogicDealerClient(Conn)

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

	m.HandlePong(handlePong)
	m.HandleConnect(func(s *melody.Session) {
		zlog.Debug("HandleConnect")
	})
	m.HandleDisconnect(func(*melody.Session) {
		zlog.Debug("HandleDisconnect")
	})
	m.HandleMessage(handlerMessage)
	return r
}

func (worker *WebsocketWorker) Stop() {
	defer Conn.Close()

	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := worker.Server.Shutdown(cxt)
	if err != nil {
		zlog.Error("shutdown error", zap.Error(err))
	}
	zlog.Info("[end]WebsocketWorker exit")
}

func handlePong(s *melody.Session) {

}

func HandleDisconnect(s *melody.Session) {

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

	resp, err := LogicClient.SendMsg(ctx, &req)
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
	resp, err := LogicClient.Match(ctx, &req)
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
	resp, err := LogicClient.CreateAccount(ctx, &req)
	if err != nil {
		zlog.Error("LogicClient.CreateAccount", zap.Error(err))
	}

	zlog.Debug("LogicClient.CreateAccount", zap.Any("resp", resp),
		zap.Uint64("accountId", resp.AccountId))
	// 2 记录accountId和session的对应关系
	Hub.SetSession(resp.AccountId, s)
	s.Set("nickName", req.Nickname)
	s.Set("accountId", resp.AccountId)
	// 3. 返回给客户端
	var result model.CmdCreateAccountResp
	result.AccountId = resp.AccountId
	result.NickName = req.Nickname
	result.Cmd = cmd.Cmd
	data, _ = json.Marshal(&result)
	s.Write(data)
}
