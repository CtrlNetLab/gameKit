package utils

import (
	ss "gamekit/pkg/ss"
	"fmt"

	"google.golang.org/protobuf/proto"
)

type AccIdObject struct {
	AccId uint64
}

func (this *AccIdObject) String() string {
	return fmt.Sprintf("%d", this.AccId)
}

func CreateAccIdObject(accId uint64) *AccIdObject {
	return &AccIdObject{AccId: accId}
}

type Event struct {
	// @NOTE: 有些地方不能用 Player，否则会造成循环引用导致编译出错
	IObj IObjString

	// 事件 id
	EventId ss.EventType
	// 参数
	Param []uint32
	// Param []uint64
	ParamI interface{}
	// EvnParam *ss.EvnParam
	EventParam proto.Message
}

func (this *Event) String() string {
	return fmt.Sprintf("{IObj: %s, EventId: %d, Param: %+v, EventParam: %+v}", this.IObj.String(), this.EventId, this.Param, this.EventParam)
}

type IEventHandler func(event *Event)

var (
	eventHandlerMap map[ss.EventType][]IEventHandler = make(map[ss.EventType][]IEventHandler)
)

func DispatchEvent(event *Event) {
	if event == nil {
		return
	}

	handlers, ok := eventHandlerMap[event.EventId]
	if !ok {
		Log.Error("not found event handler. Event: %s", event.String())
		return
	}

	if IsNetLog() {
		Log.Debug("dispatch event. Event: %s", event.String())
	}
	for _, handler := range handlers {
		handler(event)
	}
}

func AddEventHandler(eid ss.EventType, handler IEventHandler) {
	if value, ok := eventHandlerMap[eid]; ok {
		eventHandlerMap[eid] = append(value, handler)
	} else {
		handlers := make([]IEventHandler, 0)
		handlers = append(handlers, handler)
		eventHandlerMap[eid] = handlers
	}
}
