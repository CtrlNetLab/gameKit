package main

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/control"
	"gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/utils"
	"fmt"
	"miscsvr/scripts/config"
	"miscsvr/scripts/ctrl"
	"miscsvr/scripts/framework"
	"miscsvr/scripts/model"
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
		osslog.OSS.OssLogInit(config.Conf.Common.GameId, config.Conf.Common.EnvName, ss.ServiceType_name[int32(ss.ServiceType_MiscSvr)], config.Conf.Common.ServiceId, config.Conf.Locationing.LocationEnum, hostName, control.GetProjectCommitHash(), control.GetProjectBranch())

		logger.Log.Info("miscsvr starting, commit=%s", control.GetProjectCommitHash())
		control.ControlService.Serve()
		ctrl.Init()

		dao.InitRedisConfigMap(config.Conf.RedisMap)
		dao.InitRedisMgr()

		utils.InitUtils(logger.Log, false, "", false, config.GetAreaId())
		utils.SetNetLog(config.Conf.Common.NetLog)

		utils.StatMgr.Init(logger.Monitor)
		model.MiscMgr.Init()

		// 初始化线程池
		utils.WorkerPool.Init(4)

		framework.StartServe()
	}
}
