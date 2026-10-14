package main

import (
	"gamekit/pkg/configs"
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/control"
	rdao "gamekit/src/base/dao"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/utils"
	"fmt"
	"globalsvr/scripts/config"
	"globalsvr/scripts/ctrl"
	"globalsvr/scripts/dao"
	"globalsvr/scripts/framework"
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

		logger.Log.Info("globalsvr starting, commit=%s", control.GetProjectCommitHash())
		control.ControlService.Serve()
		ctrl.Init()
		hostName, _ := os.Hostname()
		osslog.OSS.OssLogInit(config.Conf.Common.GameId, config.Conf.Common.EnvName, ss.ServiceType_name[int32(ss.ServiceType_GlobalSvr)], config.Conf.Common.ServiceId, "", hostName, control.GetProjectCommitHash(), control.GetProjectBranch())
		utils.InitUtils(logger.Log, false, "", false, 0)
		utils.SetNetLog(config.Conf.Common.NetLog)

		utils.StatMgr.Init(logger.Monitor)

		// 初始化线程池
		utils.WorkerPool.Init(4)

		dao.InitMySQL()

		rdao.InitRedisConfigMap(config.Conf.RedisMap)
		rdao.InitRedisMgr()

		configs.ResLoadMgr.LoadAll(true)

		framework.StartServe()
	}
}
