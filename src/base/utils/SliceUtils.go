package utils

import (
	"errors"
	"math/rand"
	"strings"
)

// 自定义初始的Rand，并发调用时有数据竞争问题

type SliceCompareFunc[T1 any, T2 any] func(T1, T2) bool

func IsInSlice[T1, T2 any](s []T1, e T2, comFunc SliceCompareFunc[T1, T2]) bool {
	for _, a := range s {
		if comFunc(a, e) {
			return true
		}
	}
	return false
}

func RandomSlice[T any](arr []T, num int) []T {
	ret := make([]T, 0)

	if len(arr) <= num {
		return arr
	}

	used := make(map[int]bool)
	for i := 0; i < num; i++ {
		idx := rand.Intn(len(arr))
		for used[idx] {
			idx = rand.Intn(len(arr))
		}
		used[idx] = true
		ret = append(ret, arr[idx])
	}

	return ret
}

func RandomSliceOne[T any](arr []T) T {
	if len(arr) == 0 {
		var t T
		return t
	}

	if len(arr) == 1 {
		return arr[0]
	}
	idx := rand.Intn(len(arr))
	return arr[idx]
}

func RandomSeedSliceOne[T any](arr []T, seed int64) T {
	if len(arr) == 0 {
		var t T
		return t
	}

	if len(arr) == 1 {
		return arr[0]
	}
	randomSource := rand.New(rand.NewSource(seed))
	idx := randomSource.Intn(len(arr))
	return arr[idx]
}

func RandomSourceSeedSliceOne[T any](arr []T, r rand.Source) T {
	if len(arr) == 0 {
		var t T
		return t
	}

	if len(arr) == 1 {
		return arr[0]
	}
	randomSource := rand.New(r)
	idx := randomSource.Intn(len(arr))
	return arr[idx]
}

func RandomSliceWithFilter[T1, T2 any](arr []T1, num int, filters []T2, compareFunc SliceCompareFunc[T2, T1]) ([]T1, error) {
	ret := make([]T1, 0)

	if len(arr) <= num {
		return ret, errors.New("num is larger than the length of arr")
	}

	filtedArr := make([]T1, 0)
	for _, elem := range arr {
		if IsInSlice(filters, elem, compareFunc) {
			continue
		}
		filtedArr = append(filtedArr, elem)
	}

	if len(filtedArr) < num {
		return ret, errors.New("num is larger than the length of arr after filter")
	}
	ret = getRandomElements(filtedArr, num)
	//used := make(map[int]bool)
	//for i := 0; i < num; i++ {
	//	idx := rand.Intn(len(filtedArr))
	//	for used[idx] {
	//		idx = rand.Intn(len(arr))
	//	}
	//	used[idx] = true
	//	ret = append(ret, arr[idx])
	//}
	return ret, nil
}

func getRandomElements[T1 any](arr []T1, n int) []T1 {
	shuffled := make([]T1, len(arr))
	copy(shuffled, arr)
	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled[:n]
}

func FilterSlice[T1, T2 any](arr []T1, filters []T2, compareFunc SliceCompareFunc[T2, T1]) []T1 {
	ret := make([]T1, 0)

	for _, elem := range arr {
		if !IsInSlice(filters, elem, compareFunc) {
			ret = append(ret, elem)
		}
	}

	return ret
}

// 数组A是否包含数组B的所有元素
func ContainsAllElements(A, B []int32) bool {
	// 快速失败：如果 B 的长度大于 A，直接返回 false
	if len(B) > len(A) {
		return false
	}

	// 将 A 的元素存入集合（用空结构体节省内存）
	setA := make(map[int32]struct{})
	for _, num := range A {
		setA[num] = struct{}{}
	}

	// 检查 B 的所有元素是否都在集合中
	for _, num := range B {
		if _, exists := setA[num]; !exists {
			return false
		}
	}
	return true
}

func JoinBy[T any](items []T, sep string, fn func(T) string) string {
	var builder strings.Builder
	for i, item := range items {
		if i > 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(fn(item))
	}
	return builder.String()
}
