package main

import (
	"gamekit/pkg/configs"
	"gamekit/src/base/control"
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"os/signal"
	"time"

	_ "benchmark/scripts/apps"
	"benchmark/scripts/config"
	"benchmark/scripts/model"
	"context"
	"flag"
	"fmt"
	"os"

	_ "net/http/pprof"
)

var (
	EventChan = make(chan *utils.Event)
)

var pidFile = flag.String("p", "./pid.pid", "进程 pid 文件")

func main() {
	flag.Parse()

	config.InitConfig("../conf/service.yaml")

	pid := os.Getpid()
	os.WriteFile(*pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)

	ctrlConfig := config.Conf.Control
	control.ControlService.Init(ctrlConfig.Service, ctrlConfig.ScrtKey, logger.Log)
	control.ControlService.Serve()
	utils.InitTime(false)

	utils.InitUtils(logger.Log, config.Conf.Common.XXTeaOpen, config.Conf.Common.XXTeaKey, false, 1)
	utils.SetNetLog(config.Conf.Common.NetLog)
	utils.StatMgr.Init(logger.Monitor)
	configs.ResLoadMgr.LoadAll(true)
	model.PlayerMgr.Init()

	ticker := time.NewTicker(time.Duration(float64(1) / float64(config.Conf.Login.QPS) * float64(time.Second)))
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	ctx := context.Background()

	loginIndex := config.Conf.Login.StartUserIdx
	for {
		select {
		case <-ticker.C:
			if model.PlayerMgr.IsFull() {
				logger.Log.Error("all player login ok. Online: %d", model.PlayerMgr.GetOnlineNum())
				ticker.Stop()
				continue
			}
			go model.StartPlayer(context.WithValue(ctx, model.PlayerIndexKey, loginIndex))
			loginIndex++
		case <-sigs:
			// for循环cancel所有的context
			model.PlayerMgr.StopServer()
			return
		}
	}
}
