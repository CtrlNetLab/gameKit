package utils

import (
	"encoding/json"
	"reflect"
)

/**
* @author: Tavin
* @date: 2024/6/19
 */

// CoverObject 转换类型
func CoverObject[T interface{}](src interface{}) (T, error) {
	var t T
	dist := reflect.New(reflect.TypeOf(t).Elem()).Interface().(T)
	bytes, err := json.Marshal(src)
	if nil != err {
		return t, err
	}
	if err := json.Unmarshal(bytes, dist); nil != err {
		return t, err
	}
	return dist, nil
}

// CoverList 转换列表
func CoverList[T interface{}](src interface{}) ([]T, error) {
	distList := make([]T, 0)
	srcList := reflect.ValueOf(src)
	for i := 0; i < srcList.Len(); i++ {
		item := srcList.Index(i)
		dist, err := CoverObject[T](item.Interface())
		if nil != err {
			return distList, err
		}
		distList = append(distList, dist)
	}
	return distList, nil
}
