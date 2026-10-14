#coding=utf-8
import os
import sys
import socket
import getpass
from tempfile import template
import jinja2
import yaml
import requests

# 获取本机内网ip
def GetHostIP():
    s=socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.connect(('8.8.8.8', 80))
    ip=s.getsockname()[0]
    s.close()
    return ip

def LoadYamlFile(yaml_file):
    try:
        with open(yaml_file, 'r') as file:
            return yaml.safe_load(file) or {}
    except FileNotFoundError:
        return {}

# 获取本机外网IP
def GetInternetIP():
    res = requests.get("http://ifconfig.me")
    return res.text

# 优先匹配外网ip，没有匹配成功再匹配内网ip
def GetEvnByMachineIP(inIp, userName, machineData):
    for envTypes in machineData:
        envInsts = envTypes['EnvInst']
        if envInsts == None:
            continue
        for envInst in envTypes['EnvInst']:
            if envInst["Host"] == inIp and envInst["UserName"] == userName:
                return envTypes, envInst

    return None, None

def GetEnvTypeByEnvName(machine, envName):
    for envType, en in machine.items():
        if en == envName:
            return envType
    return "test"

def GetServerInfo(host, confDic):
    if confDic == None:
        return 0, "game", 0, 0
    for h in confDic:
        if h["Host"] == host:
            return h["ServerID"], h["MachineEnv"], h["ClientPort"], h["Area"]



internalIP = GetHostIP()     # 内网IP
UserName=getpass.getuser()

config = {}
# 检查是否通过参数指定了配置文件
if len(sys.argv) > 1:
    configFile = sys.argv[1]
    if not os.path.exists("config/" + configFile):
        print(f"错误: 配置文件 config/{configFile} 不存在")
        exit(1)
    # 使用指定的配置文件，默认使用命令行参数作为环境名
    envType = "custom"
    envName = os.path.splitext(configFile)[0]
    instId = 1
    print(f"使用指定配置文件: {configFile}")
    config["EnvName"] = envName+"_"+str(instId)
else:
    # 根据本机IP地址读取相关环境配置
    machine = LoadYamlFile("config/machine.yaml")
    envData, envInst = GetEvnByMachineIP(internalIP, UserName, machine)
    if envData is None or envInst is None:
        print("请先配置 config/machine.yaml 文件，或者通过参数指定配置文件名: bash admin.sh config <config_file.yaml>")
        exit(1)
    envType = envData["EnvType"]
    envName = envData["EnvName"]
    configFile = envInst["Config"]
    instId = envInst["InstId"]
    print("内网IP: " + internalIP + ", 环境名: " + envName+", 环境类型: " + envType + ", 配置文件: " + configFile)
    config["EnvName"] = envName+"_"+str(instId)

UniqueKey=UserName+"@"+internalIP
yamlConfig = LoadYamlFile("config/" + configFile)
yamlConfig["dist_user"] = {}
yamlConfig["dist_user"]["Roles"] = yamlConfig["Roles"]

for root, dirs, files in os.walk("."):
    for file in files:
        fileExts = os.path.splitext(file)
        if fileExts[1] == ".j2":
            env = jinja2.Environment(loader=jinja2.FileSystemLoader(root))  # 模板目录
            template = env.get_template(file)
            try:
                result = template.render(**yamlConfig)
            except jinja2.exceptions.UndefinedError as e :
                print("渲染配置文件失败: " + os.path.join(root, file) + ", 错误信息: " + str(e))
                continue
            # 将渲染后的配置写入配置文件
            with open(os.path.join(root, fileExts[0]), "w") as f:
                f.write(result)
                f.close()
            print("渲染配置文件成功 " + os.path.join(root, fileExts[0]))
