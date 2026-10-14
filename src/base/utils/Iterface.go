package utils

type ILogger interface {
	Debug(arg0 interface{}, args ...interface{})
	Error(arg0 interface{}, args ...interface{})
}
