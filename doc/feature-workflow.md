# 服务端新功能新增流程框架

> 适用场景：gamesvr 新增客户端协议接口 + 业务逻辑 + GM 工具的可复用规范。
> 文中以「示例功能 Example」贯穿演示，实际开发时替换为真实功能名。

## 流程总览

```text
┌──────────────────────────────────────────────────────────┐
│  0. 方案决策：修改旧协议 vs 新建独立协议                    │
├──────────────────────────────────────────────────────────┤
│  1. proto 层：定义协议号 + 请求/响应结构体                 │
│     ├── (必选) cs_cmdvalue.proto 注册协议号               │
│     ├── (必选) ac_cs_<模块>.proto 请求/响应消息体          │
│     ├── (可选) ac_account.proto 新增存储字段               │
│     ├── (可选) ss_eventtype.proto 扩展事件参数             │
│     ├── (可选) cs_retcode.proto 新增错误码                 │
│     └── (可选) ss_type.proto 新增物品流水枚举              │
├──────────────────────────────────────────────────────────┤
│  2. 生成 pb Go 代码（./AllMake.sh proto）                 │
├──────────────────────────────────────────────────────────┤
│  3. base/ 层：新增 Chunk 常量（ChunkConstant.go）          │
├──────────────────────────────────────────────────────────┤
│  4. model/ 层：实现核心业务逻辑（Player.go / 模块文件）     │
├──────────────────────────────────────────────────────────┤
│  5. apps/ 层：新建 <模块>App.go，注册 Handler              │
├──────────────────────────────────────────────────────────┤
│  6. （跨功能）修改现有 apps Handler（如 LoginApp.go）       │
├──────────────────────────────────────────────────────────┤
│  7. ctrl/ 层：添加 GM 测试命令（HandleControl.go）          │
│     ├── (推荐) 直接查询型：show<FeatureName>              │
│     └── (推荐) 走协议型：test<FeatureName>                │
├──────────────────────────────────────────────────────────┤
│  8. 编译验证：go build ./...（对应服务模块）               │
└──────────────────────────────────────────────────────────┘
```

---

## 步骤详解

### 0. 方案决策：修改旧协议 vs 新建独立协议

这是进入具体编码前**最重要的一步**，决定了后续所有工作。

| 方案 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| **A. 修改旧协议**（追加字段） | 功能高度耦合，逻辑自然扩展 | 改动集中，client 侧改动小 | 向后兼容风险；旧逻辑被迫感知新字段 |
| **B. 新建独立协议**（新 cmdvalue + 新 handler） | 功能**逻辑独立**，与旧协议无强耦合 | 零兼容风险；职责单一；易于测试/回滚 | 多一套协议流程 |

**推荐采用 B（新建独立协议）**，除非满足以下**全部**条件：
- 新功能与旧协议的业务语义**不可分割**（例如同一个界面操作需要同时提交两类数据）
- 新功能**必须**在旧逻辑的同一个事务中处理
- 与客户端约定**明确**且**有版本控制**

---

### 1. Proto 层协议定义

#### 1.1 注册协议号

**文件：** `protocol/share/cs/cs_cmdvalue.proto`

```protobuf
// 新增功能统一使用 105xx/5xxxx 区间（参考历史习惯）
// 请求号惯例：105xx，响应号惯例：505xx
ECV_CS_AC_<FeatureName>Req = 105xx;
ECV_CS_AC_<FeatureName>Res = 505xx;
```

**参考（示例功能）：**
```protobuf
// 示例功能查询
ECV_CS_AC_ExampleReq = 10501;
ECV_CS_AC_ExampleRes = 50501;
```

#### 1.2 定义请求/响应消息体

**文件：** `protocol/share/cs/ac_cs_<模块>.proto`（或对应模块文件）

**响应消息体约定：**
- **必须**包含 `int32 code = 1`（返回码）
- **必须**包含 `repeated cs.Extend extend = N [json_name = "extend"]`（有数据变更时推送增量）
- 查询类响应**必须**包含足够让客户端独立渲染的信息

