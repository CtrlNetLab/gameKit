package utils

import (
	"gamekit/pkg/cs"
	"gamekit/pkg/ss"
	"fmt"
	"sync"
	"time"

	log "github.com/walkon/log4go"
)

var (
	StatMgr *statMgr = &statMgr{}
)

const (
	maxRegionSize int = 26
	// 1000: 毫秒; 1: 微秒
	timeUnitConvert int64 = 1
	maxLoginSize    int   = 32
	maxQPSSize      int   = 32
)

type statCmdObject struct {
	cmdId uint32
	// 调用次数
	callCount uint32
	// 调用总耗时
	totalTime int64
	// 包总大小
	totalSize int

	region    *RegionObject
	qpsObject *StatQPSObject
}

func (this *statCmdObject) init() {
	this.region = &RegionObject{}
	this.region.Init()

	this.qpsObject = &StatQPSObject{}
	this.qpsObject.Init()
}

func (this *statCmdObject) reset() {
	this.callCount = 0
	this.totalTime = 0
	this.totalSize = 0

	this.region.Reset()
	this.qpsObject.Reset()
}

func (this *statCmdObject) add(msgSize int, useTime int64, startTime time.Time) {
	this.callCount++
	this.totalSize += msgSize
	this.totalTime += useTime

	this.region.Add(useTime)
	this.qpsObject.Add(time.Now(), startTime)
}

type StatQPSObject struct {
	qps []int
}

func (this *StatQPSObject) Init() {
	this.qps = make([]int, maxQPSSize)
	this.Reset()
}

func (this *StatQPSObject) Reset() {
	for i := 0; i < maxQPSSize; i++ {
		this.qps[i] = 0
	}
}

func (this *StatQPSObject) Add(now, startTime time.Time) {
	offset := int(now.Unix() - startTime.Unix())
	if offset < 0 || offset >= maxQPSSize {
		return
	}
	this.qps[offset]++
}

func (this *StatQPSObject) String() string {
	var s string = "|"
	for i := 0; i < maxQPSSize; i++ {
		if this.qps[i] <= 0 {
			continue
		}
		s += fmt.Sprintf("%d,%d|", i, this.qps[i])
	}

	if len(s) <= 1 {
		return ""
	}
	return s
}

type statObject struct {
	id        int
	name      string
	callCount uint32
	totalTime int64
	region    *RegionObject
}

func (this *statObject) init() {
	this.region = &RegionObject{}
	this.region.Init()
}

func (this *statObject) reset() {
	this.callCount = 0
	this.totalTime = 0
	this.region.Reset()
}

func (this *statObject) add(useTime int64) {
	this.callCount++
	this.totalTime += useTime
	this.region.Add(useTime)
}

type RegionObject struct {
	// 单位 miscrosecond: 微妙
	// [0, 1e3) 间隔 2 * 1e2
	// [1e3, 1e4) 间隔 2 * 1e3
	// [1e4, 1e5) 间隔 2 * 1e4
	// [1e5, 1e6) 间隔 2 * 1e5
	// [1e6, 1e7) 间隔 2 * 1e6
	// [1e7, inf)
	regList []int64
}

func (this *RegionObject) Init() {
	this.regList = make([]int64, maxRegionSize)
	this.Reset()
}

func (this *RegionObject) Reset() {
	for i := 0; i < maxRegionSize; i++ {
		this.regList[i] = 0
	}
}

func (this *RegionObject) Add(useTime int64) {
	var index int = 0
	if useTime < 1000 {
		index = int(useTime / 200)
	} else if useTime < 10000 {
		index = 5 + int(useTime/2000)
	} else if useTime < 100000 {
		index = 10 + int(useTime/20000)
	} else if useTime < 1000000 {
		index = 15 + int(useTime/200000)
	} else if useTime < 10000000 {
		index = 20 + int(useTime/2000000)
	} else {
		index = maxRegionSize - 1
	}

	if index < maxRegionSize {
		this.regList[index]++
	}
}

