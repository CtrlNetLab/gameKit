#!/bin/bash

. ~/.bashrc

if [[ $# -ne 2 ]]; then
    echo "useage: update.sh internal(external) serverName" 
    exit 1
fi

Env=$1
ServerName=$2
UpdateType=$3

function NotifyDingDingTalk() 
{
    Result=$1
    Content=`echo $2 | sed  "s/ /_/g"`
    Hook=https://oapi.dingtalk.com/robot/send?access_token=8d8eb4fd68d2bc83fb9db6e74fb665eb57eb67e99d964e16dfeb3913a876b89b
    curl -X POST $Hook \
        -H 'Content-Type: application/json' \
        -d  '{
     "msgtype": "markdown",
     "markdown": {
         "title": "服务器更新通知",
         "text": "### 服务器更新通知: '${ServerName}'\n#### 结果 '$Result' \n#### '${Content}'"
     },
 }'
}

CurDir=`cd $(dirname $0) && pwd`
PackageDir=$CurDir/packages
RunDir=`cd $CurDir/../../ && pwd`

BranchName=`grep Branch ${CurDir}/meta.txt | awk '{print $NF}'`

if [[ $Env == "internal" ]]; then
    PackageURL=`./../redis-cli -h 172.16.0.231 -p 6379  HGET NewPackageURL $BranchName`
    Package=`echo $PackageURL | awk -F"/" '{print $NF}'`
elif [[ $Env == "external" ]]; then
    PackageURL=`./../redis-cli -h 10.162.48.2 -p 6379 -a "JFQBRLsRNIqoiwglHcW0tYwmKLpGrdHb" HGET NewPackageURL $BranchName`
    Package=`echo $PackageURL | awk -F"/" '{print $NF}'`
fi

who=`whoami`

if [[ ! -d $PackageDir ]]; then
    mkdir $PackageDir
fi

if [[ -f $PackageDir/$Package ]]; then 
    echo "Pacakge $PackageURL AllReady Updated"
    exit 0
fi


cd $PackageDir
if [[ $Env == "internal" ]]; then
    wget $PackageURL
    if [[ $? -ne 0 ]]; then
        echo "wget $PackageURL Failed"
        exit 1
    fi
elif [[ $Env == "external" ]]; then
    # TODO
    aws --profile 633659218469-aws-cli-s3-readwrite-access s3 cp $PackageURL ./
    if [[ $? -ne 0 ]]; then
        echo "aws s3 download $PackageURL Failed"
        exit 1
    fi
fi

tar xf $Package -C $RunDir
if [[ $? -ne 0 ]]; then 
    echo "UnPack $Package Failed"
    exit 1
fi

if [[ $UpdateType == "reload" ]]; then
	cd $RunDir
	bash admin.sh reload
else
    cd $RunDir
	bash admin.sh config
    if [[ $? -ne 0 ]]; then
        NotifyDingDingTalk "failed" "config failed"
        exit 1
    fi

	cd $RunDir
	bash admin.sh stop
	cd $RunDir
	bash admin.sh start
    if [[ $? -ne 0 ]]; then
        NotifyDingDingTalk "failed" "start failed"
        exit 1
    fi

    NotifyDingDingTalk "Success" $PackageURL
fi