```protobuf
// 参考：示例功能查询响应
message CS_AC_ExampleRes {
    int32 code = 1;
    int32 exampleId = 2;
    repeated string exampleList = 3;
    repeated cs.Extend extend = 6 [json_name = "extend"];
}
```

#### 1.3 Database proto 新增存储字段

**文件：** `protocol/share/database/ac_account.proto`

当需要持久化新数据时，在 `Account` 或对应子 message 中新增字段。

**新增 message 类型惯例：**
- 如果新字段是简单类型，直接追加到所在父 message
- 如果新字段是结构体类型，先定义新的子 message，再在父 message 中用 `map<k,v>` / `repeated` 引用

```protobuf
// 参考：新增子 message + map 字段
message ExampleTeam {
    repeated string members = 1;
}
// Account 中（接续已有最大 field number）
map<int32, ExampleTeam> exampleData = 95;
```

**Field number 规则：** 接续已有最大 field number +1（避免冲突）。

#### 1.4 ss_eventtype.proto 扩展事件参数

**文件：** `protocol/ss/ss_eventtype.proto`

当新功能需要**将额外参数传递到已有事件**时，在对应 event message 中扩展字段。**不要新建事件类型**，优先扩展现有事件（EventType 枚举值即协议，删除/改编号会影响存量玩家数据兼容）。

```protobuf
// 参考：EvnParam 扩展示例参数
message EvnParam {
    EvnActivity Activity = 1;
    int32 exampleId = 2;   // 新增：示例功能参数
}
```

#### 1.5 cs_retcode.proto 新增错误码

**文件：** `protocol/share/cs/cs_retcode.proto`

**区间约定：** 新功能错误码建议使用 **13000~13099** 区间。

```protobuf
// 参考：
ERROR_EXAMPLE_NOT_OPEN = 13000;
ERROR_EXAMPLE_LIMIT = 13001;
```

#### 1.6 （可选）ss_type.proto 新增物品流水枚举

**文件：** `protocol/ss/ss_type.proto`

当业务涉及物品变更、需要追溯原因时，在 `EnmItemReason` 枚举中新增 `Add_<功能名>` / `Del_<功能名>` 条目（脚手架默认未包含该枚举，首次使用时新增）。

---

### 2. 生成 pb Go 代码

Proto 文件修改完成后，**必须**执行：

```bash
./AllMake.sh proto
```

该命令会：
- 重新生成 `pkg/` 目录下对应的 `.pb.go` 文件（cs / ss / database / configs / oss / game_third）
- 重跑 `tools/register` 刷新命令字与消息体的注册表
- 使新的 proto 枚举和结构体在 Go 代码中可用

> ⚠️ **验证：** 执行后检查 `pkg/cs/`、`pkg/database/`、`pkg/ss/` 等目录是否包含新字段。
>
> ⚠️ **前置条件：** 需要 protoc 与 protoc-gen-go（v1.34.2）在 PATH 中；Windows 开发机可参考 `win/build_proto.bat`。

---

### 3. base/ 层：新增 Chunk 常量

**文件：** `src/base/universal/ChunkConstant.go`

当新功能有数据变更需要推送到客户端时，新增 Chunk 类型常量：

```go
// 示例功能数据块
CHUNK_TYPE_EXAMPLE int32 = 111
```

**编号规则：** 接续已有最大编号 +1。

数据推送使用 `cs.Extend` 消息（`chunk` + `operation` + `values`）由响应消息带给客户端；
项目可在 gamesvr model 层实现统一 Push 辅助函数（参考 ac-server 的 `model.PushStrings` 模式）。

---

### 4. Model 层：核心业务逻辑

**文件：** `src/gamesvr/scripts/model/Player.go`（通用逻辑）或 `<模块>Mgr.go`（模块逻辑）

#### 4.1 业务函数设计原则

| 类别 | 命名 | 说明 |
|------|------|------|
| **查询类** | `Get<FeatureName>By<Key>` | 返回数据，无副作用 |
| **写入类** | `Set<FeatureName>` | 直接修改 data 对象 |
| **条件判断** | `Is<Condition>` | 返回 bool，用于校验 |
| **校验类** | `Check<Condition>` | 返回 bool，带副作用 |
| **计算类** | `Get<Metric>` | 返回计算结果 |

