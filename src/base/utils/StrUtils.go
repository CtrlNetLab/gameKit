package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

var (
	Sign string = "sign"
)

const (
	emailRegexString = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
)

var emailRegex = regexp.MustCompile(emailRegexString)

func IsMailFormat(email string) bool {
	if len(email) > 64 {
		return false
	}
	return emailRegex.MatchString(email)
}

// 根据 platType+userId 生成 hash 值，用于分表
func GetStrHashCode(s string) uint64 {
	var hash uint64 = 5381
	for _, k := range s {
		hash += (hash << 5) + uint64(k)
	}
	return hash
}

// 删除首位空格、以及中间多于 2 个的空格或者\t
func DeleteExtraSpace(s string) string {
	var out []byte
	var space int = 0
	for i := 0; i < len(s); i++ {
		v := s[i]
		if v != ' ' && v != '\t' {
			if len(out) > 0 && space > 0 {
				out = append(out, ' ')
			}
			out = append(out, v)
			space = 0
		} else {
			space++
		}
	}
	return string(out)
}

func StrToInt(s string) int {
	if i32, err := strconv.ParseInt(s, 10, 32); err == nil {
		return int(i32)
	}
	return 0
}

func StrToUint32(s string) uint32 {
	if i32, err := strconv.ParseInt(s, 10, 32); err == nil {
		return uint32(i32)
	}
	return 0
}

func StrToInt32(s string) int32 {
	if i32, err := strconv.ParseInt(s, 10, 32); err == nil {
		return int32(i32)
	}
	return 0
}

func StrToInt64(s string) int64 {
	if i64, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i64
	}
	return 0
}

func StrToUint64(s string) uint64 {
	if i64, err := strconv.ParseUint(s, 10, 64); err == nil {
		return i64
	}
	return 0
}

func HexToUint64(s string) uint64 {
	if i64, err := strconv.ParseUint(s, 16, 64); err == nil {
		return i64
	}
	return 0
}

func Uint64ToHex(id uint64) string {
	return fmt.Sprintf("%x", id)
}

func Int64ToString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func Uint64ToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}

func Int32ToString(id int32) string {
	return strconv.FormatInt(int64(id), 10)
}

func StrToFloat32(s string) float32 {
	if f32, err := strconv.ParseFloat(s, 32); err == nil {
		return float32(f32)
	}
	return 0.0
}

func StrToFloat64(s string) float64 {
	if f64, err := strconv.ParseFloat(s, 64); err == nil {
		return float64(f64)
	}
	return 0.0
}

func CombinePlatUser(loginType int32, userId string) string {
	return fmt.Sprintf("%d:%s", loginType, userId)
}

func GetUserCode(loginType int32, userId string) uint64 {
	return GetStrHashCode(CombinePlatUser(loginType, userId))
}

func SplitUintList(s, sep string) []uint32 {
	list := strings.Split(s, sep)
	values := make([]uint32, len(list))
	for i, v := range list {
		values[i] = uint32(StrToInt(v))
	}
	return values
}
func SplitInt32List(s, sep string) []int32 {
	list := strings.Split(s, sep)
	values := make([]int32, len(list))
	for i, v := range list {
		values[i] = StrToInt32(v)
	}
	return values
}

func SplitIntList(s, sep string) []int {
	list := strings.Split(s, sep)
	values := make([]int, len(list))
	for i, v := range list {
		values[i] = StrToInt(v)
	}
	return values
}

func SplitUint64List(s, sep string) []uint64 {
	list := strings.Split(s, sep)
	values := make([]uint64, len(list))
	for i, v := range list {
		values[i] = StrToUint64(v)
	}
	return values
}

func GetMapString(params map[string]interface{}, key string) string {
	if value, ok := params[key]; ok {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

func Md5Sum(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func CalcSign(params map[string]interface{}, secretKey string) string {
	keys := make([]string, 0, len(params))
	for k, _ := range params {
		if k == Sign {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var strParam string

	for _, k := range keys {
		value := GetMapString(params, k)
		if len(strParam) <= 0 {
			strParam = fmt.Sprintf("%s=%s", k, value)
		} else {
			strParam += fmt.Sprintf("&%s=%s", k, value)
		}
	}
	strParam += secretKey

	return Md5Sum(strParam)
}

func StringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

func Uint64ToStr(v uint64) string {
	return fmt.Sprintf("%d", v)
}

func Uint32ToStr(v uint32) string {
	return fmt.Sprintf("%d", v)
}
func IntToStr(v int) string {
	return fmt.Sprintf("%d", v)
}
func Int32ToStr(v int32) string {
	return fmt.Sprintf("%d", v)
}
func Int64ToStr(v int64) string {
	return fmt.Sprintf("%d", v)
}

func BoolToInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func Int32ToBool(i int32) bool {
	if i != 0 {
		return true
	}
	return false
}

func AsInt32(v interface{}) (int32, bool) {
	switch x := v.(type) {
	case int32:
		return x, true
	case int:
		return int32(x), true
	case int64:
		return int32(x), true
	case float64:
		return int32(x), true // 如需更严谨，可先判断是否为整数再转
	case float32:
		return int32(x), true
	case json.Number:
		if i64, err := x.Int64(); err == nil {
			return int32(i64), true
		}
	}
	return 0, false
}

func Int32ArrToStringArr(int32Arr []int32) []string {
	//将数组int32转化成字符串数组
	var strArr []string
	for _, v := range int32Arr {
		strArr = append(strArr, Int32ToString(v))
	}
	return strArr
}
