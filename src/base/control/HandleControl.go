package control

import (
	"gamekit/pkg/configs"
	ss "gamekit/pkg/ss"
	"gamekit/src/base/logger"
	"gamekit/src/base/relay"
	"gamekit/src/base/snowflake"
	"gamekit/src/base/utils"
	"encoding/json"
	"fmt"
	"os"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/walkon/log4go"
)

/**
 * @brief control/gm 命令处理函数
 *
 * @param args 执行参数，格式为: 命令字 参数列表
 *
 * @param cmdType
 *		  	CmdType_Control: 通过 -s control 进来的命令，这种命令都具有 GmType_Root 权限，此时 param 为 TalkData 类型
 *		  	CmdType_Gm，param 为 Player 类型
 *
 * @param param 参数，具体见 cmdType
 *
 * @retval string: 执行结果; error: 具体错误
 */
type HandleControlCmd func(args []string, cmdType int, param interface{}) (string, error)

type CtrlHandleData struct {
	// 命令字
	Cmd string `json:"cmd"`
	// 使用方法
	Usage string `json:"usage"`
	// 描述
	Desc string `json:"desc"`
	// 命令标签，用于分类筛选
	Tags []string `json:"tags"`

	// 命令类型，区分: CmdType_Control/CmdType_Gm，或者两者都是，理论上 Control 类型都可由 gm 执行
	// Control 如果要执行 gm 命令，则调用 execGm AccId GmCmd Args，因为 Gm 命令要求必须要有 player，所以要带有 AccId
	CmdType int `json:"cmdType"`

	// 如果是 Gm 命令则表示 Gm 权限级别，只有 >= 这个级别的账号才可使用
	GmType int `json:"gmType"`

	// 执行的命令服务器类型，默认为 gamesvr
	SvrType ss.ServiceType `json:"svrType"`

	Handle HandleControlCmd `json:"-"`
}

var (
	handlers map[string]*CtrlHandleData = make(map[string]*CtrlHandleData)
)

