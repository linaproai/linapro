## 1. 数据库与生成代码

- [x] 1.1 将 `value_type`（默认 `text`）与 `options`（默认空串）并入宿主 `005-config-management.sql` 的 `sys_config` 建表定义
- [x] 1.2 在 `005-config-management.sql` 与 `011-scheduled-job-management.sql` 的内置参数 Seed `INSERT` 中直接写入 `value_type`/`options`（布尔、布局、主题、数字、长文案等，按 design D1/D8 映射）；不保留独立 `013` 增量文件
- [x] 1.3 执行 `make db.init`（或等价流程）与 `make dao`，确认 `entity`/`do`/`dao` 含新字段

## 2. 后端契约与领域校验

- [x] 2.1 扩展 `api/config/v1` 的 `ConfigItem`、Create、Update 请求/响应：`valueType`、`options[]{label,value}`
- [x] 2.2 扩展 `sysconfig` 的 `CreateInput`/`UpdateInput`/投影模型，Create/Update/Get/List/GetByKey 读写类型元数据
- [x] 2.3 实现封闭 `valueType` 校验、options JSON 结构校验、按类型的 value 校验（boolean/number/select/radio/multi_select）
- [x] 2.4 内置记录禁止修改 `valueType`/`options`；非内置可改；与既有托管键 `validateManagedConfigValue` 叠加
- [x] 2.5 租户覆盖创建时从平台行继承 `value_type`/`options`
- [x] 2.6 导入/导出/导入模板增加 `valueType`、`options` 列，列头走 `config.field.*` 翻译键；缺省类型为 `text`
- [x] 2.7 插件 `SetValue` 新建行时默认 `value_type=text`、`options=''`（不扩展插件 API 除非必要）

## 3. 前端管理面

- [x] 3.1 更新 `api/system/config` 模型与调用类型
- [x] 3.2 改造 `config-modal`：按 `valueType` 动态渲染 value 组件；新增时可选择类型；枚举类型可编辑 options
- [x] 3.3 内置参数编辑时锁定类型与 options 为只读；`multi_select` 与 `;` 序列化互转
- [x] 3.4 `richtext`：优先复用已有富文本组件，否则降级增强 Textarea（保留类型枚举）
- [x] 3.5 补充中英文（及既有 locale）i18n：字段名、类型名、options 编辑器、校验提示；`config.field.valueType`/`options`；主要内置枚举 option 展示键（如采用 `config.<key>.option.<value>`）

## 4. 测试与验证

- [x] 4.1 后端单测：类型校验、内置类型锁定、select/boolean 非法值、导入缺省类型、multi_select 分号序列化
- [x] 4.2 E2E — TC007 参数值类型化编辑：新增 `hack/tests/e2e/settings/config/TC007-config-param-value-types.ts`
  - [x] TC-7a：编辑 `sys.auth.forgetPasswordEnabled` 使用布尔/开关组件并可保存
  - [x] TC-7b：编辑 `sys.auth.loginPanelLayout` 通过下拉/单选选择合法布局值
  - [x] TC-7c：自定义 select 参数创建后编辑可从 options 选择
- [x] 4.3 按需扩展 `ConfigPage` POM 以支持类型化控件交互
- [x] 4.4 运行相关 `go test`、前端类型检查/必要用例、`openspec validate config-param-value-types --strict`

## 5. 影响确认

- [x] 5.1 确认 runtime snapshot / `GetRaw` 无行为变化（仅字符串 value）
  - 结论：热路径仍只读 `value` 字符串；`value_type`/`options` 仅管理 CRUD/导入导出使用，不进入 runtime snapshot。
- [x] 5.2 记录 i18n、缓存一致性、数据权限、跨平台工具影响判断（预期：i18n 有影响；缓存/数据权限/工具无实质变化，除非导入导出列变更需回归）
  - **i18n**：有影响——`config.field.valueType/options`、错误码、前端 pages 文案、内置 option 展示键已补齐 zh-CN/en-US。
  - **缓存一致性**：无影响——未改 revision/snapshot 协调语义，仅写路径仍走既有 `MarkRuntimeParamsChanged`。
  - **数据权限**：无影响——仍用既有 tenant scope / fallback 可见性。
  - **开发工具跨平台**：无影响——仅新增 SQL 与业务代码，无 Makefile/脚本语义变更。
  - **测试**：已补单元测试与 E2E TC007，并同步 TC006 导出表头断言。
- [x] 5.3 任务完成后执行 `lina-review` 审查

## Feedback

- [x] **FB-1**: 参数类型下拉在选中短标签后弹层宽度过窄，选项文案显示不全
- [x] **FB-2**: 选项列表改为简单行格式（标签=值），前后端均可解析，避免用户手写 JSON
- [x] **FB-3**: 多选类型参数键值出现空标签；选项列表调整到参数类型下方、参数键值上方
- [x] **FB-4**: 富文本类型接入宿主 TiptapEditor，不再降级为普通 Textarea
- [x] **FB-5**: 富文本/长文编辑弹窗采用类型驱动密度策略（加宽、全屏、视口相对编辑高度），避免固定矮框
