#!/usr/bin/expect

set timeout 600
set fromfile  [lindex $argv 0]
set tofile    [lindex $argv 1]
set user      [lindex $argv 2]
set password  [lindex $argv 3]
set ip        [lindex $argv 4]

set pid [spawn scp ${fromfile} ${user}@${ip}:${tofile}]
expect {
 "password:" {
     send "${password}\r"
     exp_continue
 }
 "yes/no)?" {
     send "yes\r"
     exp_continue
 }
 "failed" {
	exit 1
 }
}
