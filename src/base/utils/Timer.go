package utils

import (
	"time"
)

type GameTimer struct {
	Name     string
	Id       int
	Ticker   *time.Ticker
	Duration time.Duration
	Callback func(time.Time)
}

var gameTimers []*GameTimer = make([]*GameTimer, 0)

func AddTimer(name string, d time.Duration, callback func(time.Time)) *GameTimer {
	gTimer := &GameTimer{
		Name:     name,
		Id:       len(gameTimers),
		Ticker:   time.NewTicker(d),
		Duration: d,
		Callback: callback,
	}
	StatMgr.AddTick(gTimer.Id, gTimer.Name)
	gameTimers = append(gameTimers, gTimer)
	return gTimer
}

func DelTimer(gTimer *GameTimer) bool {
	for idx, timer := range gameTimers {
		if timer == gTimer {
			StatMgr.DelTick(timer.Id)
			gameTimers = append(gameTimers[:idx], gameTimers[idx+1:]...)
			return true
		}
	}
	return false
}

func ProcessTimer() {
	// statNow := time.Now()
	for _, timer := range gameTimers {
		select {
		case <-timer.Ticker.C:
			now := time.Now()
			timer.Callback(now)
			StatMgr.StatTickCall(timer.Id, TimeSince(now))
		default:
		}
	}
	// StatMgr.StatFuncCall("ProcessTimer", TimeSince(statNow))
}
