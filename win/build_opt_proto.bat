@echo off
cd ../protocol

@rem 运营平台 pb仓库(脚手架仅保留 scommon, 项目自有 proto 仓库请自行替换)
set pbRepoCommon=https://git.example.com/your-org/operation-platform/proto-server-common.git
@rem 项目
set projScommon=scommon
@rem 当前目录
set rootDir=%cd%
@rem 运营平台pb文件根目录
set optDir=opt
@rem 输出目标目录
set pbOutputDir="%rootDir%/../pkg/%optDir%"
@rem 要替换的go选项占位符
set goOptTag=__GO_OPTS__
@rem go选项值
set goOptValCommon=gamekit/pkg/opt/scommon;optscommon


@rem 如果相关目录不存在,则创建
if not exist "%rootDir%/%optDir%" (
    mkdir "%rootDir%/%optDir%"
)
if not exist %pbOutputDir% (
    mkdir %pbOutputDir%
)


@rem 同步仓库
call :syncPbRepo %projScommon% %pbRepoCommon%

echo "==> building proto file..."

cd "%rootDir%"

@rem 为了让pb文件的go选项动态替换，拷贝一个.opt目录出来，然后动态把 __GO_OPTS__ 替换成指定的值
if exist .%optDir% (
  rd  /s /q .%optDir% > nul
)
xcopy /E /I /Y %optDir% .%optDir% > nul
@rem 替换go选项
pwsh -Command "Get-ChildItem -Recurse -Filter *.proto -File -Path '.%optDir%/%projScommon%' | ForEach-Object { (Get-Content $_.FullName) -replace '%goOptTag%', '%goOptValCommon%' | Out-File -Encoding utf8NoBOM $_.FullName }"
@rem 构建pb文件并输出

protoc -I "./.%optDir%" --go_out=paths=source_relative:%pbOutputDir% ./.%optDir%/%projScommon%/*.proto


@rem 最后删除临时的pb目录
rd  /s /q .%optDir% > nul

echo "==> build proto success"


goto :EOF
@rem 同步pb函数
:syncPbRepo
cd %rootDir%/%optDir%
set proj=%1
set repo=%2
echo "==> start sync repo:%proj%"
if exist %proj% (
  @rem 如果存在该目录，则进入里面pull
  cd %proj%
  git pull
) else (
  git clone %repo% %proj%
)
echo "==> finish sync repo:%proj%"
goto :EOF