#### 4.2 函数放置位置

与 Player 基础能力紧耦合的 → 直接放在 `Player.go`
功能独立的 → 放在对应的 `<模块>Mgr.go`

#### 4.3 辅助函数的使用

应优先使用 Player 对象上已有的辅助方法，而非自己重新遍历：

| 场景 | 用法 |
|------|------|
| 取账号 ID | `player.GetAccId()` |
| 取连接索引 | `player.GetConnIdx()` |
| 判断在线状态 | `player.IsOnlineOK()` |
| 推送变更 | 响应消息中携带 `repeated cs.Extend`（见步骤 3） |
| 触发事件 | `utils.DispatchEvent(&utils.Event{EventId: ..., IObj: player, EventParam: ...})` |

#### 4.4 配置获取方式

优先复用已存在的 `configs.*ResMgr`，避免新建配置解析：

```go
// 推荐：使用已有 ResMgr
cfg := configs.WordsConfigResMgr.Get(id)

// 不推荐：新建配置表 + ResMgr（除非现有配置都不适用）
```

> 新增配置表的完整流程见 `skills/config-table`。

---

### 5. Apps 层：协议注册 + Handler 函数

#### 5.1 文件命名

**格式：** `<模块>App.go`（注意 `App` 后缀）

**参考：**
- `ExampleApp.go` ✅（推荐）
- ~~`Example.go`~~ ❌（与包内其他文件风格不一致）

#### 5.2 注册协议

**文件：** `src/gamesvr/scripts/apps/<模块>App.go`

```go
func init() {
    // <功能描述>
    RegistClientHandle(cs.EnmCmdValue_ECV_CS_AC_ExampleReq, HandleExampleReq)
}
```

#### 5.3 Handler 实现模板

```go
func HandleExampleReq(player *model.Player, head *cs.CSHead, iMsg interface{}) {
    if player == nil || head == nil || !player.IsOnlineOK() {
        logger.Log.Error("%s|[EXAMPLE] request fail, player is empty or player not online.", player.String())
        return
    }
    request, ok := iMsg.(*cs.CS_AC_ExampleReq)
    if !ok || request == nil {
        logger.Log.Error("%s|[EXAMPLE] request fail, request params is invalid.", player.String())
        return
    }

    // 使用闭包统一返回路径
    code := func() cs.EnmRetCode {
        // 1. 校验
        // 2. 业务逻辑
        // 3. 数据变更推送(响应消息 Extend)
        // 4. 触发事件
        return cs.EnmRetCode_SUCCESS
    }()

    res := &cs.CS_AC_ExampleRes{Code: int32(code)}
    player.SendClientMsg(cs.EnmCmdValue_ECV_CS_AC_ExampleRes, head.SeqId, code, "", res)
}
```

#### 5.4 校验顺序惯例

对于一个设置类协议，校验应按**优先级从低到高**（越前置的校验越快失败）：

1. 参数合法性（exampleId == 0？列表为空？）
2. 功能是否已解锁
3. 资源是否足够
4. 业务规则校验
5. 与外部状态交叉校验

#### 5.5 日志锚点约定

**格式：** `|[<TAG>]` 其中 `TAG` 是模块/功能英文简称，**大写**。

```go
logger.Log.Error("%s|[EXAMPLE] example not open ...", player.String())
logger.Log.Info("%s|[EXAMPLE] success ...", player.String())
```

#### 5.6 数据变更推送与事件

数据变更后按需执行：

```go
// 1. 响应消息中携带 Extend 通知客户端增量变更（见步骤 3）

// 2. DispatchEvent 触发服务端事件（用于数据同步等）
utils.DispatchEvent(&utils.Event{
    EventId: ss.EventType_EId_LoginSuccess,
    IObj:    player,
    EventParam: &ss.EvnParam{
        Activity: &ss.EvnActivity{ActId: uint32(exampleId)},
    },
})
```

---

