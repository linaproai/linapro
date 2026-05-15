## MODIFIED Requirements

### Requirement: 媒体策略作为共享策略来源
媒体策略配置 SHALL 作为全平台共享的策略来源，供 media 插件自身和 water 插件按业务租户、设备与全局优先级读取；LinaPro 中旧 hotgo `hg_strategy`、`hg_strategy_tenant`、`hg_strategy_device`、`hg_strategy_device_tenant` 的语义 MUST 对应到 `media_strategy`、`media_strategy_tenant`、`media_strategy_device`、`media_strategy_device_tenant`。

#### Scenario: water 插件读取媒体策略
- **WHEN** water 插件需要为截图任务解析水印策略
- **THEN** 它从 `media_*` 表读取启用策略，而不是创建或读取 `hg_*` 表

#### Scenario: 配置仍为平台共享
- **WHEN** 宿主处于多租户环境
- **THEN** media 策略配置不按 `host_tenant_id` 隔离，water 插件入参中的 `tenant` 仅表示媒体业务租户 ID
