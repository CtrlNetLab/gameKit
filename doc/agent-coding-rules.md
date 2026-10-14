# gamekit 智能体编码规范（标准版）

> 目标：让进入本仓库的任意智能体都按同一套规则写代码、改协议、做评审。
> 适用范围：`gamekit` 全仓库，重点是 `src/*svr`、`protocol`、`pkg`、`run`。
> 协作流程（Plan 模式、上下文提醒、技能使用）以 `AGENTS.md` 为唯一准则。

## 1. MUST（必须遵守）

1. 分层落点必须正确  
- 客户端协议入口、命令注册放 `src/gamesvr/scripts/apps/*App.go`。  
- 业务状态与核心逻辑放 `src/gamesvr/scripts/model/*.go`。  
- 服务启动/初始化入口放 `src/*svr/scripts/main.go`。

2. gamesvr 主线程规则必须遵守  
- `MainServer.OnTick` 禁止写业务逻辑。  
- Redis/HTTP/DB 等耗时逻辑必须使用 `utils.WorkerPool.AddTask(...)`。  
- 任何玩家数据或公共内存修改，必须回主线程 `utils.WorkerPool.AddCallback(...)` 执行。

3. Redis 使用规范必须遵守  
- Redis 操作必须异步执行，禁止阻塞主线程。  
- 优先复用 `src/base/dao/RedisMgr.go` 已有封装。  
- 复杂原子操作用 Lua，脚本放 `run/redislua/`，并在对应 `run/*/conf/service.yaml.j2` 的 `Scripts` 注册。  
- `ExecScript(redisName, scriptName, primary, args...)` 参数语义不能改乱。

4. 协议/命令字一致性必须保证  
- `protocol/share/cs/cs_cmdvalue.proto` 的 `ECV_*Req/Res` 与 message 名（如 `CS_*Req/Res`）必须严格匹配。  
- 修改 proto 后必须重新生成 pb 和 register 映射。  
- 输出文件必须同步更新：`pkg/cs/*.pb.go`、`pkg/ss/*.pb.go`、`pkg/cs/client_register.pb.go`、`pkg/ss/service_register.pb.go`。

5. 错误处理和基础校验必须完整  
- 入口处理函数先做 `player == nil`、`!player.IsOnlineOK()`、`req == nil` 等防御性校验并尽早返回。  
- 错误日志要带关键上下文（如 `accId`、`taskId`、`cmdId`、player 标识）。

6. 注释规范必须满足可维护性  
- 新增或修改复杂业务逻辑时，必须写中文注释，说明业务意图、关键条件或边界原因。  
- 禁止“逐行翻译式”无意义注释，注释应聚焦难理解分支与规则来源。

7. 客户端错误返回前必须记录日志  
- 对客户端返回非成功错误码前，必须先打印错误日志。  
- 日志至少包含：用户信息（如 `accId`/player 标识）+ 业务信息（如接口名、参数关键字段、失败原因或错误码）。

8. 改动前后必须保持可回归  
- 至少做目标链路验证（协议链路/业务链路）。  
- Windows 本地无法全量构建时，至少保证格式正确和目标链路可验证；全量构建交给 Linux/CI。

9. 使用 Edit 工具时，oldString 必须最小化  
- `oldString` 只取**需要修改的那一行**或最小必要上下文（通常不超过 3 行），禁止把后续无关代码纳入 `oldString` 作为匹配上下文。  
- `newString` 必须包含 `oldString` 中所有**不需要改动**的行，不得遗漏。  
- 修改完成后，必须通过 `git diff` 或 Read 确认改动范围符合预期，没有误删无关代码。

---

## 2. SHOULD（推荐实践）

1. 复用现有代码模式，不引入新风格框架。  
2. 保持改动最小化，避免无关重构和大面积格式化。  
3. 新增配置表时，补齐 `primarykey.txt`、生成流程和必要 Loader。  
4. 协议改动优先使用项目现有脚本（如 `AllMake.sh proto`、`win/build_proto_dev.bat`）。

---

## 3. MUST NOT（明确禁止）

1. 未经用户确认，执行破坏性命令（如 `git reset --hard`、强制删除）。  
2. 在 reactor 主线程或 WorkerPool callback 内执行阻塞 I/O（磁盘读写、同步网络调用等）。  
   - 包括但不限于：gamesvr 主线程、`src/base/control` 的 control 命令执行 Handler（`Handler` 函数）。  
   - Handler 内部只做纯计算或内存操作；响应写回（如 `c.WriteToUDP`、`os.WriteFile`）必须在独立 goroutine 中执行。  
3. 只改 proto 不更新生成物就提交。  
4. 手工修改生成文件来规避生成流程。  
5. 把 model 逻辑塞进 app 层，或跨层随意放置代码。

---

## 4. 评审清单（Review Agent 必查）

> 评审范围约束：review **当前修改内容**时，不需要扫描以下目录：`pkg/`、`resourse/`、`protocol/configs/`。

1. 是否遵守 `AddTask -> AddCallback` 的异步+主线程回调模型？  
2. 是否存在协议枚举与 message 命名不一致？  
3. 是否漏更新 pb/register 生成文件？  
4. 是否有缺失的空指针/在线状态保护？  
5. 日志是否包含关键定位字段？  
6. 改动是否落在正确分层（apps/model/dao/protocol）？  
7. 是否引入了无关改动或潜在回归？

---

## 5. 触发式规则（给智能体）

1. 任务涉及协议（proto、命令字、注册）时，优先读取：`skills/protocol-and-register/SKILL.md`。  
2. 任务涉及 gamesvr 并发/异步时，优先读取：`skills/gamesvr-mainthread/SKILL.md`。  
3. 任务涉及 Redis 或 Lua 时，优先读取：`skills/redis-usage/SKILL.md`。  
4. 任务涉及服务结构与入口定位时，优先读取：`skills/server-architecture/SKILL.md`。

> 以上 4 条命中任意一条，应先读取对应 SKILL 再实施改动。