### 6. 跨功能修改现有 apps Handler

这是新手最容易遗漏的步骤。新功能的引入往往需要**侵入已有流程**。

#### 6.1 清单

实现新功能后，检查以下现有 Handler（以脚手架实际文件为准）：

| 现有 Handler | 是否需要修改 | 典型修改内容 |
|-------------|------------|-------------|
| `HandleLoginReq`（LoginApp.go） | 视情况 | 登录流程中初始化新功能数据 |
| `HandleLoginRes`（LoginApp.go） | 视情况 | 响应中补充新功能字段 |
| 其他业务 App Handler | 视情况 | 合并/清理新功能涉及的数据 |

---

### 7. Ctrl 层：GM 测试工具

**文件：** `src/gamesvr/scripts/ctrl/HandleControl.go`

#### 7.1 推荐模式：两种 GM 命令

| 模式 | 命令命名 | 优点 |
|------|---------|------|
| **直接查询** | `show<FeatureName>` | 不依赖 player 在线、不走协议 |
| **走协议** | `test<FeatureName>` | 完整模拟客户端请求路径 |

推荐的组合实践：**两种都加**——直接查询用于快速调试，走协议用于端到端验证。

#### 7.2 注册命令

在 `func Init()` 的 `ctrls` map 中添加：

```go
"test<FeatureName>": {
    Usage:   "test<FeatureName>#accId,param1,param2,...",
    Desc:    "描述",
    CmdType: utils.CmdType_Gm,
    GmType:  utils.GmType_Root,
    Handle:  test<FeatureName>,
    Tags:    []string{control.TagTest},
},
```

#### 7.3 实现 GM Handler

**直接查询型**（使用 `param` 中的 player）：
```go
func showExample(args []string, cmdType int, param interface{}) (string, error) {
    player, ok := param.(*model.Player)
    if !ok || !player.IsOnlineOK() {
        return "not found player", nil
    }
    exampleId := utils.StrToInt32(args[1])
    data := player.GetExampleData(exampleId)
    return fmt.Sprintf("data=%v", data), nil
}
```

**走协议型**（通过 accId 查找 player）：
```go
func testGetExample(args []string, cmdType int, param interface{}) (string, error) {
    accId := utils.StrToUint64(args[1])
    player := model.PlayerMgr.GetPlayerByAccId(accId)
    if player == nil || !player.IsOnlineOK() {
        return "player not online", nil
    }
    req := &cs.CS_AC_ExampleReq{ExampleId: utils.StrToInt32(args[2])}
    apps.HandleExampleReq(player, &cs.CSHead{}, req)
    return "success", nil
}
```

**注意：** 走协议型 GM 通过 `model.PlayerMgr.GetPlayerByAccId(accId)` 获取 player，不是通过 `param`。

---

### 8. 编译验证

```bash
# gamesvr 模块（wsgnet 为 Linux-only，开发机需交叉编译）
cd src/gamesvr && GOOS=linux GOARCH=amd64 go build ./...

# pkg 模块（proto 生成后）
cd pkg && go build ./...
```

**注意：**
- 项目为每个服务独立的 go.mod，需要用 `cd` 进入对应目录执行
- 不要使用项目根目录的 `go build ./...`（会报 `no main module`）

---

### 9. 重启服务

编译通过后重启对应服务，使变更生效：

```bash
# 进入 run 目录（按实际部署路径调整）
cd <项目根目录>/run

# 重启 gamesvr（需 sudo 权限）
sudo ./admin.sh restart
```

**注意：**
- 由编译到重启属于**高频循环**：改代码 → `go build` → `sudo ./admin.sh restart`
- 重启前确认 `go build` 无错误，否则服务启动会失败

---

## 命名规范汇总

