#!/bin/bash

ProjectDir=`cd $(dirname $0) && cd .. && pwd`
. ${ProjectDir}/run/admin.sh none
ProjectDirName=`basename $ProjectDir`
BranchName=`git rev-parse --abbrev-ref HEAD`
#只对run目录打包 
PackDir=$ProjectDir/run
GameName=gs
TmpPackDir=$ProjectDir/pack/${GameName}
GameProcess=("relaysvr" "gamesvr" "matchsvr" "roomsvr" "miscsvr" "globalsvr")
PackageDir=$ProjectDir/pack
ConfigVersionFile=${ProjectDir}/.configVersionFile
VersionFile=/tmp/$ProjectDirName_${BranchName}_`whoami`_versionInfo
GitCommitId=`git log -n 1 --format=%h --abbrev-commit`
TarBall=${GameName}_${BranchName}_${GitCommitId}_`date "+%Y_%m_%d_%H_%M_%S"`.tar.gz 
Docker=/usr/bin/docker
SaveDir=/data/${GameName}/
