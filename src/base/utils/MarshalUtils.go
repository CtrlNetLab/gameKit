package utils

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type Uint64Array []uint64

func (ua Uint64Array) MarshalBinary() ([]byte, error) {
	if len(ua) == 0 {
		return []byte("{}"), nil
	}
	var data string = ""
	for i := 0; i < len(ua); i++ {
		if i == 0 {
			data = fmt.Sprintf("{%d", ua[i])
		} else {
			data = data + fmt.Sprintf(",%d", ua[i])
		}
	}
	data = data + fmt.Sprintf("}")
	return []byte(data), nil
}

func ProtoDeepCopy(dst, src IMessage) error {
	proto.Merge(dst, src)
	return nil
	/*
		buffer := make([]byte, src.Size())
		buflen, err := src.MarshalTo(buffer)
		if err != nil {
			return err
		}
		return dst.Unmarshal(buffer[:buflen])
	*/
}
