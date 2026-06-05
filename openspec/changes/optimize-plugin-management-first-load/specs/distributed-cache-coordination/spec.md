## ADDED Requirements

### Requirement:插件管理读模型缓存必须复用 plugin-runtime 协调

系统 SHALL 将插件管理摘要列表读模型和详情读模型视为 `plugin-runtime`派生缓存。缓存 MUST 绑定插件运行时修订号、locale 和运行时翻译包版本，并复用既有单机本地 revision 或集群 Redis revision/event 完成失效；系统 MUST NOT 为插件管理读模型创建仅当前节点可见的独立缓存协调域。

#### Scenario:单节点模式本地失效插件管理读模型
- **WHEN** `cluster.enabled=false` 且插件安装、启用、禁用、卸载、升级、active release 切换或源码插件同步成功
- **THEN** 系统更新本地 `plugin-runtime` revision
- **AND** 当前进程内插件管理摘要列表缓存和对应插件详情缓存失效
- **AND** 下一次插件管理请求基于新的权威状态重建读模型

#### Scenario:集群模式通过 Redis event 失效插件管理读模型
- **WHEN** `cluster.enabled=true` 且某节点发布 `plugin-runtime` Redis revision/event
- **THEN** 其他节点观察到 revision 前进后失效本地插件管理摘要列表缓存和受影响插件详情缓存
- **AND** 后续插件管理请求不得继续返回旧 revision 下的插件安装、启用、版本或授权摘要状态

#### Scenario:语言资源变化区分缓存键
- **WHEN** 当前用户 locale 或 runtime bundle version 与已缓存插件管理读模型不同
- **THEN** 系统使用独立缓存键读取或构建摘要列表和详情读模型
- **AND** 系统不得把旧语言或旧运行时翻译包版本下的插件展示元数据返回给当前请求

#### Scenario:无法确认 freshness 时不返回过期治理状态
- **WHEN** 节点无法确认 `plugin-runtime` revision freshness
- **AND** 本地插件管理读模型超过域策略允许的陈旧窗口
- **THEN** 系统不得返回过期摘要列表或详情
- **AND** 系统按插件运行时故障策略 conservative-hide 或结构化错误处理该请求
