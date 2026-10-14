#!/bin/bash

if [ $# != 2 ];
then
    echo "请输入 db prefix"
    exit
fi

db=$1
prefix=$2
echo -n "are you sure clear redis key ${db} ${prefix} (yes/no): "
read ask;
if [ "x${ask}" != "xyes" ];then
    exit
fi

redis="redis-cli"
${redis} -n ${db} --raw keys "${prefix}*" | awk '{ printf("del %s\n", $1); }' | ${redis} -n ${db} -x >/dev/null 2>&1

echo "success del ${prefix}*"