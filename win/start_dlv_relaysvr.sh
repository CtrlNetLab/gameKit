#!/bin/bash

# 设置进程名和端口号
PROCESS_NAME="relaysvr"
PORT=8886

# 检查是否已安装 dlv
command -v dlv >/dev/null 2>&1
DLV_INSTALLED=$?

if [ $DLV_INSTALLED -ne 0 ]; then
  echo "dlv not found, installing..."
  go install github.com/go-delve/delve/cmd/dlv@latest
fi

# 检查进程是否在运行中
PIDS=($(pgrep -f "$PROCESS_NAME"))

VALID_PIDS=()

for PID in "${PIDS[@]}"; do
    TERMINAL=$(ps -p $PID -o tty --no-headers)
    if [ "$TERMINAL" != "?" ]; then
        VALID_PIDS+=($PID)
    fi
done

if [ ${#VALID_PIDS[@]} -eq 0 ]; then
    echo "Error: No valid $PROCESS_NAME process is running."
    exit 1
fi

# 关闭已经运行的 dlv 进程
DLV_PIDS=($(pgrep -x dlv))

for DLV_PID in "${DLV_PIDS[@]}"; do
    sudo kill $DLV_PID
done

# 延迟3秒
echo "Wait a moment..."
sleep 3

# 只选择第一个有效的进程ID
PID=${VALID_PIDS[0]}

# 执行命令
COMMAND="dlv attach $PID --listen=:$PORT --headless --api-version=2 --log"
echo "Executing command: $COMMAND"
$COMMAND