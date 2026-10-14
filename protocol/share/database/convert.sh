#!/bin/bash

ProjectHome=`cd $(dirname $0)/../../../;pwd`
DIR=`cd $(dirname $0);pwd`
cd $DIR/..

## proto 配置的目录是: gamekit
outdir=$ProjectHome/pkg/
if [ ! -d ${outdir} ]
then
    mkdir -p ${outdir}
fi

function convert_database()
{
    echo "convert database"
    protoc${exe} --go_out=paths=source_relative:${outdir} database/*.proto
    ## 生成数据库相关内容 
    cd ${ProjectHome}/tools/db
    ./build.sh
    ./db${exe}
}

convert_database

cd ${ProjectHome}/protocol

