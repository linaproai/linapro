## ADDED Requirements

### Requirement: 运行时翻译包缓存保持全局交付资源维度
运行时翻译包缓存 SHALL 继续以 `locale`、`sector` 与插件标识作为缓存维度,不得在本次多租户迭代中引入无实际数据来源的 `tenant_id` 分桶。

本次多租户迭代不落地租户级 i18n override,运行时翻译包来源仍为宿主 `manifest/i18n`、源码插件 `manifest/i18n` 与动态插件 i18n 资产。字典与配置的租户覆盖 SHALL 由字典缓存、配置缓存和对应业务 API 的租户 fallback/override 语义保证隔离。

#### Scenario: 租户请求共享同一运行时翻译包
- **WHEN** 租户 A 与租户 B 请求相同 `locale` 的运行时翻译包
- **AND** 当前系统未启用租户级 i18n override 数据源
- **THEN** 两个请求命中相同的全局交付资源缓存
- **AND** 不因租户不同生成重复 runtime bundle

#### Scenario: 字典覆盖不污染运行时翻译包缓存
- **WHEN** 租户 A 修改允许覆盖的字典数据
- **THEN** 失效字典缓存中租户 A 对应的字典维度
- **AND** 不失效运行时翻译包缓存
- **AND** 租户 B 的字典缓存不受影响

### Requirement: 运行时翻译包失效必须显式限定资源作用域
运行时翻译包缓存失效 SHALL 使用显式 `InvalidateScope`,按 `locale`、`sector` 与插件标识限定范围。插件生命周期失效 SHALL 限定到插件 i18n sector 与插件标识;数据库 i18n 资源导入 SHALL 限定到受影响语言和数据库 sector。

#### Scenario: 插件 i18n 资产变更失效
- **WHEN** 动态插件 `notice` 的 i18n 资产随插件版本启用而变化
- **THEN** 仅失效 `sector=dynamic-plugin` 且 `plugin_id=notice` 的运行时翻译包缓存
- **AND** 不清空其他插件或宿主 sector 的缓存

#### Scenario: 数据库 i18n 资源导入失效
- **WHEN** 平台管理员导入 `zh-CN` 数据库 i18n 资源
- **THEN** 仅失效 `locale=zh-CN` 且 `sector=database` 的运行时翻译包缓存
- **AND** 不影响其他语言缓存

### Requirement: 字典与配置租户覆盖缓存承担租户维度
字典缓存与配置缓存 SHALL 携带 `tenant_id` 或等价租户作用域;租户写入覆盖时仅失效对应租户维度,平台默认值变更时按 fallback 影响范围级联失效对应业务缓存。

#### Scenario: 租户覆盖触发本租户业务缓存失效
- **WHEN** 租户 A 修改 `sys_config[key=upload.max_size]`
- **THEN** 仅 `(tenant_id=A, key=upload.max_size)` 配置缓存失效
- **AND** 运行时翻译包缓存不受影响
