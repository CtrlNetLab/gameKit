---
name: protocol-and-register
description: 修改,新增协议 proto、生成 pb.go 输出，并维护命令字枚举与消息注册规则。
---

# 协议与注册

当你需要改 protobuf、命令字枚举或重新生成 `pkg/*` 协议输出时使用此 skill。

## 规则边界

- 本 skill 说明协议改动时的入口与生成流程。
- 跨场景硬约束（`Req/Res` 命名一致性、生成物必须同步）以 `doc/agent-coding-rules.md` 为准。

## 关键位置

- `protocol/share/cs`: 客户端-服务端 proto 源文件（通过 `protocol/cs` 软链共享）。
- `protocol/ss`: 服务端-服务端 proto 源文件。
- `protocol/convert.sh`: protobuf 生成主流程，末尾会执行 register 工具。
- `tools/register`: 生成 `pkg/cs` 与 `pkg/ss` 的命令字->消息映射。
- 输出文件：
  - `pkg/cs/*.pb.go`, `pkg/ss/*.pb.go`
  - `pkg/cs/client_register.pb.go`
  - `pkg/ss/service_register.pb.go`

## 操作流程

1) 修改 `protocol/share/cs` 或 `protocol/ss` 下的 proto。
2) 重新生成：
   - Linux: `bash AllMake.sh proto`（建议）
   - Windows: `.\build_proto_dev.bat`（需要先进入当前仓库根目录下的win目录后执行）
   - 或直接执行 `bash protocol/convert.sh`（更底层，需自己保证环境完整）。

## 注册规则（关键）

- `tools/register` 读取 `pkg/cs/cs_cmdvalue.pb.go` 与 `pkg/ss/ss_cmdvalue.pb.go`。
- 枚举命名需与 message 名一致（大小写与后缀），具体硬约束遵循 `doc/agent-coding-rules.md`。
- `Req/Res` 的配对由后缀推导：
  - 枚举名以 `Req` 结尾时，期望存在对应 `Res`（最后一位替换 `Req` -> `Res`）。
- 枚举值为 `0` 的条目会被跳过。

## 快速检查

- 修改 proto 后确认 `pkg/cs/client_register.pb.go` 与 `pkg/ss/service_register.pb.go` 重新生成。
- 若 `Req/Res` 映射缺失，检查枚举命名与枚举值是否为非零。

## Windows 兼容提示

- 仓库根目录包含 `protocol/` 与 `pkg/`（如 `D:\game_go\gameServer`）。
- 若本机 Go 版本低于 1.20（或 Go 环境不完整）导致 `tools/db` 与 `tools/register` 无法编译，可直接运行已提交的可执行文件：
  - `tools/db/db.exe`
  - `tools/register/register.exe`
