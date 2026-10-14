package snowflake

import (
	"gamekit/src/base/utils"
	"time"
)

type IDType int

// @NOTE: ID 类型位布局, 每种业务/实体类型分配一个唯一值(最大 32)。
//        新增实体 ID 时在此追加, 不要超过 IDType_Max。
const (
	IDType_Init    IDType = 0
	IDType_Player  IDType = 2
	IDType_Room    IDType = 3
	IDType_ConnIdx IDType = 4
	IDType_ChatMsg IDType = 7
	// 机器人(benchmark 压测玩家)
	IDType_Robot IDType = 12
	// @NOTE: 最大 id 不能超过此值
	IDType_Max IDType = 32
)

func Init(ymd string, serviceId int) bool {
	return sf.init(ymd, serviceId)
}

func GetIDType(id uint64) IDType {
	return IDType(id & mask_idType)
}

func GenUniqId(idType IDType) uint64 {
	return sf.genUniqId(idType)
}

func GenChatMsgId() uint64 {
	return sf.genUniqId(IDType_ChatMsg)
}
func GenChatMsgIdStr() string {
	return utils.Uint64ToString(GenChatMsgId())
}

func GenRobotId() uint64 {
	return sf.genUniqId(IDType_Robot)
}

// 暂不使用这里的方法生成AccId
// func GenAccId(accId uint64) uint64 {
// 	return ((accId << bit_idType) | uint64(IDType_Player))
// }

// 仅用于测试用途
func GenTestAccId(areaId int32) uint64 {
	seed := utils.RandInt32(1, 32767)
	accId := uint64(seed)<<shift_timestamp + uint64(areaId)
	accId = (accId << shift_serviceId) | uint64(IDType_Player)
	return accId
}

func GetTime(id uint64) time.Time {
	delta := (id >> 15) & mask_timestamp
	return sf.twepochTime.Add(time.Duration(delta) * time.Millisecond)
}

func GetDirYmdH(id uint64) string {
	t := GetTime(id)
	return t.Format("2006/01/02/15")
}

var (
	tmpSmallId uint64 = 10000
)

func SetTmpSmallId(id uint64) {
	tmpSmallId = id
}

// 生成本进程用的较小的唯一 id, 不能用到其他进程中
func GenTmpUniqId(idType IDType) uint64 {
	tmpSmallId++
	return ((tmpSmallId << bit_idType) | uint64(idType))
}