func init() {
commonCtrls := map[string]*CtrlHandleData{
		"help": &CtrlHandleData{Usage: "help [n] // n = 1 just control Cmd; n = 2 just gm Cmd; n = 3 all Cmd", Desc: "获取所有命令列表",
			CmdType: utils.CmdType_All, GmType: utils.GmType_Lv1, Handle: help, Tags: []string{TagHelp, TagQuery}},

		"stop": &CtrlHandleData{Usage: "stop", Desc: "停止服务器",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: stop, Tags: []string{TagServer, TagLifecycle}},

		"sleep": &CtrlHandleData{Usage: "sleep seconds", Desc: "服务器主线程休眠指定秒数",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: sleep, Tags: []string{TagServer, TagDebug}},

		"latency": &CtrlHandleData{Usage: "latency millseconds", Desc: "模拟心跳延迟指定毫秒数",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: latency, Tags: []string{TagServer, TagDebug}},

		"startprof": &CtrlHandleData{Usage: "startprof outfile", Desc: "开始CPU性能分析，结果写入指定文件",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: startprof, Tags: []string{TagServer, TagProfiling}},

		"stopprof": &CtrlHandleData{Usage: "stopprof", Desc: "停止CPU性能分析",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: stopprof, Tags: []string{TagServer, TagProfiling}},

		"setNetLog": &CtrlHandleData{Usage: "setNetLog 0/1", Desc: "开启或关闭网络日志",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: setNetLog, Tags: []string{TagServer, TagLog, TagDebug}},

		"loadRes": &CtrlHandleData{Usage: "loadRes all or resName", Desc: "重新加载配置数据文件",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: loadRes, Tags: []string{TagServer, TagConfig}},

		"reloadConfig": {Usage: "reloadConfig", Desc: "重新加载所有配置",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: reloadConfig, Tags: []string{TagServer, TagConfig}},

		"reloadOne": {Usage: "reloadOne resName", Desc: "重新加载指定配置",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: reloadOne, Tags: []string{TagServer, TagConfig}},

		"showRelay": &CtrlHandleData{Usage: "showRelay", Desc: "显示所有中继服务连接信息",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: showRelay, Tags: []string{TagServer, TagNetwork, TagQuery}},

		"getAliveRelay": &CtrlHandleData{Usage: "getAliveRelay", Desc: "获取存活的中继服务区域信息",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: getAliveRelay, Tags: []string{TagServer, TagNetwork, TagQuery}},

		"closeRelay": &CtrlHandleData{Usage: "closeRelay connIdx", Desc: "关闭指定中继服务连接",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: closeRelay, Tags: []string{TagServer, TagNetwork}},

		"setloglevel": &CtrlHandleData{Usage: "setloglevel lvl", Desc: "设置日志级别(2:DEBUG 3:TRACE 4:INFO 5:WARNING 6:ERROR)",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: setloglevel, Tags: []string{TagServer, TagLog}},

		"rollingUpdate": &CtrlHandleData{Usage: "rollingUpdate", Desc: "触发滚动更新",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: rollingUpdate, Tags: []string{TagServer, TagLifecycle}},

		"buildInfo": &CtrlHandleData{Usage: "buildInfo", Desc: "显示构建版本信息",
			CmdType: utils.CmdType_All, GmType: utils.GmType_Root, Handle: getBuildInfo, Tags: []string{TagServer, TagQuery}},

		"parseArea": &CtrlHandleData{Usage: "parseArea accId", Desc: "解析账号ID的区域信息",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: parseArea, Tags: []string{TagServer, TagDebug}},

		"parseConnIdx": &CtrlHandleData{Usage: "parseConnIdx idx", Desc: "解析连接索引的服务类型和序号",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: parseConnIdx, Tags: []string{TagServer, TagDebug}},

		"printConfig": {Usage: "printConfig filename", Desc: "导出指定配置文件内容到本地文件",
			CmdType: utils.CmdType_Control, GmType: utils.GmType_Root, Handle: printConfig, Tags: []string{TagServer, TagConfig, TagQuery}},
	}

	AddControls(commonCtrls)
}

func AddControl(ctrlData *CtrlHandleData) {
	if ctrlData == nil {
		return
	}
	AddControlWithName(ctrlData.Cmd, ctrlData)
}

func AddControlWithName(cmd string, ctrlData *CtrlHandleData) {
	if ctrlData == nil {
		return
	}

	if cmd == "" {
		panic("control command name is empty")
	}

	if ctrlData.Cmd != "" && ctrlData.Cmd != cmd {
		panic(fmt.Sprintf("control command mismatch. key: %s, cmd: %s", cmd, ctrlData.Cmd))
	}

	if _, exists := handlers[cmd]; exists {
		panic(fmt.Sprintf("duplicated control command: %s", cmd))
	}

	ctrlData.Cmd = cmd
	handlers[cmd] = ctrlData
}

func AddControls(ctrls map[string]*CtrlHandleData) {
	for cmd, ctrl := range ctrls {
		AddControlWithName(cmd, ctrl)
	}
}

func GetAllHandlers() map[string]*CtrlHandleData {
	return handlers
}

func GetHandler(cmd string) *CtrlHandleData {
	h, ok := handlers[cmd]
	if !ok || h == nil {
		return nil
	}
	return h
}

// 判断GM权限
func GetCmdGmType(cmd string) (bool, int) {
	if chd, ok := handlers[cmd]; !ok {
		return false, utils.GmType_None
	} else {
		return true, chd.GmType
	}
}

