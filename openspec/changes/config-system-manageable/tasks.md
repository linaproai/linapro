## 1. 数据模型与管理面

- [x] 1.1 `005` 建表增加 `system_manageable`
- [x] 1.2 管理面 List/Export/Get/Create/Update/Delete/Import 过滤与锁定
- [x] 1.3 错误码与 i18n

## 2. 能力 API 与插件

- [x] 2.1 `SetValue(ctx, key, value, options)` 与 `BatchSetValue(ctx, items, options)`（`SystemManageable` 用 `gconv.PtrBool`）
- [x] 2.2 宿主 adapter / domainhostcall 贯通；`SetValue` 委托 `BatchSetValue`
- [x] 2.3 插件 settings 多字段保存改为 `BatchSetValue`；fake/stub/README

## 3. 验证

- [x] 3.1 宿主与插件相关 Go 测试
- [x] 3.2 `make lint dir=apps/lina-core plugins=0`
