package snowflake

import (
	"gamekit/src/base/dao"

	"fmt"
	"sync"
	"time"
)

var (
	sf *snowFlake = &snowFlake{}
)

const (
	bit_seqId     = uint64(10)
	bit_timestamp = uint64(39)
	bit_serviceId = uint64(10)
	bit_idType    = uint64(5)
)

const (
	shift_seqId     = uint64(bit_timestamp + bit_serviceId + bit_idType)
	shift_timestamp = uint64(bit_serviceId + bit_idType)
	shift_serviceId = uint64(bit_idType)
	shift_idType    = uint64(0)
)

const (
	mask_seqId     = uint64(-1 ^ (-1 << bit_seqId))
	mask_timestamp = uint64(-1 ^ (-1 << bit_timestamp))
	mask_serviceId = uint64(-1 ^ (-1 << bit_serviceId))
	mask_idType    = uint64(-1 ^ (-1 << bit_idType))
)

/***
 * 10BIT：seqId，一毫秒能生成 1024 个 ID
 * 39BIT：时间（毫秒），能表示：17 年
 * 5BIT: idType
 * 10BIT: serviceId
 * 最终格式: |seqId(10BIT)|twepoch(39BIT)|serviceId(10BIT)|idType(5BIT)|
 */
type snowFlake struct {
	sync.Mutex

	// 启动时间，单位：毫秒
	twepoch     uint64
	twepochTime time.Time
	// 当前时间，单位：毫秒
	timestamp uint64
	seqId     uint64
	serviceId uint64
}

/***
 * @brief 初始化雪花算法
 * @param ymd: 服务器正式运营时间，确定之后就不能修改，格式: 2022-02-28
 * @param serviceId: 服务id, 范围: 0 ~ 1023
 *
 * @retval true: 初始化成功; false: 初始化失败
 */
func (this *snowFlake) init(ymd string, serviceId int) bool {
	if serviceId < 0 || uint64(serviceId) > mask_serviceId {
		return false
	}

	var year, month, day int
	fmt.Sscanf(ymd, "%d-%d-%d", &year, &month, &day)
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Now().Location())

	twepoch := uint64(t.UnixNano() / 1e6)
	now := uint64(time.Now().UnixNano() / 1e6)
	if twepoch > now {
		return false
	}

	this.twepochTime = t
	this.twepoch = twepoch
	this.timestamp = now
	this.seqId = 0
	this.serviceId = uint64(serviceId)

	return true
}

func (this *snowFlake) genUniqId(idType IDType) uint64 {
	if idType < 0 || uint64(idType) > mask_idType {
		return 0
	}

	// this.Lock()
	// defer func() { this.Unlock() }()

	now := uint64(time.Now().UnixNano() / 1e6)
	if this.timestamp == now {
		this.seqId = (this.seqId + 1) & mask_seqId
		if this.seqId == 0 {
			// sleep 1ms
			time.Sleep(time.Millisecond)
			// @TODO ryder: 告警
			return GenUniqId(idType)
		}
	} else {
		this.seqId = 0
	}

	t := now - this.twepoch
	t = t & mask_timestamp
	this.timestamp = now

	return (this.seqId << shift_seqId) | (t << shift_timestamp) | (this.serviceId << shift_serviceId) | (uint64(idType))
}

func ParseAreaId(accId uint64) int32 {
	if accId == 0 {
		return 0
	}
	areaId := int32(accId>>bit_idType) & 0x03FF
	if areaId == 0 {
		patchData := dao.PatchData.AccIdAreaPatchData
		if patchData == nil {
			return 4
		}
		// 旧版AccId预埋地区ID失败，需要设定默认地区ID
		if areaId, ok := patchData.AccIdAreaMap[accId]; ok {
			return areaId
		}
		// 默认NA
		return 4
	}

	return areaId
}
