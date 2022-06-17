package logic_engine

import (
	"context"
	"fmt"
	"github.com/vearne/chat/dao"
	zlog "github.com/vearne/chat/log"
	pb "github.com/vearne/chat/proto"
	"github.com/vearne/chat/resource"
	"github.com/vearne/chat/utils"
	wm "github.com/vearne/worker_manager"
	"go.uber.org/zap"
	"time"
)

type PumpDialogueLoopWorker struct {
	//RunningFlag bool // is running? true:running false:stoped
	ExitedFlag *wm.BoolFlag //  Exit Flag
	ExitChan   chan struct{}
	minCount   int
	maxCount   int
	poolSize   int
	waitGroup  utils.WaitGroupWrapper
}

func NewPumpDialogueLoopWorker(minCount, maxCount int) *PumpDialogueLoopWorker {
	worker := PumpDialogueLoopWorker{ExitChan: make(chan struct{})}
	worker.ExitedFlag = wm.NewBoolFlag(false)
	if minCount <= 1 {
		minCount = 1
	}
	if maxCount <= 1 {
		maxCount = 1
	}
	worker.minCount = minCount
	worker.maxCount = maxCount
	return &worker
}

func (w *PumpDialogueLoopWorker) Start() {
	zlog.Info("[start]PumpDialogueLoopWorker")
	w.PumpDialogueLoop()
	wm.SetTrue(w.ExitedFlag)
}

func (w *PumpDialogueLoopWorker) Stop() {
	zlog.Info("PumpDialogueLoopWorker exit...")
	close(w.ExitChan)

	for !wm.IsTrue(w.ExitedFlag) {
		time.Sleep(50 * time.Millisecond)
	}
	zlog.Info("[end]PumpDialogueLoopWorker")
}

func (w *PumpDialogueLoopWorker) PumpDialogueLoop() {
	// 用于统计使用
	successedCounter := 0
	failedCounter := 0

	// 回收处理结果
	responseCh := make(chan bool, 100)

	// TODO
	go func() {
		for result := range responseCh {
			if result {
				successedCounter++
			} else {
				failedCounter++
			}
		}
	}()

	// 用于结束worker
	closeCh := make(chan int, 1)

	brokerCount := resource.BrokerHub.Size()
	// 期望是处理的协程数 与 broker的数量相等
	w.resizePool(brokerCount, resource.WaitToBrokerDialogueChan, responseCh, closeCh)

	refreshTicker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-refreshTicker.C:
			brokerCount = resource.BrokerHub.Size()
			w.resizePool(brokerCount, resource.WaitToBrokerDialogueChan, responseCh, closeCh)
			continue
		case <-w.ExitChan:
			goto exit
		}
	}

exit:
	close(closeCh)
	refreshTicker.Stop()
	w.waitGroup.Wait()
	wm.SetTrue(w.ExitedFlag)
}

// 参考nsq动态协程池的实现
func (w *PumpDialogueLoopWorker) resizePool(idealPoolSize int, inCh chan *pb.PushDialogue,
	outCh chan bool, closeCh chan int) {
	if idealPoolSize < w.minCount {
		idealPoolSize = w.minCount
	}
	if idealPoolSize > w.maxCount {
		idealPoolSize = w.maxCount
	}
	for {
		if idealPoolSize == w.poolSize {
			break
		} else if idealPoolSize < w.poolSize {
			// contract
			closeCh <- 1
			w.poolSize--
		} else {
			// expand
			w.waitGroup.Wrap(func() {
				pumpDialogueWorker(inCh, outCh, closeCh)
			})
			w.poolSize++
		}
	}
}

func pumpDialogueWorker(inCh chan *pb.PushDialogue, outCh chan bool, closeCh chan int) {
	for {
		select {
		case c := <-inCh:
			result := pumpDialogueToBroker(c)
			outCh <- result
		case <-closeCh:
			return
		}
	}
}

func pumpDialogueToBroker(msg *pb.PushDialogue) bool {
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
	resp, err := client.ReceiveMsgDialogue(context.Background(), msg)
	if err != nil {
		zlog.Error("PumpDialogueToBroker", zap.Error(err))
		return false
	}
	zlog.Info("PumpDialogueToBroker", zap.Int32("code", int32(resp.Code)),
		zap.Uint64("ReceiverId", msg.ReceiverId),
		zap.String("content", msg.Content))
	return true
}
func CreateBrokerClient(broker string) (pb.BrokerClient, error) {
	conn, err := resource.CreateGrpcClientConn(broker, 3, time.Second*3)
	if err != nil {
		zlog.Error("can't connect to logic", zap.String("broker", broker))
		return nil, fmt.Errorf("con't connect to logic:%v", broker)
	}
	//defer conn.Close()
	client := pb.NewBrokerClient(conn)
	return client, nil
}
