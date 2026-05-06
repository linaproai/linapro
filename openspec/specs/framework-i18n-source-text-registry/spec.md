## 新增需求

### 需求：宿主必须为代码拥有的源文本命名空间提供显式注册机制
宿主系统 SHALL 在 `internal/service/i18n` 包中提供 `RegisterSourceTextNamespace(prefix, reason string)` 注册函数和对应的只读查询能力。业务模块必须在自己的 `init()` 中显式注册其代码拥有的源文本命名空间（如 `job.handler.`、`job.group.default.`）。`i18n` 包不得在自己的实现中硬编码任何特定业务模块的命名空间前缀。缺失翻译检查、覆盖源诊断和导入/导出 SHALL 通过查询此注册表识别"翻译键由代码源拥有的命名空间"。

#### 场景：业务模块通过 init 注册代码拥有的命名空间
- **当** `jobmgmt` 包在项目启动时执行 `init()` 时
- **则** 该包通过 `i18n.RegisterSourceTextNamespace("job.handler.", "code-owned cron handler display")` 注册其命名空间
- **且** 缺失检查可在不修改 `i18n` 包源码的情况下识别这些键为代码拥有

#### 场景：缺失检查基于注册表豁免代码拥有的命名空间
- **当** 系统对任何非默认目标语言（如 `en-US` 或 `zh-TW`）调用 `CheckMissingMessages` 且某些键属于已注册的代码拥有的命名空间时
- **则** 这些键不出现在缺失结果中
- **且** 这些键的显示回退由拥有模块的代码源文本处理，不要求每个目标语言冗余维护 JSON 键

### 需求：i18n 包不得再依赖特定业务模块命名空间前缀
宿主系统 SHALL NOT 允许 `i18n` 包中的任何函数（包括 `isSourceTextBackedRuntimeKey` 等辅助判断）硬编码 `job.handler.`、`job.group.default.` 或其他特定业务模块命名空间前缀。所有"此键由代码源拥有"的判断必须通过查询命名空间注册表获得。

#### 场景：删除 i18n 包对 jobmgmt 的反向依赖
- **当** 审查 `apps/lina-core/internal/service/i18n/` 中的任何源文件时
- **则** 不存在带业务模块特定前缀如 `job.handler.` 或 `job.group.default.` 的硬编码字符串
- **且** 文件改用命名空间注册表的查询接口进行判断
