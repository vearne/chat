package cmd

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
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
)

var brokerCmd = &cobra.Command{
	Use:   "broker",
	Short: "broker",
	Long:  "broker",
	Run:   RunBroker,
}

var (
	LogicClient pb.LogicDealerClient
	Hub         *model.BizHub
)

func init() {
	rootCmd.AddCommand(brokerCmd)

}

func RunBroker(cmd *cobra.Command, args []string) {
	// init Hub
	Hub = model.NewBizHub()

	// logicClient
	conn, err := grpc.Dial(config.GetOpts().LogicDealer.ListenAddress, grpc.WithInsecure())
	if err != nil {
		zlog.Fatal("con't connect to logic")
	}
	defer conn.Close()
	LogicClient = pb.NewLogicDealerClient(conn)

	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

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

	r.Run(config.GetOpts().Broker.ListenAddress)

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
	default:
		zlog.Debug("unknow cmd", zap.String("cmd", cmd.Cmd))
	}

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
	result.Cmd = consts.CmdMatch
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
	var crt model.CmdCreateAccountReq
	json.Unmarshal(data, &crt)

	// 1. 请求
	ip, _ := utils.GetIP()
	broker := ip + config.GetOpts().Broker.ListenAddress
	ctx := context.Background()
	req := pb.CreateAccountRequest{Nickname: crt.NickName, Broker: broker}
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
	result.Cmd = consts.CmdCreateAccount
	data, _ = json.Marshal(&result)
	s.Write(data)
}
