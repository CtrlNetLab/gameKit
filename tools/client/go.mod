module client

go 1.25.9

replace gamekit/pkg => ../../pkg

require (
	gamekit/pkg v0.0.0-00010101000000-000000000000
	github.com/gogo/protobuf v1.3.2
	github.com/panjf2000/gnet v1.6.7
	github.com/xxtea/xxtea-go v1.0.0
	golang.org/x/net v0.53.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.19.1 // indirect
	golang.org/x/sys v0.43.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
