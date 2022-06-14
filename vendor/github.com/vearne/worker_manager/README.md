# worker_manager
[![golang-ci](https://github.com/vearne/worker_manager/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/worker_manager/actions/workflows/golang-ci.yml)

---
### Overview
Use observer partern to easily manage the start and stop of workers.

* [中文 README](https://github.com/vearne/worker_manager/blob/master/README_zh.md)

### How to get
```
go get github.com/vearne/worker_manager
```
### Example
```
package main

import (
	"context"
	"github.com/gin-gonic/gin"
	wm "github.com/vearne/worker_manager"
	"log"
	"net/http"
	"time"
)

func main() {
	app := wm.NewApp()
	// add 2 load worker
	app.AddWorker(NewLoadWorker())
	app.AddWorker(NewLoadWorker())
	// add 1 web worker
	app.AddWorker(NewWebServer())
	// If not set, the default value will be used
	//app.SetSigs(syscall.SIGTERM, syscall.SIGQUIT)
	app.Run()
}

// some worker

type LoadWorker struct {
	RunningFlag *wm.BoolFlag
	ExitedFlag  *wm.BoolFlag
	ExitChan    chan struct{}
}

func NewLoadWorker() *LoadWorker {
	worker := &LoadWorker{}
	worker.RunningFlag = wm.NewBoolFlag()
	worker.ExitedFlag = wm.NewBoolFlag()
	wm.SetTrue(worker.RunningFlag)
	wm.SetTrue(worker.ExitedFlag)
	worker.ExitChan = make(chan struct{})
	return worker
}

func (worker *LoadWorker) Start() {
	log.Println("[start]LoadWorker")
	for wm.IsTrue(worker.RunningFlag) {
		select {
		case <-time.After(1 * time.Minute):
			//do some thing
			log.Println("LoadWorker do something")
			time.Sleep(time.Second * 3)
		case <-worker.ExitChan:
			log.Println("LoadWorker execute exit logic")
		}
	}
	wm.SetTrue(worker.ExitedFlag)
}

func (worker *LoadWorker) Stop() {
	log.Println("LoadWorker exit...")
	wm.SetFalse(worker.RunningFlag)
	close(worker.ExitChan)
	for !wm.IsTrue(worker.ExitedFlag) {
		time.Sleep(50 * time.Millisecond)
	}
	log.Println("[end]LoadWorker")
}

type WebServer struct {
	Server *http.Server
}

func NewWebServer() *WebServer {
	return &WebServer{}
}

func (worker *WebServer) Start() {
	log.Println("[start]WebServer")

	ginHandler := gin.Default()
	ginHandler.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain", []byte("hello world!"))
	})
	worker.Server = &http.Server{
		Addr:           ":8080",
		Handler:        ginHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	worker.Server.ListenAndServe()
}

func (worker *WebServer) Stop() {
	log.Println("WebServer exit...")
	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// gracefull exit web server
	err := worker.Server.Shutdown(cxt)
	if err != nil {
		log.Printf("shutdown error, %v", err)
	}
	log.Println("[end]WebServer exit")
}
```

```
go build main.go
# Start service
./main
# Send a SIGTERM signal to make the service exit gracefully
# Please replace pid by yourself
kill -15 <pid> 
```
output
```
2022/06/14 11:08:44 [start]WebServer
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

2022/06/14 11:08:44 [start]LoadWorker
[GIN-debug] GET    /                         --> main.(*WebServer).Start.func1 (3 handlers)
2022/06/14 11:08:44 [start]LoadWorker
[GIN] 2022/06/14 - 11:08:52 | 200 |       6.958µs |       127.0.0.1 | GET      "/"
2022/06/14 11:09:08 WebServer exit...
2022/06/14 11:09:08 LoadWorker exit...
2022/06/14 11:09:08 [end]LoadWorker
2022/06/14 11:09:08 LoadWorker execute exit logic
2022/06/14 11:09:08 LoadWorker exit...
2022/06/14 11:09:08 [end]LoadWorker
2022/06/14 11:09:08 LoadWorker execute exit logic
2022/06/14 11:09:08 [end]WebServer exit
```