| 层级 | 命名规则 | 示例 |
|------|---------|------|
| **proto cmdvalue** | `ECV_CS_AC_<FeatureName>Req/Res` | `ECV_CS_AC_ExampleReq` |
| **proto message** | `CS_AC_<FeatureName>Req/Res` | `CS_AC_ExampleReq` |
| **apps/ 文件名** | `<FeatureName>App.go` | `ExampleApp.go` |
| **apps/ Handler** | `Handle<FeatureName>Req` | `HandleExampleReq` |
| **model/ 业务函数** | `Get<FeatureName>` / `Is<Condition>` / `Check<Condition>` | `GetExampleData`, `IsExampleOpen` |
| **model/ 日志 TAG** | `[<TAG>]`（大写） | `[EXAMPLE]` |
| **ctrl/ 命令名** | `show<FeatureName>` / `test<FeatureName>` | `showExample`, `testGetExample` |
| **ctrl/ Handler** | `show<FeatureName>` / `test<FeatureName>`（小写驼峰） | `showExample`, `testGetExample` |
| **database proto 字段** | 驼峰小写首字母 | `exampleData` |
| **Chunk 常量** | `CHUNK_TYPE_<FEATURE>` | `CHUNK_TYPE_EXAMPLE` |
| **错误码枚举** | `ERROR_<DESCRIPTION>` | `ERROR_EXAMPLE_NOT_OPEN` |
| **ss_type 枚举** | `Add_<FEATURE_NAME>` / `Del_<FEATURE_NAME>` | `Add_ExampleReward` |
| **补丁版本号常量** | `PatchVersion<Feature>V<Version>` | `PatchVersionExampleV1` |
| **补丁处理函数** | `PatchVersion_<Feature>V<Version>` | `PatchVersion_ExampleV1` |
| **补丁日志 TAG** | `[<FEATURE>_PATCH]` | `[EXAMPLE_PATCH]` |
| **预处理 getter** | `Get<Feature>` | `GetExampleConfig()` |
| **预处理判断** | `Is<Condition>` | `IsExampleStage()` |
| **预处理日志 TAG** | `[<MODULE>_MGR]` | `[EXAMPLE_MGR]` |

---

## 验收清单

完成新功能后，逐项检查：

### 编译与重启（必须执行）
```bash
# 1. 重新生成 proto 文件（如有协议变更）
./AllMake.sh proto

# 2. 整体编译
./AllMake.sh build

# 3. 编译通过后重启服务
cd run && sudo ./admin.sh restart
```

### Proto 层
- [ ] 协议号（`cs_cmdvalue.proto`）已定义
- [ ] 请求/响应 message（`ac_cs_*.proto`）已定义
- [ ] 数据库字段（如有）已在 `ac_account.proto` 中定义
- [ ] 事件参数（如有）已在 `ss_eventtype.proto` 中扩展
- [ ] 错误码（如有）已在 `cs_retcode.proto` 中定义
- [ ] 物品流水枚举（如有）已在 `ss_type.proto` 中定义
- [ ] `./AllMake.sh proto` 已执行，无编译错误

### 服务端 Go
- [ ] Chunk 常量已在 `ChunkConstant.go` 中定义
- [ ] model/ 业务函数已实现
- [ ] apps/ Handler 已注册（`RegistClientHandle`）且响应 cmdvalue 正确
- [ ] 日志 TAG 统一（如 `[EXAMPLE]`）
- [ ] 数据变更已通过 Extend 推送 / DispatchEvent 触发
- [ ] 已有 Handler 已适配

### GM 与编译
- [ ] ctrl/ GM 命令已注册（至少一个直接查询型或走协议型）
- [ ] 整体编译通过：`./AllMake.sh build`（在项目根目录执行）
- [ ] 编译成功后重启服务：`sudo ./admin.sh restart`（在 run/ 目录执行）

---

## 补丁工作流（老玩家数据兼容）

> 适用场景：新功能涉及数据库字段新增或数据结构变更时，需要兼容老玩家历史数据。
> 脚手架未内置玩家补丁框架，以下为 ac-server 沉淀的可复用模式：在 gamesvr model 层新建
> `<模块>PatchData.go` 并按本模式实现。

### 流程总览

