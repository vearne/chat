package logic

import (
	"context"
	"github.com/vearne/chat/internal/config"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	"github.com/vearne/chat/internal/utils"
	pb "github.com/vearne/chat/proto"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

var DefaultExpiredTime = time.Time{}

const maxOffLine = 10 * time.Second

/*
确保broker都在线
如果broker已经掉线，就将与broker连接的用户全部下线
*/
type BrokerChecker struct {
	RunningFlag atomic.Bool   // 是否运行 true:运行 false:停止
	ExitedFlag  chan struct{} //  已经退出的标识
	ExitChan    chan struct{}
	// addr -> 上一次健康检查通过的时间
	brokerStatus map[string]time.Time
}

func NewBrokerChecker() *BrokerChecker {
	worker := &BrokerChecker{}
	worker.RunningFlag.Store(true)
	worker.ExitedFlag = make(chan struct{})
	worker.ExitChan = make(chan struct{})
	worker.brokerStatus = make(map[string]time.Time)
	return worker
}

func (worker *BrokerChecker) Start() {
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	zlog.Info("[start]BrokerChecker")
	for worker.RunningFlag.Load() {
		select {
		case <-ticker.C:
			zlog.Debug("BrokerChecker-ticker trigger")
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
	worker.RunningFlag.Store(false)
	close(worker.ExitChan)

	<-worker.ExitedFlag
	zlog.Info("[end]BrokerChecker")
}

func (worker *BrokerChecker) checkBroker() {
	begin := time.Now()
	zlog.Debug("[start]checkBroker")

	ip, _ := utils.GetIP()
	logicID := ip + config.GetLogicOpts().LogicDealer.ListenAddress
	in := pb.HealthCheckReq{Asker: logicID}

	addrs := GetBrokerList()
	var client pb.BrokerClient
	var err error
	var ok bool

	for _, addr := range addrs {
		client, ok = resource.BrokerHub.GetBroker(addr)
		if !ok {
			client, err = CreateBrokerClient(addr)
			if err != nil {
				zlog.Error("CreateBrokerClient", zap.Error(err))
				if _, ok := worker.brokerStatus[addr]; !ok {
					worker.brokerStatus[addr] = DefaultExpiredTime
				}
				continue
			}
		}

		_, err = client.HealthCheck(context.Background(), &in)
		if err != nil {
			zlog.Info("check broker", zap.String("broker", addr), zap.Error(err))
		} else {
			worker.brokerStatus[addr] = time.Now()
		}
	}

	// 清理已经掉线的broker，以及与这些broker关联的账号
	for addr, t := range worker.brokerStatus {
		if time.Since(t) > maxOffLine {
			zlog.Error("broker may have been offline", zap.String("broker", addr))
			ClearUserStatus(addr)
			resource.BrokerHub.RemoveBroker(addr)
			delete(worker.brokerStatus, addr)
		}
	}

	zlog.Debug("[end]checkBroker", zap.Duration("cost", time.Since(begin)))
}

type BrokerInfo struct {
	Broker string `gorm:"column:broker" json:"broker"`
}

func GetBrokerList() []string {
	brokerList := make([]BrokerInfo, 0)
	err := resource.MySQLClient.Table("account").Distinct("broker").
		Where("status = 1").Find(&brokerList).Error
	if err != nil {
		zlog.Error("GetBrokerList", zap.Error(err))
		return nil
	}
	addrs := make([]string, 0)
	for i := range brokerList {
		addrs = append(addrs, brokerList[i].Broker)
	}
	return addrs
}
