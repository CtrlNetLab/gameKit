package utils

import (
	"fmt"
	"hash/fnv"
)

// 生成结构体的哈希值
func HashStruct(v interface{}) uint32 {
	h := fnv.New32a()
	switch val := v.(type) {
	case int:
		h.Write([]byte(fmt.Sprintf("%d", val)))
	case string:
		h.Write([]byte(val))
	case []interface{}:
		for _, item := range val {
			h.Write([]byte(fmt.Sprintf("%v", item)))
		}
	case map[string]interface{}:
		for key, value := range val {
			h.Write([]byte(key))
			h.Write([]byte(fmt.Sprintf("%v", value)))
		}
	default:
		// 结构体类型
		s := fmt.Sprintf("%v", val)
		h.Write([]byte(s))
	}
	return h.Sum32()
}
