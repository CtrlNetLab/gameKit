package utils

import (
	"testing"
)

func TestRandomSliceWithoutSame(t *testing.T) {
	slice := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9}

	getKeyWeight := func(key int32) (int64, int32) {
		return int64(key), 10
	}

	randSlice, _ := RandomSliceWithoutSame(slice, 9, getKeyWeight)
	t.Logf("randSlice: %v", randSlice)
}

func BenchmarkRandomSliceWithoutSame(b *testing.B) {
	slice := []int32{1, 1, 3, 4, 5, 6, 7, 8, 9}

	getKeyWeight := func(key int32) (int64, int32) {
		return int64(key), 10
	}

	for i := 0; i < b.N; i++ {
		RandomSliceWithoutSame(slice, 8, getKeyWeight)
	}
}
