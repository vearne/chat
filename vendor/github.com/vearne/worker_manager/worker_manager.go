package worker_manager

import (
	"fmt"
	"runtime/debug"
	"sync"
)

type Worker interface {
	Start()
	Stop()
}

type WorkerManager struct {
	sync.WaitGroup
	WorkerSlice []Worker
}

func NewWorkerManager() *WorkerManager {
	workerManager := WorkerManager{}
	workerManager.WorkerSlice = make([]Worker, 0, 10)
	return &workerManager
}

func (wm *WorkerManager) AddWorker(w Worker) {
	wm.WorkerSlice = append(wm.WorkerSlice, w)
}

func (wm *WorkerManager) Start() {
	wm.Add(len(wm.WorkerSlice))
	for _, worker := range wm.WorkerSlice {
		go func(w Worker) {
			defer func() {
				r := recover()
				if r != nil {
					fmt.Printf("WorkerManager error, recover:%v, stack:%v\n",
						r, debug.Stack())
					wm.Done()
				}
			}()
			w.Start()
		}(worker)
	}
}

func (wm *WorkerManager) Stop() {
	for _, worker := range wm.WorkerSlice {
		go func(w Worker) {
			defer func() {
				r := recover()
				if r != nil {
					fmt.Printf("WorkerManager error, recover:%v, stack:%v\n",
						r, debug.Stack())
				}
			}()

			w.Stop()
			wm.Done()
		}(worker)
	}
}
