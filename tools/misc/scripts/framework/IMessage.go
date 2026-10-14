package framework

type IMessage interface {
	Marshal() (dAtA []byte, err error)
	MarshalTo(dAtA []byte) (int, error)
	String() string
	Unmarshal(dAtA []byte) error
	Size() (n int)
}
