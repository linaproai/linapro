## ADDED Requirements

### Requirement: runtime revision controller 必须属于缓存协调边界

系统 SHALL 将运行时缓存 revision controller 作为缓存协调组件能力维护，而不是作为插件领域包的非 internal 子包暴露。插件 runtime、插件管理读模型、runtime reconciler 和 i18n runtime bundle 等消费方 MUST 从缓存协调边界导入 revision controller，并按各自 domain/scope 创建实例。

#### Scenario: plugin 和 i18n 使用 revision controller

- **WHEN** plugin runtime 缓存和 i18n runtime bundle 缓存需要 observed revision
- **THEN** 二者从缓存协调组件导入 revision controller
- **AND** 不导入`internal/service/plugin/runtimecache`
- **AND** 各自实例化独立 domain/scope controller

#### Scenario: revision controller tenant scope

- **WHEN** 调用方需要设置 tenant scope
- **THEN** controller API 使用返回副本的`WithTenantScope`或语义明确的构造期 setter
- **AND** 不得在共享 controller 已被多个调用方使用后通过误导性 fluent API 原地修改作用域

#### Scenario: 审查插件 runtimecache 旧路径

- **WHEN** 生产或测试代码重新导入`internal/service/plugin/runtimecache`
- **THEN** 静态治理测试失败
- **AND** 调用方必须改用缓存协调边界下的 revision controller
