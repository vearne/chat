package worker_manager

import (
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	wm      *WorkerManager
	sigList []os.Signal
}

func NewApp() *App {
	var app App
	app.wm = NewWorkerManager()
	// default signals
	app.sigList = []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}
	return &app
}

func (a *App) AddWorker(w Worker) {
	a.wm.AddWorker(w)
}

func (a *App) SetSigs(sig ...os.Signal) {
	a.sigList = sig
}

func (a *App) Run() {
	if len(a.wm.WorkerSlice) <= 0 {
		panic("The number of workers must be greater than 0!")
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, a.sigList...)
	go func() {
		<-ch
		close(ch)
		a.wm.Stop()
	}()
	a.wm.Start()
	a.wm.Wait()
}
