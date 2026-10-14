package dao

import (
	"gamekit/src/base/logger"
	"gamekit/src/base/utils"
	"encoding/json"
	"os"
)

type OneAreaPatchData struct {
	AreaId int32    `json:"areaId"`
	AccIds []string `json:"accIds"`
}

type AreaAccIdPatchData struct {
	AreaAccIdMap []*OneAreaPatchData `json:"data"`
	AccIdAreaMap map[uint64]int32    `json:"-"`
}

type patchData struct {
	AccIdAreaPatchData *AreaAccIdPatchData
}

// 兼容性数据
var PatchData *patchData

func init() {
	PatchData = &patchData{}
	patchAccIdAreaFilePath := "../../binconf/configs/patchAccIdArea.json"
	content, err := os.ReadFile(patchAccIdAreaFilePath)
	if err != nil {
		logger.Log.Error("read patch accId failed. file: %s, err: %v", patchAccIdAreaFilePath, err.Error())
		return
	}

	var areaAccIdPatchData AreaAccIdPatchData
	err = json.Unmarshal(content, &areaAccIdPatchData)
	if err != nil {
		logger.Log.Error("unmarshal patch accId failed. file: %s, err: %v", patchAccIdAreaFilePath, err.Error())
		return
	}
	areaAccIdPatchData.AccIdAreaMap = make(map[uint64]int32)
	for _, oneArea := range areaAccIdPatchData.AreaAccIdMap {
		for _, accId := range oneArea.AccIds {
			areaAccIdPatchData.AccIdAreaMap[utils.StrToUint64(accId)] = oneArea.AreaId
		}
	}
	PatchData.AccIdAreaPatchData = &areaAccIdPatchData
}
