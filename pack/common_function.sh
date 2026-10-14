#!/bin/bash


function PackSvr()
{
    SvrName=$1

    TmpSvr=${TmpPackDir}/$SvrName
    if [ ! -d $TmpSvr ]; then 
        mkdir $TmpSvr
    fi

    TmpSvrBin=$TmpSvr/bin
    if [ ! -d $TmpSvrBin ]; then 
        mkdir $TmpSvrBin
    fi
    
    
    cp $PackDir/$SvrName/bin/$SvrName $TmpSvrBin
    if [[ $SvrName != "control" ]]; then 
        cp $PackDir/$SvrName/bin/*.sh $TmpSvrBin
    fi
    mkdir $TmpSvr/log

    PackConfig $SvrName
}

function PackSvrConfig() 
{
    SvrName=$1

    TmpSvr=${TmpPackDir}/$SvrName
    if [ ! -d $TmpSvr ]; then 
        mkdir $TmpSvr
    fi

    TmpSvrBin=$TmpSvr/bin
    if [ ! -d $TmpSvrBin ]; then 
        mkdir $TmpSvrBin
    fi
    
    cp $PackDir/$SvrName/bin/*.sh $TmpSvrBin
    if [[ $SvrName == "control" ]]; then 
        cp $PackDir/$SvrName/bin/*.tmpl $TmpSvrBin
    fi
    mkdir $TmpSvr/log

    PackConfig $SvrName
}

function PackConfig()
{
    SvrName=$1

    TmpSvrConf=$TmpSvr/conf
    if [ ! -d $TmpSvrConf ]; then 
        mkdir $TmpSvrConf
    fi
    cp $PackDir/$SvrName/conf/*.tmpl $TmpSvrConf
}

function PackDataConfig()
{
    cp -r $PackDir/binconf $TmpPackDir/binconf
}

function PackConfDir() 
{
    cp -r ${PackDir}/config ${TmpPackDir}/config
}

function PackTools()
{
    TmpTools=$TmpPackDir/tools
    if [ ! -d $TmpTools ]; then 
        mkdir $TmpTools
        mkdir $TmpTools/db
        mkdir $TmpTools/script
        mkdir $TmpTools/update
	mkdir $TmpTools/update/packages
    else
	if [ ! -d $TmpTools/db ]; then
	    mkdir $TmpTools/db
	fi
	if [ ! -d $TmpTools/script ]; then
	    mkdir $TmpTools/script
	fi
        if [ ! -d $TmpTools/update ]; then
	   mkdir $TmpTools/update
	   mkdir $TmpTools/update/packages
	elif [ ! -d $TmpTools/update/packages ]; then
	   mkdir $TmpTools/update/packages
	fi
    fi

    cp $PackDir/tools/redis-cli $TmpTools/
    cp $PackDir/tools/db/*.sh $TmpTools/db
    cp $PackDir/tools/db/*.sql $TmpTools/db 
    cp $PackDir/tools/db/*.tmpl $TmpTools/db
    cp $PackDir/tools/script/*.sh $TmpTools/script
    cp $PackDir/tools/update/* $TmpTools/update

    cp $PackDir/admin.sh $TmpPackDir
}

function PackRedis()
{
    cp -r ${PackDir}/redislua ${TmpPackDir}/redislua
}

function PackScripts() 
{
    TmpScriptDir=$TmpPackDir/script
    if [ ! -d $TmpScriptDir ]; then 
        mkdir $TmpScriptDir
    fi
    cp $PackDir/script/*.py $TmpScriptDir
    #cp $PackDir/script/GetHostIP $TmpScriptDir
    cp $PackDir/config.py $TmpPackDir
}

function Build()
{
    needBuild=$1
    if [[ ${needBuild} -ne 0 ]]; then
        exit 0
    fi

    Proto

    cd $ProjectDir
    bash AllMake.sh build 
    if [ $? -ne 0 ]; then 
        NotifyDingDingTalk "failed" "AllMake.sh build Failed"
        exit 1
    fi
}

function OnPackSuccess() 
{
    # 成功构建版本写入记录文件
    cd $ProjectDir
    echo `git log -n 1 --format=%h --abbrev-commit` > $VersionFile 

    echo "build package success. ${SaveDir}${TarBall}"
}

function BuildConfig() 
{
    NeedBuildConfig=$1
    cd $ProjectDir
    # 构建配置表
    if [[ ${NeedBuildConfig} == 1 ]]; then
        if [ $# -eq 2 ]; then
            case $2 in 
                "dev" | "stable" | "master")
                    configType=$2;
                ;;
                *)
                    configType="dev"
                ;;
            esac
            bash AllMake.sh buildconfig $configType
        else 
            bash AllMake.sh buildconfig 
        fi 
        if [ $? -ne 0 ]; then 
            NotifyDingDingTalk "failed" "build config Failed"
            exit 1
        fi

        # 构建成功，将版本号写入文件
        cd ${ProjectDir}/protocol/configs/resourse/
        echo `git log -n 1 --format=%h -- 配置表stable` > ${ConfigVersionFile} 
    fi
}

function Proto()
{
    cd $ProjectDir
    bash AllMake.sh clearpkg
    bash AllMake.sh proto
    if [ $? -ne 0 ]; then
        NotifyDingDingTalk "failed" "AllMake.sh proto Failed"
    fi
}

function BackupTarball()
{
    if [ ! -d $SaveDir ];then
        mkdir $SaveDir
    fi
    mv $1 $SaveDir
}

function IsNeedBuildConfig() 
{
    ConfigType=$1
    case $ConfigType in
        "dev")
            dirName="配置表dev";
        ;;
        "stable")
            dirName="配置表stable";
        ;;
        "master")
            dirName="配置表master";
        ;;
        *)
            dirName="配置表dev"
        ;;
    esac

    # 检查配置表是否更新
    cur=`pwd`
    cd ${ProjectDir}
    echo "ConfigVersionFile: ${ConfigVersionFile} (删除这个文件可强制全量打配置表)"
    if [[ ! -f ${ConfigVersionFile} ]]; then
        touch ${ConfigVersionFile}
    fi

    LastConfigVersion=`cat ${ConfigVersionFile}`
    cd ${ProjectDir}/protocol/configs/resourse/
    NowConfigVersion=`git log -n 1 --format=%h -- ${dirName}`
    echo "last config version ${LastConfigVersion}, now config version: ${NowConfigVersion}"
    if [[ $LastConfigVersion == ${NowConfigVersion} ]]; then
        return 0
    fi
    return 1
}

function WriteMeta() 
{
    updateDir=${ProjectDir}/run/tools/update
    echo "Branch $1" > ${updateDir}/meta.txt
}

function NotifyDingDingTalk() 
{
    Result=$1
    Content=`echo $2 | sed  "s/ /_/g"`
    Hook=https://oapi.dingtalk.com/robot/send?access_token=<your-dingtalk-robot-token>
    curl -X POST $Hook \
        -H 'Content-Type: application/json' \
        -d  '{
     "msgtype": "markdown",
     "markdown": {
         "title": "服务器打包通知",
         "text": "### 服务器打包通知\n#### 结果 '$Result' \n#### '${Content}'"
     },
 }'
}
