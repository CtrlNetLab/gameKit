#!/bin/bash

. ~/.bashrc
if [ -f env.sh ]; then
    . env.sh
fi 

dir=`cd $(dirname $0);pwd`
cd $dir

## 开发环境
ENV="develop"

## 默认 develop(本地单机开发); 多机环境通过 run/env.sh(已 gitignore)或导出 machine 变量覆盖
machine=${machine:-"develop"}

## @NOTE 新增服务时需要在这里添加
# 开发默认启动全部服务
developService=("relaysvr" "globalsvr" "matchsvr" "roomsvr" "miscsvr" "gamesvr" "benchmark")

# 内网测试环境使用两种环境类型
areaService=("relaysvr" "gamesvr" "roomsvr" "miscsvr" )
worldService=("matchsvr" "globalsvr")
gameService=("gamesvr")

case $machine in
    "area")
        gservice=${areaService[@]};
    ;;
    "develop")
        gservice=${developService[@]};
    ;;
    "world")
        gservice=${worldService[@]};
    ;; 
    "game")
        gservice=${gameService[@]};
    ;;
    *)
        if [ $1 != "config" ]; then
            echo "please set machine [all, game, global...]"
            exit 1
        fi
    ;;
esac

function Usage()
{
    echo "Usage: $0 cmd"
    echo "  help - 显示帮助列表"
    echo "  init - 第一次运行时创建数据库, 设置 redis 值"
    echo "  reinit - 删除原有的数据库、redis，重新创建数据库, 设置 redis 值"
    echo "  config - 配置相关值，一般只有配置修改时才需要执行"
    echo "  config [config_file] - 使用指定的配置文件生成配置，例如 bash admin.sh config dev.yaml"
    echo "  start - start [service], 启动服务，有 service 参数表示启动具体某个服务，否则启动所有服务"
    echo "  control - control service, 进入服务 control 命令行"
    echo "  stop - stop [service], 关闭服务，有 service 参数表示关闭具体某个服务，否则关闭所有服务"
    echo "  restart - restart [service], 重启服务，有 service 参数表示重启具体某个服务，否则重启所有服务"
    echo "  clearlog - clearlog [service], 删除日志，有 service 参数表示删除所有服务的日志，否则删除所有服的日志"
}

function IfErrorExit()
{
    if [[ $1 != 0 ]];then
        echo "start $2 failed";
        exit 1
    fi  
}

function Config() 
{
    if [ $# -ge 2 ]; then
        # shift to get the config filename after "config"
        shift
        if ( python3 config.py $@); then
            echo "successful"
        else 
            echo "failed"
            exit 1
        fi
    elif [ $# -eq 1 ]; then
        if ( python3 config.py); then
            echo "successful"
        else 
            echo "failed"
            exit 1
        fi 
    else
        if ( python3 config.py $@); then
            echo "successful"
        else 
            echo "failed"
            exit 1
        fi
    fi
}

function StartService()
{
    if [ $# -ne 1 ]
    then
        echo "Usage: ${FUNCNAME} service";
        exit 1;
    fi

    cd ${dir};
    service=$1;
    cd ${service}/bin;
    ./start.sh
	let retCode=$?

    return $retCode
}

function StartAllService()
{
    for s in ${gservice[*]}
    do
        StartService $s
        IfErrorExit $? $s
    done
}

function StopService()
{
    if [ $# -ne 1 ]
    then
        echo "Usage: ${FUNCNAME} service";
        exit 1;
    fi

    cd ${dir};

    service=$1;
    cd ${service}/bin;
    echo "stop ${service} ..."
    ./stop.sh
	let retCode=$?

    return $retCode
}

function StopAllService()
{
    for s in ${gservice[*]}
    do
        StopService $s
    done
}

function ClearAllServiceLog()
{
    for s in ${gservice[*]}
    do
        if [ ${#s} -gt 3 ]
        then
            rm -rf $s/log/*
        fi
    done
    echo "clear all service log success"
}

function ClearServiceLog()
{
    if [ $# -ne 1 ]
    then
        echo "Usage: ${FUNCNAME} service";
        exit 1;
    fi
    s=$1
    if [ ${#s} -gt 3 ]
    then
        rm -rf $s/log/*
    fi
    echo "clear $s service log success"
}

function InitEnv()
{
    cd ${dir};
    cd tools/db;
    # echo "start init mysql ..."
    # 初始化数据库 
    ./create_db.sh;
    # @TODO ryder: 初始化 redis
    cd ${dir};
}

function ReInitEnv()
{
    cd ${dir};
    cd tools/db;
    # 删除数据库
    ./drop_db.sh
    # 删除 redis
    cd ${dir};
    InitEnv;
}

case $1 in
    help)
        Usage;
    ;;
    init)
        if [ $ENV != "develop" ]
        then
            echo "just for develop env"
            exit 1
        fi
        InitEnv;
    ;;
    reinit)
        if [ $ENV != "develop" ]
        then
            echo "just for develop env"
            exit 1
        fi
        # echo -n "are you sure drop redis AccId_${who} and mysql accdb_${who}/roledb_${who} (yes/no): "
        echo -n "are you sure drop redis and mysql (yes/no): "
        read ask;
        if [ "x${ask}" == "xyes" ]
        then
            ReInitEnv;
        fi
    ;;
    config)
	    Config $@;
    ;;
    start)
		if [ $# == 1 ]
		then
            StartAllService;
		elif [ $# == 2 ];
		then
		    StartService $2;
		else
			Usage;
		fi
    ;;
    stop)
		if [ $# == 1 ]
		then
            StopAllService;
		elif [ $# == 2 ];
		then
		    StopService $2;
		else
			Usage;
		fi
    ;;
    restart)
		if [ $# == 1 ]
		then
            StopAllService;
            sleep 2
            StartAllService;
		elif [ $# == 2 ];
		then
		    StopService $2;
            sleep 2
		    StartService $2;
		else
			Usage;
		fi
    ;;
    control)
        if [ $# -ne 2 ]
        then
            Usage;
        else
            ./control.sh $2
        fi
    ;;
    clearlog)
		if [ $# == 1 ]
		then
            ClearAllServiceLog;
		elif [ $# == 2 ];
		then
		    ClearServiceLog $2;
		else
			Usage;
		fi
    ;;
    createdb)
		cd tools/db
		./create_db.sh
    ;;
    none)

    ;;
    *) 
        Usage;
    ;;
esac

