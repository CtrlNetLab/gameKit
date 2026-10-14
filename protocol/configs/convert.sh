#!/bin/bash
ProjectHome=`cd $(dirname $0) && cd ../.. && pwd`
ConfigsDir=${ProjectHome}/protocol/configs
PBOutDir=${ProjectHome}/pkg/

if [ ! -d ${PBOutDir} ]
then
    mkdir -p ${PBOutDir}
fi

cd ${ProjectHome}/protocol/
protoc${exe} --go_out=paths=source_relative:${PBOutDir} configs/*.proto

cd ${ConfigsDir}
./gen-primary.sh

cp ResLoadMgr.go.tmpl ${PBOutDir}/configs/ResLoadMgr.go
