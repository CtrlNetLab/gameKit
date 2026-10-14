#!/bin/bash

WHOAMI=`whoami`

outfile="/tmp/CheckAlive.${WHOAMI}.txt";

dir=`cd $(dirname $0);pwd`;
cd ${dir};

curl="/usr/bin/curl";
process="gamesvr";
host=`/usr/sbin/ifconfig | /usr/bin/sed -n 's#.*inet \([^ ]*\).*broadcast.*#\1#p'`

n=`/bin/ps -ef | /bin/grep "${WHOAMI}.*gamesvr/bin/gamesvr\>" | /bin/grep -v grep | wc -l`;
now=`date +"%Y-%m-%d/%H:%M:%S"`;
now0=`echo ${now}`
Webhook=https://open.feishu.cn/open-apis/bot/v2/hook/e5913274-693b-48d4-8265-30bc8b2824ee

if [ $n -eq 0 ];
then
    echo "[${now}] [$n] restart gamesvr" >> ${outfile};
    cd ${dir}/../../gamesvr/bin;
    ./start.sh
    cd ${dir};

    ${curl} -s ${Webhook} \
    -H 'Content-Type: application/json;charset=utf-8' \
    -d '
   {
       "msg_type": "post",
    "content": {
        "post": {
            "zh_cn": {
                "title": "服务器重启警告",
                "content": [
                    [
                        {
                                "tag": "text",
                                "text": "user: '${WHOAMI}'\n"
                            },
                            {
                                "tag": "text",
                                "text": "进程: '${process}'\n"
                            },
                            {
                                "tag": "text",
                                "text": "host: '${host}'\n"
                            },
                            {
                                "tag": "text",
                                "text": "时间: '${now0}'\n"
                            }
                    ]
                ]
            }
        }
    }
   }'\

fi

echo "[${now}] CheckAlive ..." >> ${outfile};
