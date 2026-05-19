## ADDED Requirements

### Requirement: 插件运行时变更必须发布 Redis coordination event
系统 SHALL 在集群模式下为插件安装、启用、禁用、卸载、升级、active release 切换、动态插件 artifact 变化发布 `plugin-runtime` Redis revision 和 event。

#### Scenario: 动态插件启用后其他节点刷新
- **WHEN** 主节点启用动态插件 P
- **THEN** 系统发布 `plugin-runtime` Redis revision
- **AND** 其他节点收到 event 后刷新 enabled snapshot
- **AND** 其他节点可路由到插件 P 的 active release

#### Scenario: 动态插件禁用后其他节点隐藏
- **WHEN** 主节点禁用动态插件 P
- **THEN** 系统发布 `plugin-runtime` Redis revision
- **AND** 其他节点失效 frontend bundle、runtime i18n 和 Wasm 派生缓存
- **AND** 后续访问插件 P 路由返回不可用或 404

### Requirement: 插件 runtime freshness 不可确认时必须 conservative-hide
系统 SHALL 在无法确认 `plugin-runtime` revision freshness 且超过最大陈旧窗口时采用 conservative-hide 策略。系统不得暴露可能已禁用、卸载或权限变化的插件能力。

#### Scenario: Redis plugin-runtime revision 不可读
- **WHEN** 请求需要访问动态插件运行时能力
- **AND** Redis `plugin-runtime` revision 不可读
- **AND** 本地插件运行时缓存超过最大陈旧窗口
- **THEN** 系统隐藏或拒绝该插件能力
- **AND** 不使用陈旧缓存继续放行

### Requirement: 动态插件 reconciler 必须由 Redis revision 唤醒
系统 SHALL 在集群模式下使用 Redis revision/event 唤醒动态插件 reconciler。安全扫描或低频 sweep MAY 保留作为兜底。

#### Scenario: active release 变化唤醒 reconciler
- **WHEN** 动态插件 active release 记录变化
- **THEN** 系统发布 reconciler scope 的 `plugin-runtime` revision
- **AND** 需要收敛的节点在观察到 revision 前进后执行收敛

#### Scenario: 事件错过后安全 sweep 兜底
- **WHEN** 节点错过 reconciler event
- **THEN** 节点通过 revision check 或低频 safety sweep 发现需要收敛
- **AND** 最终运行时状态与权威 release 记录一致

### Requirement: 插件派生缓存失效必须按 scope 精细化
系统 SHALL 在插件运行时变更时按插件 ID、sector、locale 或 global scope 精细失效 frontend bundle、runtime i18n 和 Wasm 缓存。普通路径不得无理由清空所有插件所有派生缓存。

#### Scenario: 单插件 frontend bundle 失效
- **WHEN** 动态插件 P 上传新 frontend bundle
- **THEN** 系统仅失效插件 P 相关 frontend bundle cache
- **AND** 其他插件 bundle cache 保持可用

