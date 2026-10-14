#!/bin/bash

db=$1
prefix=$2

redis="redis-cli"
${redis} -n ${db} --raw keys "${prefix}*" | awk '{ printf("del %s\n", $1); }' | ${redis} -n ${db} -x >/dev/null 2>&1

echo "success del ${prefix}*"