package broker_engine

import (
	"context"
	"github.com/gin-gonic/gin/internal/json"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	"go.uber.org/zap"
	"time"
)

type PingWorker struct {
	RunningFlag bool // is running? true:running false:stoped
	ExitedFlag  bool //  Exit Flag
	ExitChan    chan struct{}
}

func NewPingWorker() *PingWorker {
	worker := PingWorker{RunningFlag: true, ExitedFlag: false, ExitChan: make(chan struct{})}
	return &worker
}

func (w *PingWorker) start() {
	zlog.Info("[start]PingWorker")
	ch := time.Tick(config.GetOpts().Ping.Interval)

	for w.RunningFlag {
		select {
		case <-ch:
			clients := resource.Hub.GetAllClient()
			for _, client := range clients {
				if time.Since(client.LastPong) > config.GetOpts().Ping.MaxWait {
					// client可能已经掉线
					ExecuteLogout(client.AccountId)

				} else {
					// 执行一次Ping
					cmd := model.CmdPingReq{Cmd: consts.CmdPing, AccountId: client.AccountId}
					buff, _ := json.Marshal(&cmd)
					client.Session.Write(buff)
				}
			}
		case <-w.ExitChan:
			zlog.Info("PingWorker execute exit logic")
		}
	}
	w.ExitedFlag = true
}

func (w *PingWorker) stop() {
	zlog.Info("PingWorker exit...")
	w.RunningFlag = false
	close(w.ExitChan)
	for !w.ExitedFlag {
		time.Sleep(50 * time.Millisecond)
	}
	zlog.Info("[end]PingWorker")
}
