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
	"os"
	"relaysvr/scripts/config"
	"relaysvr/scripts/ctrl"
	"relaysvr/scripts/framework"
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
		osslog.OSS.OssLogInit(config.Conf.Common.GameId, config.Conf.Common.EnvName, ss.ServiceType_name[int32(ss.ServiceType_RelaySvr)], config.Conf.Common.ServiceId, config.Conf.Locationing.LocaltionEnum, hostName, control.GetProjectCommitHash(), control.GetProjectBranch())
		// 监听 udp 服务端口
		logger.Log.Info("relaysvr starting, commit=%s", control.GetProjectCommitHash())
		control.ControlService.Serve()
		ctrl.Init()

		dao.InitRedisConfigMap(config.Conf.RedisMap)
		dao.InitRedisMgr()

		utils.InitUtils(logger.Log, false, "", false, config.GetAreaId())
		utils.SetNetLog(config.Conf.Common.NetLog)

		utils.StatMgr.Init(logger.Monitor)

		// 初始化线程池
		utils.WorkerPool.Init(4)

		framework.StartServe()
	}
}
