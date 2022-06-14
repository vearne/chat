package logic_engine

import (
	"context"
	"github.com/vearne/chat/config"
	zlog "github.com/vearne/chat/log"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	"github.com/vearne/chat/utils"
	"go.uber.org/zap"
	"time"
)

const maxOffLine = 30 * time.Second

/*
	确保broker都在线
	如果broker已经掉线，就将与broker连接的用户全部下线
*/
type BrokerChecker struct {
	RunningFlag bool // 是否运行 true:运行 false:停止
	ExitedFlag  bool //  已经退出的标识
	ExitChan    chan struct{}
	// addr -> 上一次健康检查通过的时间
	brokerStatus map[string]time.Time
}

func NewBrokerChecker() *BrokerChecker {
	worker := &BrokerChecker{RunningFlag: true, ExitedFlag: false}
	worker.ExitChan = make(chan struct{})
	return worker
}

func (worker *BrokerChecker) Start() {
	ticker := time.NewTicker(time.Second * 5)
	zlog.Info("[start]BrokerChecker")
	for worker.RunningFlag {
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
	worker.ExitedFlag = true
}

func (worker *BrokerChecker) Stop() {
	worker.RunningFlag = false
	close(worker.ExitChan)
	for !worker.ExitedFlag {
		time.Sleep(50 * time.Millisecond)
	}
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
			go func() {
				ClearUserStatus(addr)
				resource.BrokerHub.RemoveBroker(addr)
			}()
		}
	}

	zlog.Info("[end]checkBroker", zap.Duration("cost", time.Since(begin)))
}
