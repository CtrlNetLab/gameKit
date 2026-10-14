#!/bin/bash
# shellcheck shell=bash source=env.sh

. env.sh

CheckDB() {
    local MYSQL="$1"
    local DBName="$2"

    if ! ${MYSQL} -e "show databases" | grep -q "\<${DBName}\>"; then
        ## 创建数据库
        (
            set -x
            # shellcheck disable=SC2086
            ${MYSQL} -e "create database ${DBName}"
        )
        return 1
    fi
    return 0
}

CreateTable() {
    local MYSQL="$1"
    local DBName="$2"
    local TmplFile="$3"
    local Index="$4"

    ## 根据表模板生成分表 SQL
    TmpSQLFile="${TmplFile%.tmpl.sql}.sql"
    test -f "${TmpSQLFile}" && rm -f "${TmpSQLFile}"

    echo "create table ${TmpSQLFile%.sql}"
    for ((i = 0; i < Index; i++)); do
        sed 's#{{idx}}#'"${i}"'#' "${TmplFile}" >>"${TmpSQLFile}"
    done

    ## 创建分表
    (
        set -x
        # shellcheck disable=SC2086
        ${MYSQL} "${DBName}" <"${TmpSQLFile}"
        test -f "${TmpSQLFile}" && rm -f "${TmpSQLFile}"
    )
}

AccMYSQL="mysql -h${AccMySQLHost} -P${AccMySQLPort} -u${AccMySQLUser} -p${AccMySQLPasswd}"
PlayerMYSQL="mysql -h${PlayerMySQLHost} -P${PlayerMySQLPort} -u${PlayerMySQLUser} -p${PlayerMySQLPasswd}"
OrderMYSQL="mysql -h${OrderMySQLHost} -P${OrderMySQLPort} -u${OrderMySQLUser} -p${OrderMySQLPasswd}"
MailMYSQL="mysql -h${MailMySQLHost} -P${MailMySQLPort} -u${MailMySQLUser} -p${MailMySQLPasswd}"
FriendMYSQL="mysql -h${FriendMySQLHost} -P${FriendMySQLPort} -u${FriendMySQLUser} -p${FriendMySQLPasswd}"
GlobalMYSQL="mysql -h${GlobalMySQLHost} -P${GlobalMySQLPort} -u${GlobalMySQLUser} -p${GlobalMySQLPasswd}"

CheckDB "${AccMYSQL}" "${AccDB}"
CheckDB "${PlayerMYSQL}" "${PlayerDB}"
CheckDB "${OrderMYSQL}" "${OrderDB}"
CheckDB "${MailMYSQL}" "${MailDB}"
CheckDB "${FriendMYSQL}" "${FriendDB}"
CheckDB "${GlobalMYSQL}" "${GlobalDB}"

CreateTable "${AccMYSQL}" "${AccDB}" DBAccData.tmpl.sql 64
CreateTable "${PlayerMYSQL}" "${PlayerDB}" DBPlayerData.tmpl.sql 64
CreateTable "${OrderMYSQL}" "${OrderDB}" DBOrderData.tmpl.sql 64
CreateTable "${MailMYSQL}" "${MailDB}" DBMailData.tmpl.sql 64
CreateTable "${FriendMYSQL}" "${FriendDB}" DBFriendData.tmpl.sql 64
# CreateTable "${GlobalMYSQL}" "${GlobalDB}"

${GlobalMYSQL} "${GlobalDB}" <DBGlobalData.sql
