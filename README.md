# gamekit 脚手架

以 `ac-server`（Go 多服务游戏服务器）为基准抽取的**框架同构、业务剥离**脚手架。
保留完整的构建系统、框架库、协议生成管线、多服务骨架与运维脚本；ac 的具体业务已剥离，
内置最小示例（登录 App / Player 模型 / 示例配置表），**clone 后即可编译**。

> 模块名前缀为 `gamekit`（原项目为 `acserver`），Go import 路径为 `gamekit/...`。

## 前置依赖

| 依赖 | 版本 | 说明 |
|---|---|---|
| Go | 1.21+ | go.mod 声明 `go 1.25.9`，`GOTOOLCHAIN=auto` 会自动下载对应工具链 |
| protoc | 任意 3.x | 协议生成（`protocol/convert.sh` 或 `win/build_proto.bat`） |
| protoc-gen-go | **v1.34.2** | `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2` |
| Python3 | 3.8+ | `run/config.py` 渲染配置，依赖 `jinja2`、`pyyaml`、`requests` |
| make | GNU make | Linux 构建入口（`AllMake.sh` 内部调用） |

> 脚手架已提交协议生成的 `.pb.go`，**未安装 protoc 也可直接编译运行**；
> 修改 proto 后才需要 protoc 工具链。

## 目录结构

```
gameServer/
├── AllMake.sh            # 一键构建入口(proto/build/buildconfig/clearpkg/...)
├── Makefile              # 构建系统总入口(make help 查看全部目标)
├── pkg/                  # module: gamekit/pkg —— 协议生成物(cs/ss/database/configs/oss/game_third/opt)
├── protocol/             # proto 源文件(内联维护, 不从外部仓库拉取)
│   ├── share/cs/         #   客户端协议 CS
│   ├── share/database/   #   数据库实体
│   ├── ss/               #   服务间协议 SS
│   ├── configs/          #   配置表 proto + primarykey.txt + 生成脚本
│   ├── game_third/       #   第三方协议
│   └── opt/scommon/      #   运营平台协议
├── src/
│   ├── base/             # module: 框架库(control/relay/dao/logger/osslog/utils/snowflake/...)
│   ├── gamesvr/          # module: 主游戏服(示例: 登录 App + Player 模型)
│   ├── relaysvr/         # module: 消息中继服务
│   ├── matchsvr/         # module: 匹配服务
│   ├── roomsvr/          # module: 房间服务
│   ├── miscsvr/          # module: 杂项服务
│   ├── globalsvr/        # module: 全局服务
│   ├── control/          # module: 控制服务(control 命令行)
│   └── benchmark/        # module: 压测机器人
├── run/                  # 运维层: 配置渲染/启停脚本/部署
│   ├── admin.sh          #   统一启停入口(config/start/stop/restart/control/clearlog)
│   ├── config.py         #   配置渲染(jinja2: *.j2 -> 服务配置)
│   ├── config/local.yaml #   本地单机开发配置示例
│   ├── <service>/{bin,conf}/  # 各服务启停脚本与配置模板
│   ├── binconf/          #   配置表二进制数据(configs/*.bytes)与公共配置
│   ├── redislua/         #   Redis Lua 脚本
│   └── tools/            #   运维脚本(db 建表/巡检/清 key 等)
├── tools/                # 工具链: buildInfo/db/register/excelToProtobuf/client/misc
├── win/                  # Windows 开发辅助(bat 脚本 + dlv 调试配置)
├── pack/                 # 打包脚本(脱敏占位, 按需修改)
├── doc/                  # 智能体规范: AGENTS.md / agent-coding-rules.md / feature-workflow.md
└── skills/               # 场景技能(协议/配置表/Redis/主线程模型/... 见 skills/README.md)
```

## 快速开始

```bash
# 1. 编译全部服务(Linux; 输出到 run/<service>/bin/)
./AllMake.sh build

# 2. 渲染服务配置(本地单机开发用内置 local.yaml)
cd run && bash admin.sh config local.yaml && cd ..

# 3. 启动全部服务(默认 develop = 单机全服务)
cd run && bash admin.sh start && cd ..

# 4. 停止 / 重启 / 进 control 命令行
cd run && bash admin.sh stop
cd run && bash admin.sh restart gamesvr     # 单个服务
cd run && bash admin.sh control gamesvr     # 进入 gamesvr control 命令行
```

> 服务依赖本机 MySQL 与 Redis（`run/config/local.yaml` 中默认指向 127.0.0.1）。
> 首次初始化数据库：`cd run && bash admin.sh init`。

## 服务清单

