package logic

import (
	"context"
	"github.com/json-iterator/go"
	"github.com/vearne/chat/internal/dao"
	zlog "github.com/vearne/chat/internal/log"
	"github.com/vearne/chat/internal/resource"
	pb "github.com/vearne/chat/proto"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
)

type PumpSignalLoopWorker struct {
	RunningFlag *wm.AtomicBool // is running? true:running false:stoped
	ExitedFlag  chan struct{}  //  已经退出的标识
	ExitChan    chan struct{}
}

func NewPumpSignalLoopWorker() *PumpSignalLoopWorker {
	worker := PumpSignalLoopWorker{ExitChan: make(chan struct{})}
	worker.RunningFlag = wm.NewAtomicBool(true)
	worker.ExitedFlag = make(chan struct{})
	return &worker
}

func (w *PumpSignalLoopWorker) Start() {
	zlog.Info("[start]PumpSignalLoopWorker")
	w.PumpSignalLoop()
}

func (w *PumpSignalLoopWorker) Stop() {
	zlog.Info("PumpSignalLoopWorker exit...")
	w.RunningFlag.Set(false)
	close(w.ExitChan)

	<-w.ExitedFlag
	zlog.Info("[end]PumpSignalLoopWorker")
}

func (w *PumpSignalLoopWorker) PumpSignalLoop() {
	for w.RunningFlag.IsTrue() {
		select {
		case msg := <-resource.WaitToBrokerSignalChan:
			pumpSignalToBroker(msg)
		case <-w.ExitChan:
			break
		}
	}
	close(w.ExitedFlag)
}

func pumpSignalToBroker(msg *pb.PushSignal) bool {
	var client pb.BrokerClient
	var err error
	var ok bool
	// 先获取目标所在的broker
	account, err := dao.GetAccount(msg.ReceiverId)
	if err != nil {
		zlog.Error("dao.GetAccount", zap.Error(err))
		return false
	}
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
