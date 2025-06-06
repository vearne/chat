package broker

import (
	"github.com/vearne/chat/internal/config"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/model"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

type PingWorker struct {
	RunningFlag atomic.Bool
	ExitedFlag  chan struct{} //  已经退出的标识
	ExitChan    chan struct{}
}

func NewPingWorker() *PingWorker {
	//RunningFlag: true, ExitedFlag: false
	worker := PingWorker{ExitChan: make(chan struct{})}
	worker.RunningFlag.Store(true)
	worker.ExitedFlag = make(chan struct{})
	return &worker
}

func (w *PingWorker) Start() {
	pingConfig := config.GetBrokerOpts().Ping
	zlog.Info("[start]PingWorker", zap.Duration("Interval", pingConfig.Interval),
		zap.Duration("MaxWait", pingConfig.MaxWait))

	ticker := time.NewTicker(pingConfig.Interval)
	defer ticker.Stop()

	for w.RunningFlag.Load() {
		select {
		case <-ticker.C:
			clients := resource.Hub.GetAllClient()
			for _, client := range clients {
				if time.Since(client.LastPong) > pingConfig.MaxWait {
					// client可能已经掉线
					ExecuteLogout(client.AccountId)

				} else {
					// 执行一次Ping
					cmd := model.NewCmdPingReq()
					cmd.AccountId = client.AccountId
					clientWrite(client, &cmd)
				}
			}
		case <-w.ExitChan:
			zlog.Info("PingWorker execute exit logic")
		}
	}
	close(w.ExitedFlag)
}

func (w *PingWorker) Stop() {
	zlog.Info("PingWorker exit...")
	w.RunningFlag.Store(false)
	close(w.ExitChan)

	<-w.ExitedFlag
	zlog.Info("[end]PingWorker")
}