```text
┌──────────────────────────────────────────────────────────┐
│  1. 判断是否需要补丁                                       │
│     ├── 新字段默认值可接受（0=false）→ 不需要补丁           │
│     └── 新字段需要从现有数据计算 → 需要补丁                 │
├──────────────────────────────────────────────────────────┤
│  2. 定义补丁版本号                                         │
│     └── <模块>PatchData.go: 新增 PatchIDType 常量          │
├──────────────────────────────────────────────────────────┤
│  3. 注册补丁入口                                           │
│     └── <模块>PatchData.go: PatchEntryMap 追加条目         │
├──────────────────────────────────────────────────────────┤
│  4. 实现补丁函数                                           │
│     └── <模块>PatchData.go: PatchXxx(player *Player)      │
├──────────────────────────────────────────────────────────┤
│  5. 调用时机                                               │
│     └── PatchPlayer(player): 玩家登录时自动执行             │
└──────────────────────────────────────────────────────────┘
```

### 详细步骤

#### 1. 判断是否需要补丁

| 情况 | 是否需要补丁 | 说明 |
|------|-------------|------|
| 新字段默认值可接受 | ❌ 不需要 | 如 `int64` 默认 0、`bool` 默认 false 表示"未开通" |
| 新字段需要从现有数据计算 | ✅ 需要 | 如根据现有关卡进度计算满星状态 |
| 新字段影响排序/展示逻辑 | ✅ 需要 | 如排行榜新增二级排序字段 |
| 新字段仅影响新功能 | ❌ 不需要 | 老玩家使用新功能时再初始化 |

#### 2. 定义补丁版本号

**文件：** `src/gamesvr/scripts/model/<模块>PatchData.go`

```go
// 在 const 区域追加，接续已有最大版本号 +1
const (
    // ... 现有补丁版本号 ...

    // 新增：本次需求补丁版本号
    PatchVersionXxxV1 PatchIDType = 55
)
```

#### 3. 注册补丁入口

**文件：** `src/gamesvr/scripts/model/<模块>PatchData.go`

```go
// 在 PatchEntryMap 切片末尾追加
var (
    PatchEntryMap = []*PatchEntry{
        // ... 现有条目 ...

        // 新增：本次需求补丁入口
        {PatchVersionXxxV1, 0, PatchVersion_XxxV1},
    }
)
```

**参数说明：**
- `PatchId`：补丁版本号（唯一标识）
- `BeforeCreateAt`：0=所有玩家执行；非0=仅在该时间戳之前创建的玩家执行
- `PatchHandler`：补丁处理函数

#### 4. 实现补丁函数

**文件：** `src/gamesvr/scripts/model/<模块>PatchData.go`

```go
// PatchVersion_XxxV1 初始化老玩家的新字段
func PatchVersion_XxxV1(player *Player) {
    if player == nil || !player.IsOnlineOK() || player.GetAccountData() == nil {
        return
    }

    // 计算并保存新字段
    // ... 业务逻辑 ...

    logger.Log.Info("%s|[XXX_PATCH] success, newField=%v", player.String(), newValue)
}
```

#### 5. 调用时机

补丁在 `PatchPlayer(player)` 中按版本号顺序自动执行，无需额外调用。

**执行流程：**
```
玩家登录
  → PatchPlayer(player)
    → 遍历 PatchEntryMap
      → 按 PatchId 升序执行未执行过的补丁
        → 更新 basicInfo.PatchVersion
```

### 命名规范

| 元素 | 命名规则 | 示例 |
|------|---------|------|
| **版本号常量** | `PatchVersion<Feature>V<Version>` | `PatchVersionExampleV1` |
| **补丁入口** | `{版本号, 时间戳, 处理函数}` | `{PatchVersionExampleV1, 0, PatchVersion_ExampleV1}` |
| **处理函数** | `PatchVersion_<Feature>V<Version>` | `PatchVersion_ExampleV1` |
| **日志 TAG** | `[<FEATURE>_PATCH]` | `[EXAMPLE_PATCH]` |

---

## 预处理工作流（配置加载时计算）

> 适用场景：需要在服务启动时预计算配置数据，避免运行时重复遍历。

### 流程总览

