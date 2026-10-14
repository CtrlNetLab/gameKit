#!/bin/bash
DIR=`cd $(dirname $0);pwd`
cd $DIR

## 拷贝协议到项目跟目录 
# cp ../../pkg ./ -rf

rm -f pkg/database/db_characterdata.go pkg/database/db_convert.go

go${exe} mod tidy

go${exe} build -o db${exe} scripts/main.go

chmod +x db${exe}
sleep 1

