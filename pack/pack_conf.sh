#!/bin/bash

Env=$1
Tag=$2

. common_args.sh
. common_function.sh

MakeRecord

SvnUp

# 清理临时存放目录
if [ ! -d $TmpPackDir ]; then 
    mkdir $TmpPackDir
else
    rm -rf $TmpPackDir
    mkdir $TmpPackDir
fi

for Process in ${GameProcess[@]} 
do 
    PackSvrConfig $Process
done

PackDataConfig
PackConfDir
PackTools
PackScripts
PackRedisLua

if [ ! -d $SaveDir ]; then
    mkdir $SaveDir
fi

cd $PackageDir/$GameName
tar cvfz $SaveDir/$TarBall `ls`
if [[ $? -ne 0 ]]; then
    NotifyFeiShu "failed" "create tar ball failed"
    exit 1
fi
rm -rf $GameName

if [[ $Env == "internal" ]]; then 
    NewPackageURL="http://<your-internal-host>:8088/$TarBall"
    NotifyFeiShu "success" $NewPackageURL
    echo "Success : $NewPackageURL"
    redis-cli -h <your-redis-host> -p 6379 -a <your-redis-password> HSET NewPackageURL $BranchName $NewPackageURL
elif [[ $Env == "external" ]]; then
    CosURLPre="${GameName}/${ProjectDirName}/${BranchName}/"
    echo "coscmd upload $SaveDir/$TarBall $CosURLPre"
    uploadResult=`coscmd upload $SaveDir/$TarBall $CosURLPre`
    if [[ $? -ne 0 ]]; then
        NotifyFeiShu "failed" "上传到COS失败"
        exit 1 
    fi
    echo $uploadResult

    cosUrl="cos://gamekit-1316747386/"$CosURLPre${TarBall}
    echo ${cosUrl}
    NotifyFeiShu "success" $cosUrl
    echo "Success : $cosUrl"
    redis-cli -h <your-redis-host> -p <your-redis-port> -a <your-redis-password> HSET NewPackageURL $BranchName ${CosURLPre}${TarBall} 
fi
#if [ ! -f $Docker ]; then 
#    echo "Please Install docker First"
#    exit 1
#fi 
#
#cat > Dockerfile << EOF
#EOF
#
#$Docker build -t kleins:latest . | tee -a $RecordFile
#
#BackupTarball $PackageDir/$TarBall
#ImageID=`grep "Successfully built" $RecordFile | awk '{print $NF}'`
#echo $ImageID
#
#echo "Docker Saving Image..."
#$Docker save -o kleins.tar kleins:latest
#echo "Docker Image $ImageID Saved"
#
#echo "Compressing Image $ImageID..."
#gzip -f kleins.tar
#echo "Compressing Done"
