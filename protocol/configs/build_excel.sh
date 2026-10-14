#!/bin/bash
ProjectHome=`cd $(dirname $0) && cd ../.. && pwd`
ConfigsDir=${ProjectHome}/protocol/configs
PBOutDir=${ProjectHome}/pkg/

if [ ! -d ${PBOutDir} ]
then
    mkdir -p ${PBOutDir}
fi

if [[ ! -f ${ProjectHome}/tools/excelToProtobuf/excelToProtobuf ]]; then 
    cd ${ProjectHome}/tools/excelToProtobuf/
    go build 
fi

BranchName=`git rev-parse --abbrev-ref HEAD`
if [ $# -ne 0 ]; then
    ConfigDir="配置表$1"
else 
    ConfigDir="配置表dev"
fi

# TODO 参数指定地区目录

echo "分支：${BranchName}，使用配置目录：${ConfigDir}"
echo "${ProjectHome}/tools/excelToProtobuf/excelToProtobuf --excelPath ${ConfigsDir}/resourse/${ConfigDir}/NA/ --dataPath ${ProjectHome}/run/binconf/ --protoPath ${ProjectHome}/protocol/"
${ProjectHome}/tools/excelToProtobuf/excelToProtobuf --excelPath ${ConfigsDir}/resourse/${ConfigDir}/NA/ --dataPath ${ProjectHome}/run/binconf/ --protoPath ${ProjectHome}/protocol/
${ProjectHome}/tools/excelToProtobuf/excelToProtobuf --excelPath ${ConfigsDir}/resourse/${ConfigDir}/language/ --excelName "Language.xlsx" --dataPath ${ProjectHome}/run/binconf/ --protoPath ${ProjectHome}/protocol/
echo "分支：${BranchName}，使用配置目录：${ConfigDir}"
