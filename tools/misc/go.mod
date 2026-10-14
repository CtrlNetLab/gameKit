module misc

go 1.25.9

replace gamekit/pkg => ../../pkg

require (
	gamekit/pkg v0.0.0-00010101000000-000000000000
	github.com/panjf2000/gnet v1.6.7
	github.com/xxtea/xxtea-go v1.0.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/sys v0.0.0-20211204120058-94396e421777 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
