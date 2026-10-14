#!/bin/bash
DIR=`cd $(dirname $0);pwd`
cd $DIR

## 拷贝协议到项目跟目录 
# cp ../../pkg ./ -rf

rm -f ../../pkg/ss/service_register.pb.go
rm -f ../../pkg/cs/client_register.pb.go

go${exe} mod tidy

go${exe} build -o register${exe} scripts/main.go

chmod +x register${exe}
sleep 1
