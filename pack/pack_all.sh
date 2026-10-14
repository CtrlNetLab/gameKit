#!/bin/bash

source ~/.bashrc

# 判断参数个数
if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <pack_type> <config_dir>"
    exit 1
fi

PackType=$1
ConfigDir=$2

. common_function.sh
. common_args.sh

cd $ProjectDir
# 注：别在自己的开发服执行打包脚本， 会重置本地所有修改
git reset --hard
bash AllMake.sh pull
if [ $? -ne 0 ]; then 
    NotifyDingDingTalk "failed" "AllMake.sh pull Failed"
    exit 1
fi

GitCommitId=`git log -n 1 --format=%h --abbrev-commit`
TarBall=${GameName}_${BranchName}_${ConfigDir}_${GitCommitId}_`date "+%Y_%m_%d_%H_%M_%S"`.tar.gz 

# 如果策划表有更新，执行buildconfig
# 1: 需要构建配置表
needBuildConfig=1
BuildConfig $needBuildConfig $ConfigDir
Build $?

# 清理临时存放目录
if [ ! -d $TmpPackDir ]; then 
    mkdir $TmpPackDir
else
    rm -rf $TmpPackDir
    mkdir $TmpPackDir
fi

for Process in ${GameProcess[@]} 
do 
    PackSvr $Process
done
PackSvr "control"

CodeAndConfigBranch="${BranchName}-${ConfigDir}"
WriteMeta ${CodeAndConfigBranch}
PackDataConfig
PackConfDir
PackTools
PackScripts
PackRedis

if [ ! -d $SaveDir ]; then
    mkdir $SaveDir
fi

cd $PackageDir/$GameName
tar cfz $SaveDir/$TarBall `ls`
if [[ $? -ne 0 ]]; then
    NotifyDingDingTalk "failed" "create tar ball faile"
    exit 1
fi
rm -rf $GameName

OnPackSuccess

case $PackType in 
    "internal")
        NewPackageURL="http://<your-internal-host>:8088/$TarBall"
        redis-cli -h <your-redis-host> -p 6379  HSET NewPackageURL ${CodeAndConfigBranch}  $NewPackageURL
        echo "pack succeess. package url: ${NewPackageURL}"
        NotifyDingDingTalk "success" $NewPackageURL
    ;;
    "external")
        NewPackageURL="s3://<your-s3-bucket>/$TarBall"
        retryNum=0
        while true; do 
            aws --profile <your-aws-profile> s3 cp $SaveDir/$TarBall s3://<your-s3-bucket>/
            if [ $? -eq 0 ];then
                NotifyDingDingTalk "success" "s3://<your-s3-bucket>/$TarBall"
                redis-cli -h <your-redis-host> -p 6379 -a <your-redis-password> HSET NewPackageURL ${CodeAndConfigBranch}  $NewPackageURL
                break
            else
                retryNum=$((retryNum+1))
                if [ $retryNum -gt 5 ]; then
                    NotifyDingDingTalk "s3 upload failed" "retry over max times"
                    echo "upload tar ball failed. retry times ${retryNum}"
                    exit 1
                fi
                sleep 3
            fi
        done
    ;;
esac
