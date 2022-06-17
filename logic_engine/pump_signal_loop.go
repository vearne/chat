package logic_engine

import (
	"context"
	"github.com/json-iterator/go"
	"github.com/vearne/chat/dao"
	zlog "github.com/vearne/chat/log"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
	"time"
)

type PumpSignalLoopWorker struct {
	RunningFlag *wm.BoolFlag // is running? true:running false:stoped
	ExitedFlag  *wm.BoolFlag //  Exit Flag
	ExitChan    chan struct{}
}

func NewPumpSignalLoopWorker() *PumpSignalLoopWorker {
	worker := PumpSignalLoopWorker{ExitChan: make(chan struct{})}
	worker.RunningFlag = wm.NewBoolFlag(true)
	worker.ExitedFlag = wm.NewBoolFlag(false)
	return &worker
}

func (w *PumpSignalLoopWorker) Start() {
	zlog.Info("[start]PumpSignalLoopWorker")
	w.PumpSignalLoop()
	wm.SetTrue(w.ExitedFlag)
}

func (w *PumpSignalLoopWorker) Stop() {
	zlog.Info("PumpSignalLoopWorker exit...")
	wm.SetFalse(w.RunningFlag)
	close(w.ExitChan)

	for !wm.IsTrue(w.ExitedFlag) {
		time.Sleep(50 * time.Millisecond)
	}
	zlog.Info("[end]PumpSignalLoopWorker")
}

func (w *PumpSignalLoopWorker) PumpSignalLoop() {
	for wm.IsTrue(w.RunningFlag) {
		select {
		case msg := <-resource.WaitToBrokerSignalChan:
			pumpSignalToBroker(msg)
		case <-w.ExitChan:
			break
		}
	}
}

func pumpSignalToBroker(msg *pb.PushSignal) bool {
	var client pb.BrokerClient
	var err error
	var ok bool
	// 先获取目标所在的broker
	account := dao.GetAccount(msg.ReceiverId)
	if client, ok = resource.BrokerHub.GetBroker(account.Broker); !ok {
		client, err = CreateBrokerClient(account.Broker)
		if err != nil {
			zlog.Error("CreateBrokerClient fail", zap.Error(err))
			return false
		}
		resource.BrokerHub.SetBroker(account.Broker, client)
	}
	str, _ := jsoniter.MarshalToString(msg)
	zlog.Debug("----2---", zap.String("msg", str))

	resp, err := client.ReceiveMsgSignal(context.Background(), msg)
	if err != nil {
		zlog.Error("PumpSignalToBroker", zap.Error(err))
		return false
	}
	zlog.Info("PumpSignalToBroker", zap.Int32("code", int32(resp.Code)),
		zap.Uint64("ReceiverId", msg.ReceiverId),
		zap.String("signalType", pb.SignalTypeEnum_name[int32(msg.SignalType)]))

	return true
}
