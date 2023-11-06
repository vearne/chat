package logic

import (
	"context"
	"github.com/vearne/chat/internal/config"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/internal/utils"
	pb "github.com/vearne/chat/proto"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
	"time"
)

const maxOffLine = 30 * time.Second

/*
确保broker都在线
如果broker已经掉线，就将与broker连接的用户全部下线
*/
type BrokerChecker struct {
	RunningFlag *wm.AtomicBool // 是否运行 true:运行 false:停止
	ExitedFlag  chan struct{}  //  已经退出的标识
	ExitChan    chan struct{}
	// addr -> 上一次健康检查通过的时间
	brokerStatus map[string]time.Time
}

func NewBrokerChecker() *BrokerChecker {
	worker := &BrokerChecker{}
	worker.RunningFlag = wm.NewAtomicBool(true)
	worker.ExitedFlag = make(chan struct{})
	worker.ExitChan = make(chan struct{})
	return worker
}

func (worker *BrokerChecker) Start() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	zlog.Info("[start]BrokerChecker")
	for worker.RunningFlag.IsTrue() {
		select {
		case <-ticker.C:
			zlog.Info("BrokerChecker-ticker trigger")
			worker.checkBroker()
		case <-worker.ExitChan:
			zlog.Info("BrokerChecker-got exit signal from ExitChan")
		}

	}
	zlog.Info("BrokerChecker exit")
	// mark
	close(worker.ExitedFlag)
}

func (worker *BrokerChecker) Stop() {
	worker.RunningFlag.Set(false)
	close(worker.ExitChan)

	<-worker.ExitedFlag
	zlog.Info("[end]BrokerChecker")
}

func (worker *BrokerChecker) checkBroker() {
	begin := time.Now()
	zlog.Info("[start]checkBroker")
	brokerList := resource.BrokerHub.GetBrokerList()
	for _, broker := range brokerList {
		ip, _ := utils.GetIP()
		logicID := ip + config.GetLogicOpts().LogicDealer.ListenAddress
		in := pb.HealthCheckReq{Asker: logicID}
		resp, err := broker.Client.HealthCheck(context.Background(), &in)
		if err != nil {
			zlog.Info("check broker", zap.String("broker", broker.Addr), zap.Error(err))
			continue
		}
		if resp.Code == pb.CodeEnum_C000 {
			worker.brokerStatus[broker.Addr] = time.Now()
		}
	}

	// 清理已经掉线的broker，以及与这些broker关联的账号
	for addr, t := range worker.brokerStatus {
		if time.Since(t) > maxOffLine {
			ClearUserStatus(addr)
			resource.BrokerHub.RemoveBroker(addr)
		}
	}

	zlog.Info("[end]checkBroker", zap.Duration("cost", time.Since(begin)))
}