package utils

import "strings"

// 左移按位取或
func LeftMoveBitwiseOr(left, right int32) int64 {
	return int64(left)<<32 | int64(right)
}

func GetFixedOrRangeNum(args string) int32 {
	numRange := strings.Split(args, "~")
	if len(numRange) == 0 {
		return StrToInt32(args)
	} else if len(numRange) == 2 {
		return int32(RandInt(int(StrToInt32(numRange[0])), int(StrToInt32(numRange[1]))))
	} else if len(numRange) == 1 {
		return StrToInt32(numRange[0])
	}
	return 0
}