// 执行GM命令
func Handler(args []string, cmdType int, param utils.IObjString) (string, error) {
	chd, ok := handlers[args[0]]
	if !ok {
		Log.Error("not found control command. Cmd: %s", args[0])
		return "not found command, please use help for detail", nil
	}

	// 判断命令类型
	if (chd.CmdType | cmdType) == 0 {
		Log.Error("invalid cmdType. cmdType: %d, Args: %#v, Data: %#v", cmdType, args, chd)
		return "invalid cmdType", nil
	}

	result, err := chd.Handle(args, cmdType, param)

	if param == nil {
		Log.Debug("handle cmd. CmdType: %d, Args: %#v, Result: %s, err: %#v", cmdType, args, result, err)
	} else {
		Log.Debug("handle cmd. Param: %s, CmdType: %d, Args: %#v, Result: %s, err: %#v", param.String(), cmdType, args, result, err)
	}

	return result, err
}

func stop(args []string, cmdType int, param interface{}) (string, error) {
	event := &utils.Event{
		IObj:    utils.CreateAccIdObject(0),
		EventId: ss.EventType_EId_StopService,
	}
	utils.DispatchEvent(event)
	return "success", nil
}

func rollingUpdate(args []string, cmdType int, param interface{}) (string, error) {
	event := &utils.Event{
		IObj:    utils.CreateAccIdObject(0),
		EventId: ss.EventType_EId_RollServerUpdate,
	}
	utils.DispatchEvent(event)
	return "success", nil
}

func help(args []string, cmdType int, param interface{}) (string, error) {
	var keys []string
	for k, _ := range handlers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var result string
	for _, k := range keys {
		v, _ := handlers[k]
		result = result + fmt.Sprintf("cmd: %s, usage: %s, desc: %s, tags: %v\n", v.Cmd, v.Usage, v.Desc, v.Tags)
	}

	return result, nil
}

func sleep(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) < 2 {
		return "Usage: sleep seconds", nil
	}

	n := utils.StrToInt(args[1])
	time.Sleep(time.Duration(n) * time.Second)

	return "success", nil
}

func latency(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) != 2 {
		return "Usage: latency Milseconds", nil
	}

	t := uint32(utils.StrToInt(args[1]))
	event := &utils.Event{
		IObj:    utils.CreateAccIdObject(0),
		EventId: ss.EventType_EId_Latency,
		Param:   []uint32{t},
	}
	utils.DispatchEvent(event)

	return "success", nil
}

var (
	profing     bool     = false
	outfile     *os.File = nil
	outfileName string
)

func startprof(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) != 2 {
		return "Usage: startprof outfile", nil
	}

	if profing {
		return "already startprof, please stopprof", nil
	}

	var err error
	if outfile, err = os.OpenFile(args[1], os.O_CREATE|os.O_RDWR, 0644); err != nil || outfile == nil {
		return "open file failed", err
	}

	outfileName = args[1]
	pprof.StartCPUProfile(outfile)
	profing = true

	return "success", nil
}

func stopprof(args []string, cmdType int, param interface{}) (string, error) {
	if !profing {
		return "not startprof, please startprof", nil
	}

	if outfile != nil {
		pprof.StopCPUProfile()
		outfile.Close()
	}

	profing = false
	outfile = nil

	// go tool pprof -http={ip}:{port} outfileName
	// return "success", nil
	return fmt.Sprintf("useage: go tool pprof -http={ip}:{port} %s", outfileName), nil
}

func setNetLog(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) != 2 {
		return "Usage: setNetLog 0/1", nil
	}
	able := utils.StrToInt(args[1])
	utils.SetNetLog(able != 0)
	return "success", nil
}

func loadRes(args []string, cmdType int, param interface{}) (string, error) {
	/*
		if len(args) != 2 {
			return "Usage: loadRes all or datacfg", nil
		}

		if args[1] == "all" {
			dc.ResLoadMgr.LoadAll(false, false)
		} else {
			err := dc.ResLoadMgr.Load(args[1])
			if err != nil {
				return fmt.Sprintf("%v", err), nil
			}
		}
	*/
	// @TODO ryder: 加载配置
	return "success", nil
}

