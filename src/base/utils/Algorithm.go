package utils

import (
	"errors"
	"math/rand"
	"sort"
)

const (
	FloatZero float32 = 0.000001
)

var (
	ErrSliceDupKey            = errors.New("slice has duplicate key")
	ErrRandomNumLessThenSlice = errors.New("random num less then slice")
	ErrZeroRandomNum          = errors.New("random num is zero")
	ErrZeroSlice              = errors.New("zero slice")
	ErrNilFunc                = errors.New("nil func")
	ErrNegativeWeight         = errors.New("negative weight")
)

func SafeAdd(a uint32, b int32) uint32 {
	if b < 0 && a <= uint32(-b) {
		return 0
	}
	return uint32(int32(a) + b)
}

func SafeMinusUInt32(a uint32, b uint32) uint32 {
	if a <= b {
		return 0
	}
	return a - b
}

func SafeMinusInt32(a int32, b int32) int32 {
	if a <= b {
		return 0
	}
	return a - b
}

func MaxInt(a int32, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func MaxUint(a uint32, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func MinUint(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func MinInt32(a int32, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt64(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}

func MinUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func IsFloatZero(v float32) bool {
	return v < FloatZero && v > -FloatZero
}

// 0~100随机，返回成功
func SelectPercent(a int) bool {
	return RandInt(0, 100) <= a
}

func SelectPercent10W(a int) bool {
	return RandInt(0, 100000) <= a
}

func RandFloat(a, b float32) float32 {
	if a > b {
		a, b = b, a
	}
	return a + (b-a)*rand.Float32()
}

func RandFloat64(a, b float64) float64 {
	if a > b {
		a, b = b, a
	}
	return a + (b-a)*rand.Float64()
}

// 生成 [a, b] 之间的随机数
func RandInt(a, b int) int {
	if a == b {
		return a
	}
	if a > b {
		a, b = b, a
	}
	if a < 0 {
		return a + RandInt(0, b-a)
	}
	return a + rand.Intn(b-a+1)
}

func RandInt32(a, b int32) int32 {
	if a == b {
		return a
	}
	if a > b {
		a, b = b, a
	}
	if a < 0 {
		return a + RandInt32(0, b-a)
	}
	return a + rand.Int31n(b-a+1)
}

func RandInt64(a, b int64) int64 {
	if a == b {
		return a
	}
	if a > b {
		a, b = b, a
	}
	if a < 0 {
		return a + RandInt64(0, b-a)
	}
	return a + rand.Int63n(b-a+1)
}

var (
	BASESTR string = "0123456789abcdefghijklnmopqrstuvwxyzABCDEFGHIJKLNMOPQRSTUVWXYZ~!@#$%^&*()-_=+[]{}\\|;:'\",<.>/?"
)

func RandBytes(n int) []byte {
	lenBASESTR := len(BASESTR)
	if n <= 0 {
		n = rand.Intn(lenBASESTR)
	}
	if n <= 0 {
		return []byte{}
	}
	if n > lenBASESTR {
		n = lenBASESTR
	}
	idxs := RandNK(lenBASESTR, n)
	results := make([]byte, n)
	for i, idx := range idxs {
		results[i] = BASESTR[idx]
	}
	return results
}

func RandNK(n, k int) []int {
	if n <= 0 || k <= 0 {
		return []int{}
	}

	randnk := make([]int, n)
	array := make([]int, k)
	for i := 0; i < n; i++ {
		randnk[i] = i
	}

	if k > n {
		k = n
	}

	for i := 0; i < k; i++ {
		p := RandInt(i, n-1)
		t := randnk[i]
		randnk[i] = randnk[p]
		randnk[p] = t
		array[i] = randnk[i]
	}

	return array
}

func GetRandomIDByWeight[T int | int32 | uint32 | int64 | uint64](mapWeight map[T]int32) T {
	if len(mapWeight) <= 0 {
		return 0
	}

	var (
		weightSum int32 = 0
		rankValue int32 = 0
		findRank  int32 = 0
	)

	for _, v := range mapWeight {
		weightSum += v
	}

	rankValue = int32(RandInt(1, int(weightSum)))
	for key, v := range mapWeight {
		findRank += v
		if findRank >= rankValue {
			return key
		}
	}

	return 0
}

func GetRandomIndexByWeight[T int | int32 | uint32 | int64 | uint64](weights []T) int {
	if len(weights) <= 0 {
		return 0
	}

	var (
		weightSum int = 0
		rankValue int = 0
		findRank  int = 0
	)

	for _, v := range weights {
		weightSum += int(v)
	}

	rankValue = int(RandInt(1, int(weightSum)))
	for i, v := range weights {
		findRank += int(v)
		if findRank >= rankValue {
			return i
		}
	}

	return 0
}

func UniqArray[T comparable](array []T) []T {
	out := make([]T, 0)
	set := make(map[T]struct{})
	for _, v := range array {
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

func Binornd(n int, w []uint32) []int {
	total := int(0)
	result := make([]int, len(w))
	pList := make([]int, len(w))

	for idx, value := range w {
		total += int(value)
		pList[idx] = total
	}

	if total == 0 {
		return result
	}

	for i := 0; i < n; i++ {
		// [1, total]
		seed := rand.Intn(total) + 1
		for idx, value := range pList {
			if seed <= value {
				result[idx]++
				break
			}
		}
	}
	return result
}

// 根据ID，weight随机一个ID
func RandomWeight(idWeight map[int32]int32) int32 {
	if len(idWeight) == 0 {
		return 0
	}

	var totalWeight int32 = 0
	for _, weight := range idWeight {
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}
	randWeight := rand.Intn(int(totalWeight))
	randWeight++
	idList := make([]int, 0)
	for id := range idWeight {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)

	var tmpWeight int32 = 0
	for idx, id := range idList {
		tmpWeight += idWeight[int32(id)]
		if tmpWeight >= int32(randWeight) || idx+1 == len(idList) {
			return int32(id)
		}
	}

	return 0
}

// 根据 key-weight 映射随机选择一个 key（泛型版本）
// K: key类型（支持可排序的类型）
// W: weight类型（支持整数类型）
func RandomWeightT[K comparable, W int | int32 | uint32 | int64 | uint64](keyWeight map[K]W) K {
	if len(keyWeight) == 0 {
		var zero K
		return zero
	}

	// 计算总权重（仅统计正权重）
	var totalWeight int64 = 0
	for _, weight := range keyWeight {
		if weight <= 0 {
			continue
		}
		totalWeight += int64(weight)
	}

	// 没有正权重则返回零值
	if totalWeight <= 0 {
		var zero K
		return zero
	}

	// 生成随机权重值 [1, totalWeight]，避免从 int64 缩小为 int
	randWeight := rand.Int63n(totalWeight) + 1

	// 累加权重直到找到目标区间（仅考虑正权重）
	var accWeight int64 = 0
	for key, weight := range keyWeight {
		if weight <= 0 {
			continue
		}
		accWeight += int64(weight)
		if accWeight >= randWeight {
			return key
		}
	}

	// 理论上不会到这里，但作为fallback返回第一个key
	for key := range keyWeight {
		return key
	}

	var zero K
	return zero
}

type GetSlicenElemKeyWeight[SliceType any] func(elem SliceType) (int64, int32)

func RandomSliceWithoutSame[SliceType any](elemSlice []SliceType, num int32, fn GetSlicenElemKeyWeight[SliceType]) (randSlice []SliceType, err error) {
	if len(elemSlice) == 0 {
		return randSlice, ErrZeroSlice
	}

	if num <= 0 {
		return elemSlice, ErrZeroRandomNum
	}

	if fn == nil {
		return elemSlice, ErrNilFunc
	}

	if len(elemSlice) < int(num) {
		return elemSlice, nil
	}

	type ElemWeight struct {
		Elem   SliceType
		Weight int32
	}

	// 构建带权重的元素池，同时校验 key 唯一性（仅对有效权重元素）
	keySet := make(map[int64]struct{}, len(elemSlice))
	pool := make([]ElemWeight, 0, len(elemSlice))
	var totalWeight int32
	for _, elem := range elemSlice {
		key, weight := fn(elem)
		if weight <= 0 {
			continue
		}
		if _, dup := keySet[key]; dup {
			return nil, ErrSliceDupKey
		}
		keySet[key] = struct{}{}
		pool = append(pool, ElemWeight{Elem: elem, Weight: weight})
		totalWeight += weight
	}

	if totalWeight <= 0 {
		return elemSlice, ErrNegativeWeight
	}

	// 过滤零权重后，有效元素可能不足 num
	if len(pool) < int(num) {
		return elemSlice, nil
	}

	randSlice = make([]SliceType, 0, num)
	for num > 0 {
		if totalWeight <= 0 {
			// 剩余元素全为零权重，无法继续选取
			break
		}
		// 在 [1, totalWeight] 范围内随机
		target := rand.Int31n(totalWeight) + 1
		var acc int32
		for i, ew := range pool {
			acc += ew.Weight
			if acc >= target {
				randSlice = append(randSlice, ew.Elem)
				totalWeight -= ew.Weight
				// 交换到末尾并收缩切片，O(1) 删除
				pool[i] = pool[len(pool)-1]
				pool = pool[:len(pool)-1]
				break
			}
		}
		num--
	}
	return randSlice, nil
}

// 求map与slice的交集
func TakeMapSliceIntersection[T comparable, V any](m map[T]V, s []T) []T {
	intersection := make([]T, 0)
	for _, elem := range s {
		if _, ok := m[elem]; ok {
			intersection = append(intersection, elem)
		}
	}

	return intersection
}

// 求两个slice的交集
func SliceIntersection[T comparable](a, b []T) []T {
	intersection := make([]T, 0)
	setB := make(map[T]struct{})
	for _, elem := range b {
		setB[elem] = struct{}{}
	}

	for _, elem := range a {
		if _, ok := setB[elem]; ok {
			intersection = append(intersection, elem)
		}
	}
	return intersection
}

// 对map value排序，并返回排序后的Key
func SortMapByValue[T comparable](m map[T]int32, desc bool) []T {
	if len(m) == 0 {
		return []T{}
	}

	type kv struct {
		Key   T
		Value int32
	}
	ss := make([]*kv, 0, len(m))
	for k, v := range m {
		ss = append(ss, &kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		if desc {
			return ss[i].Value > ss[j].Value
		}
		return ss[i].Value < ss[j].Value
	})

	keys := make([]T, 0, len(ss))
	for _, kv := range ss {
		keys = append(keys, kv.Key)
	}
	return keys
}
