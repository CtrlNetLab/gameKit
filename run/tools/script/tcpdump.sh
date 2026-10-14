#!/bin/bash

portList=(9000 9004)
eth=ens192
date=`date +"%Y%m%d%H"`

for port in ${portList[*]}
do
    outfile=/tmp/tcpdump/${port}.${date}.txt

    ## kill 掉老的再重启 
    ps -ef | grep "tcpdump.*${port}" | grep -v grep  | awk ' { print $2; } ' | xargs -i kill -9 {}

    tcpdump -i ${eth} port ${port} -A -tttt >> ${outfile} 2>&1 &
done

