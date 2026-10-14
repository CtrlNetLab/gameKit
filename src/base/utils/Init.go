package utils

import gnet "github.com/walkon/wsgnet"

var (
	Log       ILogger
	XXTeaOpen bool
	XXTeaKey  string
	PressOpen bool
	AreaId    int32
	netLog    bool

	codec gnet.ICodec
)

func InitUtils(log ILogger, xxTeaOpen bool, xxTeaKey string, pressOpen bool, areaId int32) {
	Log = log
	XXTeaOpen = xxTeaOpen
	XXTeaKey = xxTeaKey
	PressOpen = pressOpen
	AreaId = areaId
}

func SetNetLog(able bool) {
	netLog = able
}

func IsNetLog() bool {
	return netLog
}

func SetCodec(cc gnet.ICodec) {
	codec = cc
}

func GetAreaId() int32 {
	return AreaId
}
