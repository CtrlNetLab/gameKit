#!/bin/bash

DIR=`cd $(dirname $0);pwd`
cd $DIR

## 拷贝协议相关的到项目跟目录 
# cp ../../pkg ./ -rf

PROJECT_NAME=`basename ${DIR}`
PROJECT_BIN=${PROJECT_NAME}${exe}
PROJECT_HOME=${DIR}/../../run/${PROJECT_NAME}/bin

# echo "PROJECT_BIN: ${PROJECT_BIN}, PROJECT_HOME: ${PROJECT_HOME}"

go${exe} mod tidy

# go${exe} build -o ${PROJECT_BIN} scripts/main.go
go${exe} build -gcflags "all=-N -l" -o ${PROJECT_BIN} scripts/main.go
let retCode=$?

if [ -e ${PROJECT_BIN} ]
then
    if [ "x${exe}" == "x" ] 
    then
        chmod +x ${PROJECT_BIN};
    fi

    mv ${PROJECT_BIN} ${PROJECT_HOME};
fi

exit $retCode