| 服务 | 目录 | 职责 |
|---|---|---|
| gamesvr | `src/gamesvr` | 主游戏服：客户端协议入口、Player 模型、业务 App |
| relaysvr | `src/relaysvr` | 消息中继：服务注册、SS 转发、在线查询 |
| matchsvr | `src/matchsvr` | 匹配协调 |
| roomsvr | `src/roomsvr` | 房间/战斗逻辑 |
| miscsvr | `src/miscsvr` | 杂项服务 |
| globalsvr | `src/globalsvr` | 全局服务：跨区数据、好友等 |
| control | `src/control` | 控制服务：GM/control 命令行接入 |
| benchmark | `src/benchmark` | 压测机器人：模拟登录/心跳/收发协议 |

## 构建系统

```bash
./AllMake.sh proto        # 全量重新生成 pb(等价 protocol/convert.sh)
./AllMake.sh protox       # 增量生成
./AllMake.sh build        # 编译全部服务(自动先执行 make buildInfo)
./AllMake.sh build gamesvr# 只编译单个服务
./AllMake.sh buildconfig  # 读取 Excel 生成配置表 proto/json(需要策划表)
./AllMake.sh modtidy      # 批量 go mod tidy
./AllMake.sh clearpkg     # 清理生成的 .pb.go
```

要点：
- **每个服务独立 go.mod**，`replace` 指向 `gamekit/pkg` 与 `gamekit/src/base` 相对路径；
  编译必须 `cd src/<service>` 后执行，根目录没有主 module。
- **wsgnet(gnet) 仅支持 Linux**：开发机编译务必 `GOOS=linux GOARCH=amd64`（`AllMake.sh`/`Makefile` 已内置）。
- 生成物策略：脚手架**提交** `.pb.go` 保证 clone 后可编译；修改 proto 后重新生成即可
  （如遵循原项目"生成物不入库"惯例，删除 `.gitignore` 底部"脚手架例外"中的 pb 段落，
  并在 CI 中执行 `protocol/convert.sh`）。

## 常用开发任务（按场景查技能）

| 任务 | 参考 |
|---|---|
| 新增/修改协议、命令字注册 | `skills/protocol-and-register` + `doc/feature-workflow.md` |
| 新增配置表（Excel → pb/json → ResMgr） | `skills/config-table` |
| 新增事件订阅与派发 | `skills/add-new-event` |
| Redis/Lua 使用规范 | `skills/redis-usage` |
| gamesvr 主线程模型与 WorkerPool | `skills/gamesvr-mainthread` |
| 本地环境配置与启停 | `skills/run-and-env` |
| 服务架构总览 | `skills/server-architecture` |
| 离线数据补发模式 | `skills/offline-data-mgr` |
| 日志分析 | `skills/log-analysis` |

## 与原项目（ac-server）的差异 / 业务补齐指引

脚手架保留了框架的全部机制，业务实现按最小示例内置。开发新项目时通常需要：

1. **账号体系**：`src/gamesvr/scripts/model/PlayerMgr.go` 的 `GenAccId` 目前用 snowflake 生成，
   接平台账号时替换为平台 AccId。
2. **配置表**：`protocol/configs/` 仅有 2 个示例表；按 `skills/config-table` 流程新增，
   并在 `primarykey.txt` 登记。
3. **数据库表**：`protocol/share/database/` 仅含 `AccData`/`Account`/`PlayerData`；
   新表在 proto 中新增后，在 `tools/db/scripts/main.go` 追加代码块并重跑
   `tools/db`（生成 db 访问层）与 `tools/register`。
4. **业务消息与事件**：`ss_type.proto` 的 `EnmItemReason`、`ss_eventtype.proto` 的业务事件、
   `ss_common.proto` 的业务消息均未内置，按 `doc/feature-workflow.md` 的流程追加。
5. **任务/活动/离线数据等业务模块**：未内置，`skills/offline-data-mgr`、
   `doc/feature-workflow.md`（补丁/预处理工作流）提供了可复用模式。
6. **部署**：`run/ansible/` 为占位目录，按 `run/ansible/README.md` 初始化；
   `pack/` 下脚本已脱敏（Token/IP/桶名等为 `<your-xxx>` 占位），按需填写。

## Windows 开发说明

- 服务只能在 Linux 运行；Windows 开发机用 `GOOS=linux GOARCH=amd64` 交叉编译验证。
- `win/` 目录提供：`build_proto.bat`（协议生成）、`build_excel.bat`（配置表）、
  `build_opt_proto.bat`（运营协议）、`gen-primary.ps1`、`start_dlv_*.sh`（dlv 远程调试）。
- `go vet ./...` 可用于 tools 目录的编译验证（部分工具 `go build` 会与目录名冲突）。

## 文档入口

- `doc/AGENTS.md`：智能体协作流程与技能路由
- `doc/agent-coding-rules.md`：编码规范（MUST/SHOULD/MUST NOT）
- `doc/feature-workflow.md`：新功能新增全流程（proto → model → apps → ctrl → 编译 → 重启）
- `skills/README.md`：技能索引
