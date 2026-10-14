package utils

import (
	"sync"
	"time"
	// "gamekit/pkg/oss"
	// "gamekit/src/base/osslog"
)

var (
	WorkerPool *workerPool = &workerPool{}
)

type Task struct {
	// 任务 id
	taskId uint64
	// 任务分流依据，例如: AccId
	primary uint64

	// 任务创建时间
	createTm time.Time
	// 超过多久之后不处理默认丢弃，避免引起雪崩，不设置默认60秒
	timeOut int64

	// 任务
	task func()

	taskName string
}

func (this *Task) doTask() {
	defer func() {
		err := recover()
		if err != nil {
			// osslog.OSS.ElkAlarm(oss.EnmElk_Panic_Err, 0, osslog.OSS.PanicTrace(err))
		}
	}()
	this.task()
}

type Worker struct {
	mutex *sync.RWMutex
	cond  *sync.Cond
	idx   int
	queue []Task
}

func (w *Worker) Init(idx int) {
	w.mutex = new(sync.RWMutex)
	w.cond = sync.NewCond(w.mutex)
	w.idx = idx
	w.queue = make([]Task, 0)
}

func (w *Worker) run() {
	for {
		w.cond.L.Lock()
		for len(w.queue) == 0 {
			w.cond.Wait()
		}
		t := w.queue[0]
		w.queue = w.queue[1:]
		ts := len(w.queue)
		w.cond.L.Unlock()

		now := time.Now()
		createUnix := t.createTm.Unix()

		if netLog {
			Log.Error("%d|start process task. task: {T: %d, P: %d, C: %d}, now: %d, task.size: %d", w.idx, t.taskId, t.primary, createUnix, now.Unix(), ts)
		}

		// @TODO 处理超时回调
		t.doTask()
		elapsed := TimeSince(now)
		StatMgr.StatTaskCall(t.taskName, elapsed)

		if netLog {
			usems := elapsed / 1e6
			Log.Error("%d|finish process task. task: {T: %d, P: %d, C: %d}, cost: %d(ms), task.size: %d", w.idx, t.taskId, t.primary, createUnix, usems, ts)
		}
	}
}

func (w *Worker) addTask(primary, taskId uint64, taskName string, t func()) {
	now := time.Now()
	if netLog {
		Log.Error("%d|add task: %s, Primary: %d, TaskId: %d, now: %d", w.idx, taskName, primary, taskId, now.Unix())
	}
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.queue = append(w.queue, Task{taskId: taskId, primary: primary, createTm: now, timeOut: 60, taskName: taskName, task: t})
	w.cond.Signal()
}

type taskCallback struct {
	callbackName string
	callback     func()
}

type workerPool struct {
	poolSize int
	taskId   uint64
	worker   []*Worker

	cbMutex          *sync.RWMutex
	taskCallbackList []taskCallback
	// callback []func()
}

func (wp *workerPool) Init(poolSize int) {
	wp.poolSize = poolSize
	wp.taskId = 0
	wp.worker = make([]*Worker, 0)

	wp.cbMutex = new(sync.RWMutex)
	// wp.callback = make([]func(), 0)
	wp.taskCallbackList = make([]taskCallback, 0)

	for i := 0; i < wp.poolSize; i++ {
		w := &Worker{}
		w.Init(i)
		wp.worker = append(wp.worker, w)
	}
	wp.run()
}

func (wp *workerPool) run() {
	for i := 0; i < wp.poolSize; i++ {
		go wp.worker[i].run()
	}
}

func (wp *workerPool) AddTask(primary uint64, taskName string, t func()) {
	if wp.poolSize <= 0 {
		return
	}
	idx := int32(primary % uint64(wp.poolSize))
	wp.taskId++
	taskId := wp.taskId
	wp.worker[idx].addTask(primary, taskId, taskName, t)
}

func (wp *workerPool) AddCallback(cbName string, cb func()) {
	tcb := taskCallback{callbackName: cbName, callback: cb}
	wp.cbMutex.Lock()
	wp.taskCallbackList = append(wp.taskCallbackList, tcb)
	// wp.callback = append(wp.callback, cb)
	lenCB := len(wp.taskCallbackList)
	wp.cbMutex.Unlock()

	if netLog {
		Log.Error("add %s callback length: %d", cbName, lenCB)
	}
}

func (wp *workerPool) Callback() {
	now := time.Now()
	defer func() {
		StatMgr.StatFuncCall("WorkerPool::Callback", TimeSince(now))
	}()
	for {
		wp.cbMutex.Lock()
		if len(wp.taskCallbackList) <= 0 {
			wp.cbMutex.Unlock()
			return
		}
		tcb := wp.taskCallbackList[0]
		wp.taskCallbackList = wp.taskCallbackList[1:]
		lenCB := len(wp.taskCallbackList)
		wp.cbMutex.Unlock()

		startTime := time.Now()
		tcb.callback()
		elapsed := TimeSince(startTime)
		StatMgr.StatFuncCall("cb:"+tcb.callbackName, elapsed)

		if netLog {
			Log.Error("process callback. name: %s, length: %d", tcb.callbackName, lenCB)
		}
	}
}

// NOTE: 只用于延迟性特别高的任务
func (wp *workerPool) Do(primary uint64, taskName string, t func()) {
	now := time.Now()
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				Log.Error("do task failed. task: %s, Primary: %d, err: %v", taskName, primary, err)
			}
		}()
		if netLog {
			Log.Error("start do task. task: %s,  Primary: %d", taskName, primary)
		}
		t()
		elapsed := TimeSince(now)
		StatMgr.StatGoTaskCall(taskName, elapsed)
		if netLog {
			usems := elapsed / 1e6
			Log.Error("do task finish. task: %s, Primary: %d, cost: %d(ms)", taskName, primary, usems)
		}
	}()
}
