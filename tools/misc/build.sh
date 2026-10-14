#!/bin/bash

DIR=`cd $(dirname $0);pwd`
cd $DIR

## 拷贝协议到项目跟目录 
# cp ../../pkg ./ -rf

go${exe} mod tidy

go${exe} build -o misc${exe} scripts/*.go

chmod +x misc${exe}

