package snowflake

import (
	"fmt"
	"testing"
)

func TestGenUniqId(t *testing.T) {
	Init("2022-02-27", 1)
	mask := make(map[uint64]bool)
	for i := 0; i < 10000; i++ {
		id := GenUniqId(IDType_Player)
		if _, ok := mask[id]; ok {
			fmt.Printf("duplicate id: %d\n", id)
			return
		}
		mask[id] = true
	}
	fmt.Printf("success!\n")

	fmt.Printf("shift_seqId: %d\n", shift_seqId)
	fmt.Printf("shift_timestamp: %d\n", shift_timestamp)
	fmt.Printf("shift_serviceId: %d\n", shift_serviceId)
	fmt.Printf("shift_idType: %d\n", shift_idType)

	fmt.Printf("mask_seqId: %d\n", mask_seqId)
	fmt.Printf("mask_timestamp: %d\n", mask_timestamp)
	fmt.Printf("mask_serviceId: %d\n", mask_serviceId)
	fmt.Printf("mask_idType: %d\n", mask_idType)
}
