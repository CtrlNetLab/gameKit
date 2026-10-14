chcp 65001
setlocal enabledelayedexpansion

set rootDir=%cd%

:: 脚手架说明: proto 源文件位于 protocol/ 目录内(内联维护, 不从外部 proto 仓库拉取)
:: 完整构建流程也可以直接执行 protocol/convert.sh (Git Bash)

cd ../pkg

::清理掉以下几个目录
powershell -Command Remove-Item -Recurse -Force -Path ./configs -ErrorAction SilentlyContinue
powershell -Command Remove-Item -Recurse -Force -Path ./cs -ErrorAction SilentlyContinue
powershell -Command Remove-Item -Recurse -Force -Path ./ss -ErrorAction SilentlyContinue
powershell -Command Remove-Item -Recurse -Force -Path ./database -ErrorAction SilentlyContinue
powershell -Command Remove-Item -Recurse -Force -Path ./game_third -ErrorAction SilentlyContinue
powershell -Command Remove-Item -Recurse -Force -Path ./oss -ErrorAction SilentlyContinue
powershell -Command Remove-Item -Recurse -Force -Path ./opt -ErrorAction SilentlyContinue

cd ../protocol

::构建Pb
protoc --proto_path=share --proto_path=. --go_out=paths=source_relative:../pkg/ ss/*.proto
protoc --proto_path=share --proto_path=. --go_out=paths=source_relative:../pkg/ game_third/*.proto
protoc --proto_path=share --proto_path=. --go_out=paths=source_relative:../pkg/ oss/*.proto
protoc --proto_path=share --proto_path=. --go_out=paths=source_relative:../pkg/ configs/*.proto

cd ../win
::配置文件的转换
powershell -ExecutionPolicy Bypass -File "gen-primary.ps1"
cd ../protocol
echo F |xcopy /E /I /Y .\configs\ResLoadMgr.go.tmpl ..\pkg\configs\ResLoadMgr.go

::构建Pb
cd share
protoc --go_out=paths=source_relative:../../pkg/ cs/*.proto
protoc --go_out=paths=source_relative:../../pkg/ database/*.proto
protoc --go_out=paths=source_relative:../../pkg/ oss/*.proto
cd ..
cd ../tools/db
go mod tidy
go build -o db.exe scripts/main.go
db.exe

cd ../register
go mod tidy
go build -o register.exe scripts/main.go
register.exe

cd %rootDir%
call build_opt_proto.bat
