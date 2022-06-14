package broker_engine

import (
	"encoding/json"
	"github.com/vearne/chat/config"
	"github.com/vearne/chat/consts"
	zlog "github.com/vearne/chat/log"
	"github.com/vearne/chat/model"
	"github.com/vearne/chat/resource"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
	"time"
)

type PingWorker struct {
	RunningFlag *wm.BoolFlag
	ExitedFlag  *wm.BoolFlag //  Exit Flag
	ExitChan    chan struct{}
}

func NewPingWorker() *PingWorker {
	//RunningFlag: true, ExitedFlag: false
	worker := PingWorker{ExitChan: make(chan struct{})}
	worker.RunningFlag = wm.NewBoolFlag()
	wm.SetTrue(worker.RunningFlag)
	worker.ExitedFlag = wm.NewBoolFlag()
	wm.SetFalse(worker.ExitedFlag)
	return &worker
}

func (w *PingWorker) Start() {
	pingConfig := config.GetBrokerOpts().Ping
	zlog.Info("[start]PingWorker", zap.Duration("Interval", pingConfig.Interval),
		zap.Duration("MaxWait", pingConfig.MaxWait))
	ch := time.Tick(pingConfig.Interval)

	for wm.IsTrue(w.RunningFlag) {
		select {
		case <-ch:
			clients := resource.Hub.GetAllClient()
			for _, client := range clients {
				if time.Since(client.LastPong) > pingConfig.MaxWait {
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
	wm.SetTrue(w.ExitedFlag)
}

func (w *PingWorker) Stop() {
	zlog.Info("PingWorker exit...")
	wm.SetFalse(w.RunningFlag)
	close(w.ExitChan)
	for !wm.IsTrue(w.ExitedFlag) {
		time.Sleep(50 * time.Millisecond)
	}
	zlog.Info("[end]PingWorker")
}
