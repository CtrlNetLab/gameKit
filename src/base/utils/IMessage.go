package utils

import "google.golang.org/protobuf/proto"

type IMessage interface {
	proto.Message
	String() string
	ProtoMessage()
	Reset()
}

type IHeadMsg interface {
	proto.Message
	GetCmdId() uint32
	String() string
}
