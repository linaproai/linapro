# plugin-lock-service Specification

## Purpose
TBD - created by archiving change dynamic-plugin-host-service-extension. Update Purpose after archive.
## Requirements
### Requirement: 动态插件通过命名锁资源获取宿主锁能力

系统 SHALL 为动态插件提供受治理的锁服务，插件只能对宿主授权的命名锁资源执行获取、续租和释放操作。

#### Scenario: 插件获取授权锁资源

- **WHEN** 插件调用锁服务获取一个已授权的`host-lock`资源
- **THEN** 宿主按该锁资源的租约和超时策略执行加锁
- **AND** 宿主将逻辑锁名自动绑定到插件隔离的实际锁名
- **AND** 宿主返回锁票据或失败结果

#### Scenario: 插件释放或续租锁资源

- **WHEN** 插件调用锁服务释放或续租一个已持有的锁
- **THEN** 宿主校验锁票据和锁资源匹配关系
- **AND** 仅对当前插件持有的有效锁执行续租或释放

