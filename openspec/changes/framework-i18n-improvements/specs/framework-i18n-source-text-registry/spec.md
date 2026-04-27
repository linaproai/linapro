## ADDED Requirements

### Requirement: 宿主必须提供代码源文案命名空间显式注册机制
宿主系统 SHALL 在 `internal/service/i18n` 包内提供 `RegisterSourceTextNamespace(prefix, reason string)` 注册函数与对应只读查询能力,业务模块 MUST 在自身 `init()` 中显式注册其代码源文案命名空间(例如 `job.handler.`、`job.group.default.`)。`i18n` 包 MUST NOT 在自身实现中硬编码任何具体业务模块的命名前缀。缺失翻译检查、覆写来源诊断和导入导出 SHALL 通过查询该注册表识别"由代码源拥有翻译键"的命名空间。

#### Scenario: 业务模块通过 init 注册代码源命名空间
- **WHEN** 项目启动时 `jobmgmt` 包执行 `init()`
- **THEN** 该包通过 `i18n.RegisterSourceTextNamespace("job.handler.", "code-owned cron handler display")` 注册其命名空间
- **AND** 不需要修改 `i18n` 包源码也能让缺失检查识别这些键由代码源拥有

#### Scenario: 缺失检查根据注册表豁免代码源命名空间
- **WHEN** 系统对任意非默认目标语言(如 `en-US` 或 `zh-TW`)调用 `CheckMissingMessages` 且某些键属于已注册的代码源命名空间
- **THEN** 这些键不出现在缺失结果中
- **AND** 这些键的展示兜底由 owning module 的代码源文案负责,不要求每个目标语言都重复维护 JSON 键

### Requirement: i18n 包必须不再依赖具体业务模块的命名前缀
宿主系统 SHALL 不允许 `i18n` 包内的任何函数(包括 `isSourceTextBackedRuntimeKey` 这类辅助判定)硬编码 `job.handler.`、`job.group.default.` 或其他具体业务模块的命名前缀。所有"该键由代码源拥有"的判定 MUST 通过查询命名空间注册表得到。

#### Scenario: 删除 i18n 包内对 jobmgmt 的反向依赖
- **WHEN** 审查 `apps/lina-core/internal/service/i18n/` 包内任意源文件
- **THEN** 不存在以 `job.handler.`、`job.group.default.` 等业务模块特定字符串为前缀的硬编码
- **AND** 该文件改为通过命名空间注册表的查询接口判定
