#!/bin/bash

DIR=`cd $(dirname $0);pwd`
cd $DIR

source ~/.bashrc
stat database > /dev/null 2>&1
if [ $? -eq 1 ]
then
    echo "create database link";
    ln -s share/database database
fi

stat cs > /dev/null 2>&1
if [ $? -eq 1 ]
then
    echo "create cs link";
    ln -s share/cs cs
fi

stat oss > /dev/null 2>&1
if [ $? -eq 1 ]
then
    echo "create oss link";
    ln -s share/oss oss
fi


prefix="../pkg"
dirs=(cs configs ss oss game_third)
for d in ${dirs[*]}
do 
    if [ ! -d ${prefix}/${d} ]
    then 
        mkdir -p ${prefix}/${d}
    fi
done


which protoc-gen-go  2>&1 > /dev/null
if [ $? -ne 0 ]; then
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

## 放在环境检查那里 
GOPATH=`go${exe} env | grep PATH | sed -n 's#.*GOPATH="\([^"]*\)".*#\1#p'`
echo ${PATH} | grep -q "${GOPATH}"
if [ $? -eq 1 ];
then
    echo "export PATH=${PATH}:${GOPATH}/bin" >> ${HOME}/.bashrc
    source ${HOME}/.bashrc
fi

cd ${DIR}/share/cs
s=`date +%s`
./convert.sh
e=`date +%s`
cost=$((e-s))
echo "convert cs: ${cost}"

cd $DIR/ss
s=`date +%s`
./convert.sh
e=`date +%s`
cost=$((e-s))
echo "convert ss: ${cost}"

cd ${DIR}/game_third
s=`date +%s`
./convert.sh
e=`date +%s`
cost=$((e-s))
echo "convert game_third: ${cost}"

cd $DIR/oss
s=`date +%s`
./convert.sh
e=`date +%s`
cost=$((e-s))
echo "convert oss: ${cost}"

cd $DIR/configs
s=`date +%s`
./convert.sh
e=`date +%s`
cost=$((e-s))
echo "convert configs: ${cost}"

# opt pb build
cd $DIR
s=`date +%s`
chmod +x ./opt.sh
./opt.sh
e=`date +%s`
cost=$((e-s))
echo "build opt: ${cost}"

cd $DIR/share/database
s=`date +%s`
./convert.sh
e=`date +%s`
cost=$((e-s))
echo "convert database: ${cost}"

cd ${DIR}
## 注册命令字与 message 关系 
cd $DIR/../tools/register

s=`date +%s`
./build.sh

./register${exe}
e=`date +%s`
cost=$((e-s))
echo "convert register: ${cost}"

cd $DIR

