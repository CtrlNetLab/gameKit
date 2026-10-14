package main

import (
	"gamekit/pkg/configs"
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/control"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/utils"
	"fmt"
	"matchsvr/scripts/config"
	"matchsvr/scripts/ctrl"
	"matchsvr/scripts/framework"
	"matchsvr/scripts/model"
	"os"
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
		osslog.OSS.OssLogInit(config.Conf.Common.GameId, config.Conf.Common.EnvName, ss.ServiceType_name[int32(ss.ServiceType_MatchSvr)], config.Conf.Common.ServiceId, "", hostName, control.GetProjectCommitHash(), control.GetProjectBranch())

		logger.Log.Info("matchsvr starting, commit=%s", control.GetProjectCommitHash())
		control.ControlService.Serve()
		ctrl.Init()

		utils.InitUtils(logger.Log, false, "", false, 0)
		utils.SetNetLog(config.Conf.Common.NetLog)

		utils.StatMgr.Init(logger.Monitor)

		// 初始化线程池
		utils.WorkerPool.Init(4)

		model.RoomServiceMgr.Init()
		configs.ResLoadMgr.LoadAll(true)

		framework.StartServe()
	}
}
