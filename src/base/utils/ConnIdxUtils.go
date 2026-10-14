package utils

import (
	ss "gamekit/pkg/ss"
)

const (
	MaxServiceConnIdx uint64 = uint64(1) << 24
	AccIdOffset       uint64 = uint64(1e14)
	InvalidContext    uint64 = uint64(1<<64 - 1)
)

func ServiceConnIdx(stype ss.ServiceType, idx int32) uint64 {
	// iType < 2^8, idx < 2^16
	return (((uint64(stype) & 0xff) << 16) | (uint64(idx) & 0xffff))
}

func ServiceTypeIdx(connIdx uint64) (itype, idx int32) {
	itype = int32((connIdx >> 16) & 0xff)
	idx = int32(connIdx & 0xffff)
	return
}

func IsGameConnIdx(connIdx uint64) bool {
	itype, _ := ServiceTypeIdx(connIdx)
	return itype == int32(ss.ServiceType_GameSvr)
}

func ServiceTypeIdxName(connIdx uint64) (string, uint32) {
	itype := ss.ServiceType((connIdx >> 16) & 0xff)
	idx := uint32(connIdx & 0xffff)
	return itype.String(), idx
}

func GetGameConnIdx(serviceId uint32) uint64 {
	iType := ss.ServiceType_GameSvr
	return (((uint64(iType) & 0xff) << 16) | (uint64(serviceId) & 0xffff))
}
