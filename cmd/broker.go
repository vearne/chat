package cmd

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
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
	// init resource
	initLogicClient()
	// init Hub
	Hub = model.NewBizHub()

}

func initLogicClient() {
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
		// /ws?nickName=zhangsan
		nickName := c.Query("nickname")
		// 获取broker信息
		broker := "127.0.0.1:18224"
		ctx := context.Background()
		req := pb.CreateAccountRequest{Nickname: nickName, Broker: broker}
		resp, err := LogicClient.CreateAccount(ctx, &req)
		if err != nil {
			// XXX
		}

		// 创建一个用户，以获得用户ID
		m.HandleRequestWithKeys(c.Writer, c.Request, map[string]interface{}{
			"accountId": resp.AccountId,
			"nickName":  nickName,
		})
	})

	m.HandlePong(handlePong)
	m.HandleConnect(handleConnect)
	m.HandleMessage(handlerMessage)
	r.Run(":5000")
}

func handlePong(s *melody.Session) {

}

func handleConnect(s *melody.Session) {
	// 1. 放入hub中，便于以后推送消息
	accountId, _ := s.Get("accountId")
	Hub.SetSession(accountId.(uint64), s)

	// 2. try to match

}

func handlerMessage(s *melody.Session, data []byte) {

}
