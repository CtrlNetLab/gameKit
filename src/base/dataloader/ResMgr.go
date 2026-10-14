package dataloader

import (
	"gamekit/pkg/configs"
	"gamekit/src/base/utils"
)

func GetArgsConfigInt64sByName(name string) []int64 {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return nil
	}

	var intValues []int64
	for _, v := range object.Val {
		intValues = append(intValues, utils.StrToInt64(v))
	}
	return intValues
}

func GetArgsConfigInt32sByName(name string) []int32 {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return nil
	}

	var intValues []int32
	for _, v := range object.Val {
		intValues = append(intValues, utils.StrToInt32(v))
	}
	return intValues
}

func GetArgsConfigInt32ByName(name string) int32 {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return 0
	}
	return utils.StrToInt32(object.Val[0])
}

func GetArgsConfigFirstValueByName(name string) string {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return ""
	}
	return object.Val[0]
}

func GetArgsConfigFirstIntValueByName(name string) int {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return 0
	}
	return utils.StrToInt(object.Val[0])
}

func GetArgsConfigFirstFloat64ValueByName(name string) float64 {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return 0
	}
	return utils.StrToFloat64(object.Val[0])
}

func GetArgsConfigFirstInt32ValueByName(name string) int32 {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return 0
	}
	return utils.StrToInt32(object.Val[0])
}

func GetArgsConfigFirstInt64ValueByName(name string) int64 {
	object := configs.ConfigInitArgsConfigResMgr.Get(name)
	if object == nil || len(object.Val) <= 0 {
		return 0
	}
	return utils.StrToInt64(object.Val[0])
}

func GetArgsConfigValueByName(name string) []string {
	argsConfig := configs.ConfigInitArgsConfigResMgr.Get(name)
	if argsConfig != nil {
		return argsConfig.Val
	}
	return []string{}
}
