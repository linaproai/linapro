## Why

后端 Go 代码中存在多处 `panic` 调用，其中一部分属于启动期和插件注册期 fail-fast，符合项目规范；但也有运行期请求、导入导出、动态插件输入和运行时配置读取链路通过 `panic` 处理普通错误，容易绕过统一错误返回、日志与接口响应治理。

本变更需要明确 `panic` 的允许边界，并将非必要 `panic` 收敛为显式 `error` 返回或局部受控处理，避免普通业务路径因为可处理错误触发进程级异常。

## What Changes

- 明确后端生产代码中 `panic` 仅允许用于启动期、初始化期、不可回滚关键链路、`Must*` 语义构造函数和确需重新抛出的未知异常。
- 将 Excel 单元格坐标转换、通用资源关闭、运行时配置读取、动态插件 hostServices 规范化等运行期路径中的非必要 `panic` 改为显式错误返回；仅对无法返回错误且不影响主链路的辅助清理场景保留受控日志。
- 保留源码插件注册契约、插件扩展点声明、DB driver 注册等启动/注册期 fail-fast 行为，并补充测试覆盖避免误改。
- 增加面向后端 panic 治理的单元测试和静态检查脚本，防止普通业务路径再次引入非必要 `panic`。
- 本次变更不新增对外 API，不涉及数据库结构，不涉及前端运行时文案或 i18n 资源变更。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `backend-conformance`: 增加后端生产代码中 `panic` 使用边界与运行期错误显式返回要求。

## Impact

- 影响代码：`apps/lina-core/internal/service/config`、`apps/lina-core/internal/service/{user,dict,sysconfig}`、`apps/lina-core/pkg/{excelutil,closeutil,pluginbridge}`、相关动态插件与监控插件导出逻辑。
- 影响测试：新增或调整 Go 单元测试，覆盖 Excel 辅助函数、运行时配置异常值、动态插件 hostServices 校验和允许的启动期 fail-fast。
- 影响治理：新增静态检查脚本或测试入口，对后端生产代码中的新增 `panic` 做 allowlist 约束。
- i18n 影响：不新增、修改或删除用户可见文案；错误信息沿用现有后端错误返回路径，不需要同步维护前端语言包、manifest i18n 或 apidoc i18n 资源。
