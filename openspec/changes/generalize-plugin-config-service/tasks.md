## 1. 公共配置服务接口

- [x] 1.1 调整 `apps/lina-core/pkg/pluginservice/config` 的公开接口，提供 `Get`、`Exists`、`Scan`、基础类型读取和 `Duration` 读取能力。
- [x] 1.2 移除 `MonitorConfig` 类型别名和 `GetMonitor()` 插件业务专用方法，确保公共组件不再引用插件业务配置结构。
- [x] 1.3 为通用配置读取方法补充 Go 注释、错误处理和默认值语义，符合 GoFrame v2 与项目后端代码规范。

## 2. Monitor Server 插件迁移

- [x] 2.1 在 `monitor-server` 插件内部新增私有配置加载逻辑，维护监控配置结构、默认值、duration 解析和业务校验。
- [x] 2.2 将 `monitor-server` 定时采集任务注册与清理逻辑迁移到新的通用配置服务读取方式。
- [x] 2.3 全仓搜索并清理 `GetMonitor()`、`MonitorConfig` 等旧公共接口引用。

## 3. 测试与验证

- [x] 3.1 为 `pluginservice/config` 增加单元测试，覆盖任意 key 读取、缺失 key 默认值、结构体扫描、基础类型读取、duration 成功解析和失败返回。
- [x] 3.2 为 `monitor-server` 插件配置加载增加或更新单元测试，覆盖默认值、配置覆盖、非法 duration 和业务校验。
- [x] 3.3 运行受影响 Go 测试并修复失败项。

## 4. 治理检查

- [x] 4.1 确认本次变更不影响前端页面、菜单、路由、按钮、表单、提示文案、运行时 i18n、manifest i18n 或 apidoc i18n 资源，并在实现结论中记录。
- [x] 4.2 确认本次变更仅读取静态配置文件，不新增运行时可变缓存；若实现中新增缓存，补充权威数据源、一致性模型、失效触发点、跨实例同步机制和故障降级说明。
- [x] 4.3 完成实现后调用 `lina-review` 进行代码与规范审查。

## Feedback

- [x] **FB-1**: 复用 `monitor-server` 定时任务中的监控服务实例，避免每次 cron 执行重新构造 `monitorsvc.New()`
- [x] **FB-2**: 允许动态插件通过 `config` host service 读取完整静态配置
- [x] **FB-3**: 历史反馈：曾将动态插件 `config` host service 方法收敛为仅支持 `get`（已由 FB-5 调整）
- [x] **FB-4**: 历史反馈：保留动态插件 guest config SDK 的丰富读取方法，并允许 `config` host service 省略 methods 时默认授权 `get`（已由 FB-5 调整）
- [x] **FB-5**: 允许动态插件 `config` host service 直接调用全部只读配置方法，省略 `methods` 时默认授权全部只读方法
