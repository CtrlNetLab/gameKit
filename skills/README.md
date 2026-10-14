# Project Skills Index

## Available skills
- run-and-env: 使用 run/ 脚本与配置模板完成本地环境配置、渲染服务配置、启动/停止服务。  
  file: `skills/run-and-env/SKILL.md`
- protocol-and-register: 修改,新增协议 proto、生成 pb.go 输出，并维护命令字枚举与消息注册规则。  
  file: `skills/protocol-and-register/SKILL.md`
- server-architecture: gamekit 服务架构总览：入口、配置与主要服务关系。  
  file: `skills/server-architecture/SKILL.md`
- gamesvr-mainthread: gamesvr 主线程模型与协程使用规则，包括 WorkerPool 的异步任务与回调约束。  
  file: `skills/gamesvr-mainthread/SKILL.md`
- redis-usage: Redis 使用规范、异步调用约束、Lua 脚本加载与 ExecScript 约定。  
  file: `skills/redis-usage/SKILL.md`
- config-table: 游戏配置表的 Excel 设计、生成规则、primarykey 维护、构建脚本与 ResMgr/预加载用法。  
  file: `skills/config-table/SKILL.md`
- offline-data-mgr: 玩家离线数据处理规范：离线写入 Redis、登录拉取补发、处理器映射与新增类型扩展流程。  
  file: `skills/offline-data-mgr/SKILL.md`
- add-new-event: gamesvr 事件订阅与派发（EventType 枚举 → pb 重新生成 → AddEventHandler 注册 → DispatchEvent 派发）。  
  file: `skills/add-new-event/SKILL.md`
- log-analysis: 游戏服务器日志分析：先确认日志位置（本地默认 `run/<service>/log/`，玩家相关默认 gamesvr），再识别进程并限定代码查找范围，识别代码逻辑问题与配置错误。Loki 查询使用 `service_type` label。  
  file: `skills/log-analysis/SKILL.md`  
  agent: `.opencode/agents/log-analysis.md`（subagent，内网日志通过 Loki MCP 查询）
## Notes
- 路径均相对于仓库根目录。
- 使用前按路径打开对应 `SKILL.md`。
- 若任务明显匹配某 skill，优先使用该 skill。
