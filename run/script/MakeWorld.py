#!/usr/bin/python
# coding=utf-8
import re
import sys

#ip 点分式字符串
split_world = "[A-Za-z-0-9$\{\}_]+"
#ip 点分式正则表达式
ip_regex = "%s\.%s\.%s\.%s"%(split_world,split_world,split_world,split_world)
bus_id_regex_re = re.compile(ip_regex)
#捕获 ip 点分式字符
split_bus_regex = "(%s)\.(%s)\.(%s)\.(%s)"%(split_world,split_world,split_world,split_world)
split_busid_re = re.compile(split_bus_regex)
serverid_str = "ServerID"
world_str = "WorldID"
#battle_world_key = "${BattleZoneWorldID}"
#battle_world_str = "BattleZoneWorldID"

def remake_bus(line,config_hash,variable_key,key_str):
    str_world_id = config_hash[key_str] 
    #获得 worldid 的值
    int_world_id = int(str_world_id)
    for match in bus_id_regex_re.findall(line):
        #通道中是否有 WorldID 变量
        index = match.find(variable_key)  
        if index >= 0:
            #匹配点分式 4 部份
            split_array = split_busid_re.findall(match)
            #点分式第二部份
            tow = split_array[0][1].strip("${}")
            if config_hash.has_key(tow):
                tow = config_hash[tow]
            #点分式第三部份
            three = split_array[0][2].strip("${}")
            if config_hash.has_key(three):
                three = config_hash[three]
            #点分式第四部份
            four = split_array[0][3].strip("${}")
            if config_hash.has_key(four):
                four = config_hash[four]
            four = int(four)
            #重新组装通道
            #保留前 8 位
            one = str(int_world_id & 0xff)
            #保留高 3 位，再和第四个点分式 | 运算
            four = str((((int_world_id >> 8)&0x7) << 5) | four)
            result = "%s.%s.%s.%s"%(one,tow,three,four)
            #替换通道
            line = line.replace(match,result)
    return line

'''
    @brief 替换字符串中 worldid 大过 255 的通道,第一个 ip 分点式必须为 worldid,worldid 目前支持最大 2047(11 位),使用第四个点分式高 3 位 
    @param line 表示文件中的一行的字符串
    @param config_hash Config.py 中的 config 字典
'''
def replace_worldid(line,config_hash):
    line = remake_bus(line,config_hash,world_key,world_str)
    line = remake_bus(line,config_hash,battle_world_key,battle_world_str)
    return line

#测试生成通道
def test_world():
    line = "<Address>${WorldID}.${ZoneID}.${XLGatewayGSFuncID_2}.1</Address>"
    config_array = {"WorldID":"256","ZoneID":"1","XLGatewayGSFuncID_2":"21"}
    print(replace_worldid(line,config_array))

#生成 online_list 通道,可使用命令调用(python MakeWorld.py 2047.1.4.1),表示对 2047.1.4.1 的通道进行转换
def get_online_list(bus_str):
    split_array    = split_busid_re.findall(bus_str)
    world_id       = split_array[0][0]
    ip_point_tow   = split_array[0][1]
    ip_point_three = split_array[0][2]
    ip_point_four  = split_array[0][3]
    config_array   = {"WorldID":world_id,"BattleZoneWorldID":world_id}
    line = "${BattleZoneWorldID}.%s.%s.%s"%(ip_point_tow,ip_point_three,ip_point_four)
    print(replace_worldid(line,config_array))

if __name__ == "__main__":
    get_online_list(sys.argv[1])