func (this *RegionObject) indexRegion(index int) string {
	if index < 5 {
		return fmt.Sprintf("[%d, %d)us", index*200, (index+1)*200-1)
	} else if index >= (maxRegionSize - 1) {
		return "[10, inf)s"
	}

	var (
		step     int
		timeUnit string
	)

	if index < 10 {
		index -= 5
		step = 2
		timeUnit = "ms"
	} else if index < 15 {
		index -= 10
		step = 20
		timeUnit = "ms"
	} else if index < 20 {
		index -= 15
		step = 200
		timeUnit = "ms"
	} else if index < 25 {
		index -= 20
		step = 2
		timeUnit = "s"
	}

	start := index * step
	end := (index+1)*step - 1
	if index == 0 {
		start = step / 2
	}

	return fmt.Sprintf("[%d, %d)%s", start, end, timeUnit)
}

func (this *RegionObject) String() string {
	var s string = "|"
	for i := 0; i < maxRegionSize; i++ {
		if this.regList[i] <= 0 {
			continue
		}
		s += fmt.Sprintf("%s, %d|", this.indexRegion(i), this.regList[i])
	}
	return s
}

type statLoginObject struct {
	statTime     time.Time
	newCount     []int
	successCount []int
	failedCount  []int
}

func (this *statLoginObject) init() {
	this.statTime = time.Now()
	this.newCount = make([]int, maxLoginSize)
	this.successCount = make([]int, maxLoginSize)
	this.failedCount = make([]int, maxLoginSize)

	this.reset()
}

func (this *statLoginObject) reset() {
	this.statTime = time.Now()

	for i := 0; i < maxLoginSize; i++ {
		this.newCount[i] = 0
		this.successCount[i] = 0
		this.failedCount[i] = 0
	}
}

func (this *statLoginObject) getOffset() int {
	now := time.Now()
	offset := int(now.Unix() - this.statTime.Unix())

	if offset >= maxLoginSize {
		// 理论上不会出现，做容错处理
		StatMgr.outLoginStat()
		this.reset()
		offset = 0
	}
	if offset < 0 {
		offset = 0
	}
	return offset
}

type statMgr struct {
	startTime time.Time

	// 命令字统计
	cmdMap map[uint32]*statCmdObject
	// tick 统计
	tickMap map[int]*statObject
	// 函数统计，函数统计一般用 string 作为 key，统计比较频繁有可能会有性能问题，所以跟 tick 分开
	funcMap   map[string]*statObject
	funcMapEx []*statObject
	// 登录相关统计
	loginStat *statLoginObject
	Log       *log.Filter

	taskMap   map[string]*statObject
	taskMutex sync.Mutex

	goTaskMap   map[string]*statObject
	goTaskMutex sync.Mutex
}

func (this *statMgr) Init(out *log.Filter) {
	this.startTime = time.Now()
	this.Log = out

	this.cmdMap = make(map[uint32]*statCmdObject)
	this.tickMap = make(map[int]*statObject)
	this.funcMap = make(map[string]*statObject)
	this.funcMapEx = make([]*statObject, 37)
	for i, _ := range this.funcMapEx {
		this.funcMapEx[i] = &statObject{id: i + 1, callCount: 0, totalTime: 0}
		this.funcMapEx[i].init()
	}

	this.loginStat = &statLoginObject{}
	this.loginStat.init()
	this.taskMap = make(map[string]*statObject)
	this.goTaskMap = make(map[string]*statObject)

	AddTimer("StatMgr::tick30s", time.Second*30, this.tick30s)
}

func (this *statMgr) StatCmdCall(cmdId uint32, msgSize int, useTime time.Duration) {
	cmdData, ok := this.cmdMap[cmdId]
	if !ok {
		cmdData = &statCmdObject{cmdId: cmdId}
		cmdData.init()
		this.cmdMap[cmdId] = cmdData
	}

	cmdData.add(msgSize, int64(useTime)/1000, this.startTime)
}

func (this *statMgr) AddTick(id int, name string) {
	if _, ok := this.tickMap[id]; ok {
		return
	}
	tick := &statObject{id: id, name: name, callCount: 0, totalTime: 0}
	tick.init()
	this.tickMap[id] = tick
}

