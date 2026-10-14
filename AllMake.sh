#!/bin/bash

dir=`cd $(dirname $0);pwd`
cd $dir

## @NOTE 新增服务时需要在这里添加 
# control 也加入，避免不同linux glibc不兼容
gservice=("gamesvr" "relaysvr" "globalsvr" "matchsvr" "roomsvr" "miscsvr")

who=`whoami`
buildOutfile=/tmp/${who}.cmake;

function Usage()
{
    echo "Usage: $0 cmd"
    echo "  help - 显示帮助列表"
    echo "  pull - 下拉代码"
    echo "  proto - .proto 转 .go 全量"
    echo "  protox - .proto 转 .go 增量"
    echo "  build - build [service], 没有 service 参数时编译 src 目录下所有服务, 有 service 参数则只编译对应的 service"
    echo "  onecmd - 一键更新启动，将重建数据库"
    echo "  clearpkg - 删除生成的 .pb.go 文件"
    echo "  buildconfig - 打表工具读取策划仓库Excel生成配置json与proto"
}

if [ $# -lt 1 ];
then
    Usage;
    exit 1;
fi

function MakeProto()
{
    cd ${dir}/protocol;
    ./convert.sh;
    cd ${dir};
}

function IfErrorExit()
{
    if [[ $1 != 0 ]];then
        echo "build $2 failed";
        exit 1
    fi  
}

function ModTidy()
{
    for s in ${gservice[*]}
    do
        cd ${dir}
        cd src/${s};
        go mod tidy
    done
}

function BuildExcel()
{
    cd ${dir}/tools/excelToProtobuf
    ./build.sh

    cd ${dir}/protocol/configs/
    if [ $# -ne 0 ]; then 
        ./build_excel.sh $1
    else 
        ./build_excel.sh
    fi
}


function MakeAllService()
{
    for s in ${gservice[*]}
    do
        MakeService ${s};
        IfErrorExit $? ${s}; 
        echo "build ${s} success"
    done
    echo "ALL_SERVER_BUILD_SUCCESS"
}

function ClearPKG()
{
    ls -d  pkg/*/ | while read subDir; do
        if [[ "$subDir" != *"pkg/mod/"* && "$subDir" != *"pkg/sumdb/"* ]]; then
            find ${subDir} -regex ".*\.pb.go" -exec rm {} \;
        fi
    done
    echo "clear pkg success ..."
}

function GitPull()
{
    git pull
}
        git pull origin ${curbranch}
    done
}

function OneCmd()
{
    cd ${dir}/run
    ./admin.sh stop
    ./admin.sh clearlog
    echo "yes" | ./admin.sh reinit
    cd ${dir}
    ./AllMake.sh buildconfig
    ./AllMake.sh proto
    ./AllMake.sh build
    cd ${dir}/run
    ./admin.sh config
    ./admin.sh start
    cd ${dir}
}

function MakeService()
{
    if [ $# -ne 1 ]
    then
        echo "Usage: ${FUNCNAME} service";
        exit 1;
    fi
    ## 生成编译信息
    make -C ${dir} buildInfo -s > /dev/null
    serv=$1;
    cd ${dir}/src/${serv};
    ./build.sh 2> ${buildOutfile};
	let retCode=$?

    sed -i "s/^\(scripts\/.*\)/src\/${serv}\/\1/g; s/^\.\.\/\(base\/.*\)/src\/\1/g; s/^\.\.\/\.\.\/pkg/pkg/g;" ${buildOutfile}
    cat ${buildOutfile}

    if [ $retCode -ne 0 ]; then
        exit $retCode
    fi

    return $retCode
}

case $1 in
    help)
        Usage;
        ;;
    proto)
        MakeProto;
		;;
	modtidy)
        ModTidy;
        ;;
	build)
        > ${buildOutfile}
		if [ $# == 1 ]
		then
			MakeAllService;
		elif [ $# == 2 ];
		then
			MakeService $2;
		else
			Usage;
		fi
		;;
    pull)
        GitPull
        cd ${dir}
        ;;
    onecmd)
        OneCmd
        ;;
    clearpkg)
        ClearPKG
        ;;
    buildconfig)
        BuildExcel $2
        ;;
    *)
        Usage;
esac

