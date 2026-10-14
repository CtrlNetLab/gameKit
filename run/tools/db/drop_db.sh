#!/bin/bash
# shellcheck shell=bash source=env.sh

. env.sh

DropDB() {
    local MYSQL="$1"
    local DBName="$2"

    if ${MYSQL} -e "show databases" | grep -qw "\<${DBName}\>"; then
        ## 删除数据库
        (
            set -x
            # shellcheck disable=SC2086
            ${MYSQL} -e "drop database ${DBName}"
        )
        return 1
    fi
    return 0
}

echo "start drop database ..."

AccMYSQL="mysql -h${AccMySQLHost} -P${AccMySQLPort} -u${AccMySQLUser} -p${AccMySQLPasswd}"
PlayerMYSQL="mysql -h${PlayerMySQLHost} -P${PlayerMySQLPort} -u${PlayerMySQLUser} -p${PlayerMySQLPasswd}"
OrderMYSQL="mysql -h${OrderMySQLHost} -P${OrderMySQLPort} -u${OrderMySQLUser} -p${OrderMySQLPasswd}"
MailMYSQL="mysql -h${MailMySQLHost} -P${MailMySQLPort} -u${MailMySQLUser} -p${MailMySQLPasswd}"
FriendMYSQL="mysql -h${FriendMySQLHost} -P${FriendMySQLPort} -u${FriendMySQLUser} -p${FriendMySQLPasswd}"
GlobalMYSQL="mysql -h${GlobalMySQLHost} -P${GlobalMySQLPort} -u${GlobalMySQLUser} -p${GlobalMySQLPasswd}"

DropDB "${AccMYSQL}" "${AccDB}"
DropDB "${PlayerMYSQL}" "${PlayerDB}"
DropDB "${OrderMYSQL}" "${OrderDB}"
DropDB "${MailMYSQL}" "${MailDB}"
DropDB "${FriendMYSQL}" "${FriendDB}"
DropDB "${GlobalMYSQL}" "${GlobalDB}"

echo "finish drop database ..."
