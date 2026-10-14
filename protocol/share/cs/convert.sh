#!/bin/bash

ProjectHome=`cd $(dirname $0)/../../../;pwd`
DIR=`cd $(dirname $0);pwd`
cd $DIR/..

outdir=$ProjectHome/pkg/
if [ ! -d ${outdir} ]
then
    mkdir -p ${outdir}
fi

# protoc${exe} --gofast_out=${outdir} cs/*.proto
protoc${exe} --go_out=paths=source_relative:${outdir} cs/*.proto

