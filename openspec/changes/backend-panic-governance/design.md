## Context

当前后端生产代码中的 `panic` 主要集中在四类位置：启动期命令树和驱动注册、源码插件注册契约、运行期配置解析、Excel/动态插件辅助函数。前两类属于启动或注册时的不可恢复错误，符合项目规范；后两类位于普通请求、文件导入导出、动态插件装载和运行时参数读取路径，应该通过 `error` 返回、统一业务错误响应或受控降级处理。

本项目是全新项目，无需兼容旧 API 或历史行为；因此可以直接调整内部函数签名、调用链和测试，而不保留会继续诱导误用的旧 helper 形态。

## Goals / Non-Goals

**Goals:**

- 明确生产代码 `panic` allowlist：启动期、初始化期、不可回滚关键链路、`Must*` 语义构造函数、未知 panic 重新抛出。
- 将普通业务路径中的可处理错误改成显式 `error` 返回，确保控制器、服务和插件装载流程能按现有错误通道返回。
- 对运行时配置异常值采用“写入严格校验、读取显式报错”的策略，避免错误数据在高频业务接口中被静默吞掉。
- 为新增或保留的 `panic` 增加静态检查，要求生产代码中的新增 panic 必须落在 allowlist。
- 记录 i18n 判断：本变更不新增用户界面文案、不修改 API DTO 文档源文本、不调整 manifest/apidoc i18n 资源。

**Non-Goals:**

- 不移除启动期和源码插件注册契约中的 fail-fast 行为。
- 不改变数据库结构、不新增 SQL 初始化文件。
- 不新增前端页面交互，不创建 E2E 用例；本变更以 Go 单元测试和静态检查覆盖为主。
- 不调整 GoFrame 框架自身可能产生的 panic，只在宿主 middleware 已识别路径中做受控转换。

## Decisions

1. 保留 `Must*` 与启动注册期 panic，运行期路径改为 `error`。

   备选方案是全仓库禁止 `panic`，但这会削弱启动期配置、DB driver、源码插件契约错误的可诊断性。最终选择 allowlist，因为它与项目规范一致，也能让不可恢复错误快速暴露。

2. Excel 坐标辅助函数不再返回裸 `string`。

   当前 `cellName(col,row) string` 在 `CoordinatesToCellName` 失败时只能 panic。实现时优先直接调用 `excelutil.SetCellValue(file, sheet, col, row, value)`；确实需要 A1 坐标字符串的位置改为 `cellName(...)(string,error)` 并逐层返回错误。

3. 动态插件 hostServices 规范化拆分为 error 版本和 Must 版本。

   `NormalizeHostServiceSpecs` 当前在校验失败时 panic，但它被动态插件产物、catalog release、授权流程等动态输入调用。实现时新增 `NormalizeHostServiceSpecsE` 返回 `([]*HostServiceSpec, error)`，动态输入路径使用 error 版本；如确有编译期常量路径需要，可保留 `MustNormalizeHostServiceSpecs`。

4. 运行时配置读取采用显式错误返回。

   写入 sys_config 的受保护参数仍由 `ValidateProtectedConfigValue` 严格校验；读取快照时如果手工 SQL、外部写入或缓存污染造成异常值，业务 getter 不再 panic，也不通过 warning 日志后继续返回默认值，而是沿调用链返回 `error`。这样可以让控制器、中间件和业务服务按统一错误通道暴露配置异常，避免请求看似成功但实际使用了错误配置。

5. 静态检查采用 allowlist 文件或脚本。

   仅依靠代码审查容易复发。新增检查应扫描 `apps/lina-core` 与 `apps/lina-plugins` 的生产 Go 文件，排除 `_test.go`，并要求每个 `panic(` 调用点匹配 allowlist 说明。这样保留少量必要 panic，同时阻止运行期 helper 再次引入。

## Risks / Trade-offs

- 运行时配置显式返回错误会让依赖该配置的请求失败 → 这是更符合规范的 fail-visible 行为，写入路径严格校验仍负责防止正常管理入口保存非法值。
- 调整 helper 签名会触碰多个导出流程 → 通过局部改造和 `go test` 覆盖相关 service/plugin 包，避免引入批量重构。
- allowlist 可能变成“登记即可通过”的形式 → allowlist 条目必须写明所属类别和理由，并由测试固定当前允许集合。
- 动态插件 hostServices 规范化改为 error 后调用链需要补齐错误上下文 → 在 artifact、catalog、authorization 路径分别包装错误，确保用户能定位具体插件或产物。
