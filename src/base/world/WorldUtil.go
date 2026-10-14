package world

import (
	"gamekit/pkg/ss"
	"gamekit/src/base/snowflake"
	"gamekit/src/base/utils"
)

type GetAccIdFromData[T any] func(data T) uint64

func ClassfyAccId[T any](arr []T, ga GetAccIdFromData[T]) ([]uint64, map[int32][]uint64) {
	localArea := make([]uint64, 0)
	otherArea := make(map[int32][]uint64)
	for _, a := range arr {
		accId := ga(a)
		areaId := snowflake.ParseAreaId(ga(a))
		if areaId == utils.GetAreaId() {
			localArea = append(localArea, accId)
		} else {
			if _, ok := otherArea[areaId]; !ok {
				otherArea[areaId] = make([]uint64, 0)
			}
			otherArea[areaId] = append(otherArea[areaId], accId)
		}
	}

	return localArea, otherArea
}

func ClassfyAccIdArray(accIds []uint64) ([]uint64, map[int32][]uint64) {
	localArea := make([]uint64, 0)
	otherArea := make(map[int32][]uint64)
	for _, accId := range accIds {
		areaId := snowflake.ParseAreaId(accId)
		if areaId == utils.GetAreaId() {
			localArea = append(localArea, accId)
		} else {
			if _, ok := otherArea[areaId]; !ok {
				otherArea[areaId] = make([]uint64, 0)
			}
			otherArea[areaId] = append(otherArea[areaId], accId)
		}
	}

	return localArea, otherArea
}

func IsLocalAccId(accId uint64) (bool, int32) {
	areaId := snowflake.ParseAreaId(accId)
	return areaId == utils.AreaId, areaId
}

func GetGameConnIdx(gameId int32) uint64 {
	iType := ss.ServiceType_GameSvr
	return (((uint64(iType) & 0xff) << 16) | (uint64(gameId) & 0xffff))
}