func (this *statMgr) DelTick(id int) {
	if _, ok := this.tickMap[id]; ok {
		delete(this.tickMap, id)
	}
}

func (this *statMgr) StatTickCall(id int, useTime time.Duration) {
	tickData, ok := this.tickMap[id]
	if !ok {
		Log.Error("not found tick. Id: %d", id)
		return
	}
	tickData.add(int64(useTime) / 1000)
}

func (this *statMgr) StatFuncCall(name string, useTime time.Duration) {
	funcData, ok := this.funcMap[name]
	if !ok {
		funcData = &statObject{id: 0, name: name, callCount: 0, totalTime: 0}
		funcData.init()
		this.funcMap[name] = funcData
	}
	funcData.add(int64(useTime) / 1000)
}

func (this *statMgr) StatTaskCall(name string, useTime time.Duration) {
	this.taskMutex.Lock()
	defer this.taskMutex.Unlock()

	funcData, ok := this.taskMap[name]
	if !ok {
		funcData = &statObject{id: 0, name: name, callCount: 0, totalTime: 0}
		funcData.init()
		this.taskMap[name] = funcData
	}
	funcData.add(int64(useTime) / 1000)
}

func (this *statMgr) StatGoTaskCall(name string, useTime time.Duration) {
	this.goTaskMutex.Lock()
	defer this.goTaskMutex.Unlock()

	funcData, ok := this.goTaskMap[name]
	if !ok {
		funcData = &statObject{id: 0, name: name, callCount: 0, totalTime: 0}
		funcData.init()
		this.goTaskMap[name] = funcData
	}
	funcData.add(int64(useTime) / 1000)
}

func (this *statMgr) StatFuncCallEx(idx int, useTime time.Duration) {
	if idx > 0 && idx <= len(this.funcMapEx) {
		this.funcMapEx[idx-1].add(int64(useTime) / 1000)
	}
}

func (this *statMgr) StatNewLogin() {
	offset := this.loginStat.getOffset()
	this.loginStat.newCount[offset]++
}

func (this *statMgr) StatFinishLogin(success bool, useTime time.Duration) {
	offset := this.loginStat.getOffset()
	if success {
		this.StatFuncCall("LoginSuccess", useTime)
		this.loginStat.successCount[offset]++
	} else {
		this.StatFuncCall("LoginFailed", useTime)
		this.loginStat.failedCount[offset]++
	}
}

func (this *statMgr) StatFinishCreate(success bool, useTime time.Duration) {
	offset := this.loginStat.getOffset()
	if success {
		this.StatFuncCall("CreateSuccess", useTime)
		this.loginStat.successCount[offset]++
	} else {
		this.StatFuncCall("CreateFailed", useTime)
		this.loginStat.failedCount[offset]++
	}
}

func (this *statMgr) tick30s(now time.Time) {
	this.outCmdStat()
	this.outTickStat()
	this.outFuncStat()
	this.outTaskStat()
	this.outGoTaskStat()
	this.outFuncStatEx()
	this.outLoginStat()
	this.startTime = time.Now()
}

func (this *statMgr) outCmdStat() {
	for _, co := range this.cmdMap {
		if co.callCount <= 0 || co.qpsObject == nil {
			continue
		}

		var name string

		enmSSCmdValue := ss.EnmCmdValue(co.cmdId)
		name = enmSSCmdValue.String()
		if len(name) <= 7 {
			enmCSCmdValue := cs.EnmCmdValue(co.cmdId)
			name = enmCSCmdValue.String()
		}

		if len(name) > 4 {
			name = name[4:]
		}
		avgTime := co.totalTime / int64(co.callCount)
		avgSize := co.totalSize / int(co.callCount)

		this.Log.Error("[cmdstat] cmd: %20s, count: %5d, totalTime: %5d, size: %5d, avgTime: %5d, avgSize: %5d, region: %s",
			name, co.callCount, co.totalTime/timeUnitConvert, co.totalSize, avgTime/timeUnitConvert, avgSize, co.region.String())

		this.Log.Error("[qpsstat] cmd: %20s, time: %d, qps: %s", name, this.startTime.Unix(), co.qpsObject.String())

		co.reset()
	}
}