```text
┌──────────────────────────────────────────────────────────┐
│  1. 判断是否需要预处理                                     │
│     ├── 配置数据需要遍历计算 → 需要预处理                   │
│     └── 计算结果运行时只读 → 需要预处理                     │
├──────────────────────────────────────────────────────────┤
│  2. 选择预处理位置                                         │
│     ├── 已有 *Mgr 结构体 → 扩展字段                        │
│     └── 无合适结构体 → 新建 *Mgr 或使用 init()             │
├──────────────────────────────────────────────────────────┤
│  3. 定义预计算字段                                         │
│     └── <模块>Mgr.go: 在对应 *Mgr 中新增字段               │
├──────────────────────────────────────────────────────────┤
│  4. 实现计算逻辑                                           │
│     └── <模块>Mgr.go: 在 Load() 方法中完成计算             │
├──────────────────────────────────────────────────────────┤
│  5. 提供公共接口                                           │
│     └── <模块>Mgr.go: 新增 getter 函数                     │
└──────────────────────────────────────────────────────────┘
```

### 详细步骤

#### 1. 判断是否需要预处理

| 情况 | 是否需要预处理 | 说明 |
|------|---------------|------|
| 配置数据需要遍历计算 | ✅ 需要 | 如计算最终关卡、满星总数 |
| 计算结果运行时只读 | ✅ 需要 | 如配置表映射关系 |
| 配置数据实时更新 | ❌ 不需要 | 使用时实时查询 |
| 计算逻辑简单 | ❌ 不需要 | 直接在业务代码中实现 |

#### 2. 选择预处理位置

**文件：** `src/gamesvr/scripts/model/<模块>Mgr.go`

| 场景 | 位置选择 | 说明 |
|------|---------|------|
| 已有 `xxxMgr` 等结构体 | 扩展字段 | 最优选择，复用现有加载机制 |
| 无合适结构体 | 新建 `*Mgr` 或 `init()` | 简单场景用 init() |

#### 3. 定义预计算字段

**文件：** `src/gamesvr/scripts/model/<模块>Mgr.go`

```go
type exampleMgr struct {
    // ... 现有字段 ...

    // 新增：预计算字段
    ExampleMaxId int32
    ExampleMap   map[int32]*ExampleData
}
```

#### 4. 实现计算逻辑

**文件：** `src/gamesvr/scripts/model/<模块>Mgr.go`

```go
func (r *exampleMgr) Load(start bool) {
    // ... 现有逻辑 ...

    // ========== 新增：预计算 ==========
    for _, item := range configs.ExampleConfigResMgr.GetAllData().GetItems() {
        if item.Id > r.ExampleMaxId {
            r.ExampleMaxId = item.Id
        }
        r.ExampleMap[item.Id] = item
    }

    logger.Log.Info("[EXAMPLE_MGR] load success, ExampleMaxId=%d", r.ExampleMaxId)
}
```

#### 5. 提供公共接口

**文件：** `src/gamesvr/scripts/model/<模块>Mgr.go`

```go
// GetExampleMaxId 获取示例配置最大 ID
func GetExampleMaxId() int32 {
    return ExampleMgr.ExampleMaxId
}
```

### 命名规范

| 元素 | 命名规则 | 示例 |
|------|---------|------|
| **预计算字段** | 语义化命名 | `ExampleMaxId`, `ExampleMap` |
| **getter 函数** | `Get<Feature>` | `GetExampleMaxId()` |
| **日志 TAG** | `[<MODULE>_MGR]` | `[EXAMPLE_MGR]` |

---

## 补丁与预处理工作流对比

| 维度 | 补丁工作流 | 预处理工作流 |
|------|-----------|-------------|
| **触发时机** | 玩家登录时 | 服务启动/重启时 |
| **执行频率** | 每个玩家执行一次 | 服务生命周期内执行一次 |
| **数据来源** | 玩家存档数据 | 配置表数据 |
| **存储位置** | 玩家 Account 数据 | 内存中的 *Mgr 结构体 |
| **典型场景** | 新增字段初始化 | 配置表遍历计算 |
| **文件位置** | `<模块>PatchData.go` | `<模块>Mgr.go` |
