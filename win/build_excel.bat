chcp 65001

:: 构建配置表: excel -> proto -> pb -> ResMgr
:: 依赖 tools/excelToProtobuf 与 protocol/configs/convert.sh (Git Bash 环境)

cd ../tools/excelToProtobuf
go mod tidy
go build -o excelToProtobuf.exe main.go
excelToProtobuf.exe

cd ../../protocol/configs
bash convert.sh

echo "==> build excel config success"