func (this *statMgr) outTickStat() {
	for _, ts := range this.tickMap {
		if ts.callCount <= 0 {
			continue
		}

		avgTime := ts.totalTime / int64(ts.callCount)
		this.Log.Error("[tickstat] tick: %20s, count: %5d, totalTime: %5d, avgTime: %5d, region: %s",
			ts.name, ts.callCount, ts.totalTime/timeUnitConvert, avgTime/timeUnitConvert, ts.region.String())

		ts.reset()
	}
}

func (this *statMgr) outFuncStat() {
	for _, fobj := range this.funcMap {
		if fobj.callCount <= 0 {
			continue
		}

		avgTime := fobj.totalTime / int64(fobj.callCount)
		this.Log.Error("[funcstat] name: %20s, count: %5d, totalTime: %5d, avgTime: %5d, region: %s",
			fobj.name, fobj.callCount, fobj.totalTime/timeUnitConvert, avgTime/timeUnitConvert, fobj.region.String())

		fobj.reset()
	}
}

func (this *statMgr) outTaskStat() {
	this.taskMutex.Lock()
	taskMap := this.taskMap
	this.taskMap = make(map[string]*statObject)
	this.taskMutex.Unlock()

	for _, fobj := range taskMap {
		if fobj.callCount <= 0 {
			continue
		}

		avgTime := fobj.totalTime / int64(fobj.callCount)
		this.Log.Error("[taskstat] name: %20s, count: %5d, totalTime: %5d, avgTime: %5d, region: %s",
			fobj.name, fobj.callCount, fobj.totalTime/timeUnitConvert, avgTime/timeUnitConvert, fobj.region.String())

		fobj.reset()
	}
}

func (this *statMgr) outGoTaskStat() {
	this.goTaskMutex.Lock()
	goTaskMap := this.goTaskMap
	this.goTaskMap = make(map[string]*statObject)
	this.goTaskMutex.Unlock()

	for _, fobj := range goTaskMap {
		if fobj.callCount <= 0 {
			continue
		}

		avgTime := fobj.totalTime / int64(fobj.callCount)
		this.Log.Error("[gotaskstat] name: %20s, count: %5d, totalTime: %5d, avgTime: %5d, region: %s",
			fobj.name, fobj.callCount, fobj.totalTime/timeUnitConvert, avgTime/timeUnitConvert, fobj.region.String())

		fobj.reset()
	}
}

func (this *statMgr) outFuncStatEx() {
	for _, fobj := range this.funcMapEx {
		if fobj.callCount <= 0 {
			continue
		}

		avgTime := fobj.totalTime / int64(fobj.callCount)
		this.Log.Error("[funcstat] id: %d, count: %5d, totalTime: %5d, avgTime: %5d, region: %s",
			fobj.id, fobj.callCount, fobj.totalTime/timeUnitConvert, avgTime/timeUnitConvert, fobj.region.String())

		fobj.reset()
	}
}

func (this *statMgr) outLoginStat() {
	var (
		newLog     string = "|"
		successLog string = "|"
		failedLog  string = "|"
	)

	for i := 0; i < maxLoginSize; i++ {
		if this.loginStat.newCount[i] > 0 {
			newLog += fmt.Sprintf("%d %d|", i, this.loginStat.newCount[i])
		}
		if this.loginStat.successCount[i] > 0 {
			successLog += fmt.Sprintf("%d %d|", i, this.loginStat.successCount[i])
		}
		if this.loginStat.failedCount[i] > 0 {
			failedLog += fmt.Sprintf("%d %d|", i, this.loginStat.failedCount[i])
		}
	}

	if len(newLog) > 1 {
		// if newLog != "|" {
		this.Log.Error("newLogin: %s", newLog)
	}
	if len(successLog) > 1 {
		// if successLog != "|" {
		this.Log.Error("successLogin: %s", successLog)
	}
	if len(failedLog) > 1 {
		// if failedLog != "|" {
		this.Log.Error("failedLogin: %s", failedLog)
	}

	this.loginStat.reset()
}
