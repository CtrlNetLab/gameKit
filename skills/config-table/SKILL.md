---
name: config-table
description: 游戏配置表的 Excel 设计、表结构约束、primarykey 维护、构建脚本与 ResMgr/预加载用法。用户涉及设计新配置表、调整字段、生成配置表 pb/json、补充读取接入时使用。
---

# 配置表（Excel）

用于设计配置表、生成 pb/json 数据源，并指导项目中读取配置。

## 执行原则

- 任务涉及“设计配置表”“新增 Excel 配置”“调整配置字段/类型”“生成配置表 pb/json”时，优先按本 skill 执行。
- agent 不直接手工生成、拼装或修改配置表相关 pb 产物，必须走项目既有构建流程。
- Windows 下统一进入仓库根目录的 `win` 目录后执行 `build_execl.bat` 和 `build_proto_dev.bat`；不要跳过任一步，也不要用临时脚本替代。

## Excel 设计规范

- 生成位置：`protocol/configs/resourse/配置表dev/NA/`
- 文件命名：`<sheet-name>-<中文说明>.xlsx`，例如 `words- words配置.xlsx`
- 如果多个配置表都属于同一业务域，优先放在同一张 Excel 的不同 Sheet 中维护，不要无必要拆成多张 Excel。
- Sheet 名称：配置用途英文（用于 pb/json 标识）
- 表头三行：
  1) 中文备注
  2) 字段名称
  3) 数据类型（当前仅支持 `int`、`string`、`[int]`、`[string]`）
- 新建表格需补充少量测试数据（至少 2-3 行），保证生成流程与读取逻辑可验证。

### 配置测试数据规则

- 数组字段：如类型为 `[int]`，默认测试值使用逗号分隔，例如：`1,2,3,4`
- 奖励相关字段：默认使用类型 `[string]`
- 奖励字段默认测试值：`4000001|100,4000002|100`
- 奖励值格式说明：`,` 分隔多组奖励，`|` 前为奖励 ID，`|` 后为奖励数量

参考脚本：`tools/excelToProtobuf/main.go`

## primarykey.txt 维护

- 新建表后，必须在 `protocol/configs/primarykey.txt` 中写入配置名称，否则不会生成 pb/json。
- 索引规则以文件备注为准；支持多字段索引，例如：
  - 字符串主键（参考脚手架 ConfigInitArgsConfig）: `ConfigInitArgsConfig:item.Name:string`
  - 多字段复合键: `int64(item.Id)<<32|int64(item.Step)<<24|int64(item.Sort):int64`

## 构建与生成

- Windows:
  1) 先进入当前仓库根目录下的 `win` 目录
  2) 执行 `.\build_excel.bat`
  3) 再执行 `.\build_proto.bat`
  4) 禁止 agent 直接生成配置表 pb，必须以上述 bat 流程产出 pb/json
- Linux:
  1) `bash AllMake.sh buildconfig`
  2) `bash AllMake.sh proto`

## 项目中使用

- 单键查询：
  - `configs.WordsConfigResMgr.Get(configId)`
- 获取全部：
  - `configs.MoneyItemConfigResMgr.GetAllData().Items`
- 字符串主键查询（参考 ConfigInitArgsConfig）:
  - `configs.ConfigInitArgsConfigResMgr.Get(name)`

## 预加载（复杂配置）

当配置表需要预处理数据时，在 `gamesvr/model/` 里新增 Loader：

- 命名：`<SheetName>ConfigLoader`（每个表只允许一个 Loader，避免覆盖）
- Loader 代码单独放在配置预加载文件中，不要和具体业务逻辑实现耦合在同一个文件里。
- 注册：

```go
type wordsConfigLoader struct {
configs.BaseResCallbackMgr
}

func (r *wordsConfigLoader) GetName() string {
return configs.GetTypeName(r)
}

var (
WordsConfigLoader = new(wordsConfigLoader)
)

func init() {
configs.ResLoadMgr.AddCallback(configs.WordsConfigResMgr.GetFileName(), WordsConfigLoader)
}

func (r *wordsConfigLoader) Check(iArr interface{}, start bool) error {
_, ok := iArr.(*configs.WordsConfigArray)
if !ok {
return fmt.Errorf("convert *configs.WordsConfigArray failed, i:%v", iArr)
}
return nil
}

func (r *wordsConfigLoader) Load(start bool) {
items := configs.WordsConfigResMgr.GetAllData().GetItems()
if items == nil || len(items) == 0 {
return
}
for _, config := range items {
// 预处理字段赋值
}
}
```
