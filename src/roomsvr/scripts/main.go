package main

import (
	"gamekit/pkg/configs"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/control"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/utils"
	"fmt"
	"os"
	"roomsvr/scripts/config"
	"roomsvr/scripts/ctrl"
	"roomsvr/scripts/framework"
	"roomsvr/scripts/model"
)

func main() {
	cmdflag.ParseFlag()

	switch *cmdflag.StartType {
	case "start":
		config.InitConfig(cmdflag.GetConfigFilePath())

		// 监听 udp 服务端口
		ctrlConfig := config.Conf.Control
		control.ControlService.Init(ctrlConfig.Service, ctrlConfig.ScrtKey, logger.Log)

		pid := os.Getpid()
		os.WriteFile(*cmdflag.PidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
		hostName, _ := os.Hostname()
		osslog.OSS.OssLogInit(config.Conf.Common.GameId, config.Conf.Common.EnvName, config.Conf.Common.ServiceType, config.Conf.Common.ServiceId, config.Conf.Locationing.LocaltionEnum, hostName, control.GetProjectCommitHash(), control.GetProjectBranch())

		logger.Log.Info("roomsvr starting, commit=%s", control.GetProjectCommitHash())
		control.ControlService.Serve()
		ctrl.Init()

		utils.InitUtils(logger.Log, false, "", false, config.GetAreaId())
		utils.SetNetLog(config.Conf.Common.NetLog)

		utils.StatMgr.Init(logger.Monitor)

		// 初始化线程池
		utils.WorkerPool.Init(4)
		configs.ResLoadMgr.LoadAll(true)
		model.RoomMgr.Init()

		framework.StartServe()
	}
}