func showRelay(args []string, cmdType int, param interface{}) (string, error) {
	return relay.RelayMgr.RelayString(), nil
}

func getAliveRelay(args []string, cmdType int, param interface{}) (string, error) {
	alive := relay.RelayMgr.GetAliveRelayArea(0)
	return alive.String(), nil
}

func closeRelay(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) != 2 {
		return "Usage: closeRelay connIdx", nil
	}
	idx := uint64(utils.StrToInt(args[1]))
	connIdx := (((uint64(ss.ServiceType_RelaySvr) & 0xff) << 16) | (uint64(idx) & 0xffff))
	relay.RelayMgr.CloseRelay(connIdx, fmt.Errorf("control close relay connect"))
	return "success", nil
}

func setloglevel(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) < 2 {
		return fmt.Sprintf("curlevel:%s\ndebug:2, trace:3, info:4, warning:5, error:6", logger.Log.Level.String()), nil
	}

	newLv := utils.StrToInt(args[1])
	if newLv < 2 || newLv > 7 {
		return "error log level", nil
	}

	switch newLv {
	case 2:
		logger.Log.Level = log4go.DEBUG
	case 3:
		logger.Log.Level = log4go.TRACE
	case 4:
		logger.Log.Level = log4go.INFO
	case 5:
		logger.Log.Level = log4go.WARNING
	case 6:
		logger.Log.Level = log4go.ERROR
	}

	return fmt.Sprintf("set ok, newLv:%s", logger.Log.Level.String()), nil

}

// 查看当前版本号
func getBuildInfo(args []string, cmdType int, param interface{}) (string, error) {
	o, err := json.MarshalIndent(BuildInfo, "", "    ")
	if err != nil {
		Log.Warn(err)
	}
	return string(o), nil
}

func reloadConfig(args []string, cmdType int, param interface{}) (string, error) {
	start := time.Now()
	configs.ResLoadMgr.LoadAll(false)
	logger.Log.Error("gm reload config")
	return fmt.Sprintf("success, cost: %v", time.Since(start)), nil
}

func reloadOne(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) < 2 {
		return "Usage: reloadOne resName, eg: reloadOne TaskMainConfig", nil
	}
	start := time.Now()
	err := configs.ResLoadMgr.Load(args[1])
	logger.Log.Error("gm reload one config. config: %s, err: %v", args[1], err)
	return fmt.Sprintf("success, cost: %v", time.Since(start)), nil
}

func parseArea(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) != 2 {
		return "Usage: parseArea accId", nil
	}
	accId := utils.StrToUint64(args[1])
	return fmt.Sprintf("accId: %d, idType: %d, areaId: %d", accId, snowflake.GetIDType(accId), snowflake.ParseAreaId(accId)), nil
}

func parseConnIdx(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) != 2 {
		return "Usage: parseConnIdx idx", nil
	}
	idx := utils.StrToUint64(args[1])
	sType, sIdx := utils.ServiceTypeIdx(idx)
	return fmt.Sprintf("Type: %d, Id: %d", sType, sIdx), nil
}

func printConfig(args []string, cmdType int, param interface{}) (string, error) {
	if len(args) < 2 {
		return "usage: printConfig filename, example: printConfig TaskMainConfig", nil
	}
	loadMgr := configs.ResLoadMgr.GetRes(args[1])
	if nil == loadMgr {
		return fmt.Sprintf("failed to find load Mgr by file: %s\n", args[1]), nil
	}

	utils.WorkerPool.Do(0, "printConfig", func() {
		nowStr := time.Now().Format("2006-01-02_15-04-05")
		fileName := fmt.Sprintf("./%s.%s.txt", args[1], nowStr)
		dataFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return
		}
		defer dataFile.Close()
		if _, err := dataFile.Write([]byte(loadMgr.PrintData())); err != nil {
			logger.Log.Error("printConfig write file failed, file: %s, err: %v", fileName, err)
			return
		}
	})
	return "success", nil
}
