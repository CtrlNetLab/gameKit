package utils

import (
	"fmt"
	"testing"
)

func TestRandSlice(t *testing.T) {
	s := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	ret := RandomSlice(s, 1)

	fmt.Printf("ret: %v\n", ret)
}
