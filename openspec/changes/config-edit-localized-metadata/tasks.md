## 1. 后端详情投影与写回保护

- [x] 1.1 抽取 raw 按 ID 加载；对外 `GetById` 本地化 `name`/`remark`，**不**投影 `value`
- [x] 1.2 `Update`（及需要未投影实体的 mutation）使用 raw 加载；内置记录忽略 `name`/`remark` 写回
- [x] 1.3 更新 `sysconfig` 单元测试：详情 en-US 元数据、value 保持原文、内置 Update 不污染 name/remark

## 2. 前端编辑表单

- [x] 2.1 内置参数编辑时 `name`、`remark` 字段只读（disabled）
- [x] 2.2 确认回填仍使用详情 API 投影结果，无前端业务键映射

## 3. E2E 与验证

- [x] 3.1 新增或扩展 E2E：英文环境打开内置参数编辑弹窗，断言 name/remark 为英文且无中文 seed
- [x] 3.2 运行相关 `go test`、`openspec validate config-edit-localized-metadata --strict`
