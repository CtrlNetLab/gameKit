#!/bin/bash

# Jenkins 触发构建占位脚本(脚手架不内置任何凭证)
# 使用方式: 在下方填入你自己项目的 Jenkins 地址与 API Token, 或直接删除本文件
#
# projectUrlMap["dev"]="https://<your-jenkins>/job/<your-project>/build"
#
# curl -s -X POST -u "<user>:<api-token>" \
#     -H "Jenkins-Crumb: <crumb>" \
#     "${projectUrlMap[$1]}"

echo "请先配置 Jenkins 地址与凭证(见本文件注释)"
