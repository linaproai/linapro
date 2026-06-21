## ADDED Requirements

### Requirement: builtin 源码插件启动升级必须复用安全升级治理

系统 SHALL 将`distribution=builtin`源码插件的启动升级视为运行时源码插件升级治理的启动期受控入口。该入口 MUST 只允许发现版本高于有效版本，必须复用依赖校验、反向依赖保护、生命周期回调、`phase=upgrade`迁移账本、治理资源同步、失败诊断和缓存发布规则。发现版本低于有效版本时，启动 MUST 失败并保留可诊断异常。

#### Scenario: builtin 启动升级成功

- **WHEN** `builtin`源码插件有效版本为`v0.1.0`
- **AND** 启动扫描发现版本为`v0.2.0`
- **THEN** 启动升级执行现有源码插件升级编排
- **AND** 升级成功后`sys_plugin.version`和有效 release 指向`v0.2.0`
- **AND** 迁移账本记录`phase=upgrade`

#### Scenario: builtin 发现版本低于有效版本

- **WHEN** `builtin`源码插件有效版本为`v0.2.0`
- **AND** 启动扫描发现版本为`v0.1.0`
- **THEN** 启动收敛失败
- **AND** 系统不得自动降级有效 release

#### Scenario: builtin 启动升级依赖不满足

- **WHEN** `builtin`源码插件发现新版本依赖当前环境不满足
- **THEN** 启动升级失败
- **AND** 有效版本保持升级前版本
- **AND** 错误包含目标插件和不满足的依赖信息

### Requirement: 普通管理入口不得手动治理 builtin 插件生命周期

系统 SHALL 拒绝普通插件管理入口对`distribution=builtin`插件执行安装、启用、禁用、卸载、手动升级和租户供应策略更新。拒绝 MUST 使用稳定业务错误码，并不得修改插件治理数据或触发缓存刷新。该规则不影响启动期`builtin`收敛入口。

#### Scenario: 禁用 builtin 插件被拒绝

- **WHEN** 管理员通过普通插件管理 API 禁用`distribution=builtin`插件
- **THEN** 系统返回`plugin.builtin.management.action.denied`业务错误
- **AND** 插件状态保持不变

#### Scenario: 卸载 builtin 插件被拒绝

- **WHEN** 管理员通过普通插件管理 API 卸载`distribution=builtin`插件
- **THEN** 系统返回`plugin.builtin.management.action.denied`业务错误
- **AND** 系统不得执行卸载 SQL、资源清理或缓存发布

#### Scenario: 手动升级 builtin 插件被拒绝

- **WHEN** 管理员通过普通插件管理 API 手动升级`distribution=builtin`插件
- **THEN** 系统返回`plugin.builtin.management.action.denied`业务错误
- **AND** 系统不得切换有效 release

#### Scenario: 启动期 builtin 收敛不受普通管理 guard 阻断

- **WHEN** 启动引导需要安装、启用或升级`distribution=builtin`插件
- **THEN** 系统通过启动期内部收敛入口执行
- **AND** 普通管理入口的`builtin`拒绝 guard 不得阻断该启动流程
