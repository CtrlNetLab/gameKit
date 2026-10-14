package main

import (
	"gamekit/pkg/configs"
	"gamekit/pkg/ss"
	"gamekit/src/base/cmdflag"
	"gamekit/src/base/control"
	rdao "gamekit/src/base/dao"
	"gamekit/src/base/dirty"
	"gamekit/src/base/logger"
	"gamekit/src/base/osslog"
	"gamekit/src/base/snowflake"
	"fmt"
	"math/rand"
	"os"
	"gamesvr/scripts/config"
	"gamesvr/scripts/ctrl"
	"gamesvr/scripts/dao"
	"gamesvr/scripts/framework"
	"gamesvr/scripts/model"
	"gamesvr/scripts/pprof"

	"gamekit/src/base/utils"
)

func main() {
	cmdflag.ParseFlag()

	switch *cmdflag.StartType {
	case "start":
		config.InitConfig(cmdflag.GetConfigFilePath())

		// 监听 udp 控制指令端口
		ctrlConfig := config.Conf.Control
		control.ControlService.Init(ctrlConfig.Service, ctrlConfig.ScrtKey, logger.Log)

		utils.InitTime(false)
		pid := os.Getpid()
		os.WriteFile(*cmdflag.PidFile, []byte(fmt.Sprintf("%d", pid)), 0644)

		logger.Log.Error("gamesvr starting, commit=%s", control.GetProjectCommitHash())
		control.ControlService.Serve()
		ctrl.Init()
		ctrl.InitDocCache()

		hostName, _ := os.Hostname()
		osslog.OSS.OssLogInit(
			config.Conf.Common.GameId,
			config.Conf.Common.EnvName,
			ss.ServiceType_name[int32(ss.ServiceType_GameSvr)],
			config.Conf.Common.ServiceId,
			config.Conf.Locationing.LocaltionEnum,
			hostName,
			control.GetProjectCommitHash(),
			control.GetProjectBranch(),
		)
		utils.InitUtils(logger.Log, config.Conf.Common.XXTeaOpen, config.Conf.Common.XXTeaKey, config.Conf.Common.PressOpen, config.GetAreaId())
		utils.SetNetLog(config.Conf.Common.NetLog)

		utils.StatMgr.Init(logger.Monitor)
		utils.InitHttpClient(config.Conf.HttpClient)

		model.InitPrometheus()

		// 初始化线程池(Redis/DB 等耗时逻辑必须走 WorkerPool, 见 gamesvr-mainthread skill)
		utils.WorkerPool.Init(config.Conf.Common.WorkerPoolSize)
		// 加载配置表
		configs.ResLoadMgr.LoadAll(true)
		model.PlayerMgr.Init()
		rand.New(rand.NewSource(int64(utils.GetTimeMS())))
		dao.InitMySQL()
		rdao.InitRedisConfigMap(config.Conf.RedisMap)
		rdao.InitRedisMgr()
		snowflake.SetTmpSmallId(10000)
		// 敏感词过滤
		dirty.FilterMgr.Init(true)
		pprof.InitPProf()

		framework.StartServe()
	}
}
