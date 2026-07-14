## Why

参数设置页列表已按请求语言本地化内置参数的名称与描述，但编辑详情 `GetById` 故意回填库内中文 seed，导致英文环境下打开「修改」弹窗时输入框仍显示中文。根因是为防止本地化投影被写回 `sys_config`，却牺牲了编辑面展示一致性。

## What Changes

- 详情接口对 `name`/`remark` 按当前语言投影（与列表一致），**`value` 始终返回库内原文**。
- 内置参数更新时**忽略 `name`/`remark` 写回**，避免英文化展示文案污染权威存储。
- 前端编辑弹窗对内置参数将 `name`/`remark` 设为只读，主要可编辑字段为参数值。
- 补充单元测试与 E2E，覆盖英文环境编辑内置参数的元数据展示与保存不污染。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `config-management`：明确编辑详情的元数据本地化边界，以及内置参数 name/remark 不可写回策略。

## Impact

- 后端：`apps/lina-core/internal/service/sysconfig`（`GetById`、`Update`、i18n 测试）
- 前端：`apps/lina-vben/apps/web-antd/src/views/system/config`（编辑表单 schema）
- 测试：sysconfig 单元测试、`hack/tests/e2e/i18n` 或 `settings/config` E2E
- 无 API 路径/DTO 结构 **BREAKING** 变更；详情响应中内置 `name`/`remark` 语义从「库原文」调整为「请求语言投影」
