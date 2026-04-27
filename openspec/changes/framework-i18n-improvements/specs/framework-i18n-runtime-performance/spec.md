## ADDED Requirements

### Requirement: 翻译查找热路径必须避免对运行时消息包的整体克隆
宿主系统 SHALL 让 `Translate`、`TranslateSourceText`、`TranslateOrKey`、`TranslateWithDefaultLocale` 等返回单值的翻译查找方法在缓存命中时直接对内部消息包持读锁查值,不得对运行时消息包整体进行克隆或拷贝。仅当方法语义上需要把消息集合返回给调用方时(例如运行时翻译包接口、消息导出接口),系统 MAY 在返回前克隆一次。

#### Scenario: 单 key 翻译查找命中缓存时不进行整张消息包克隆
- **WHEN** 业务模块调用 `Translate(ctx, key, fallback)` 且当前语言运行时消息缓存已存在
- **THEN** 系统仅对内部消息包持读锁后查值,直接返回查到的字符串
- **AND** 不进行 `cloneFlatMessageMap` 或等价的整张 `map[string]string` 拷贝
- **AND** 调用方仍能拿到与此前一致的语义结果

#### Scenario: 翻译包对外返回时仍保持克隆语义
- **WHEN** 控制器调用 `BuildRuntimeMessages` 或 `ExportMessages` 把消息集合返回给前端或导出
- **THEN** 系统在交出消息集合前克隆一次,确保调用方可安全独立持有
- **AND** 这次克隆不会污染或改写内部缓存

### Requirement: 运行时翻译缓存必须按语言与扇区分层失效
宿主系统 SHALL 把运行时翻译消息缓存组织为"按语言 × 扇区(host / source-plugin / dynamic-plugin)"分层结构,并按扇区维度提供精细失效能力。任何业务事件触发的失效 MUST 仅清除受影响的语言或扇区,不得"一锅清"所有语言、所有扇区的缓存。宿主核心 i18n 不得引入数据库覆写扇区或运行时业务内容缓存;翻译内容以开发期 JSON/YAML 资源为唯一事实源。

#### Scenario: 宿主资源失效仅清除目标语言宿主扇区
- **WHEN** 维护工具或测试流程触发 `en-US` 宿主资源缓存失效
- **THEN** 系统只清除 `en-US` 语言的宿主扇区缓存与合并视图
- **AND** `zh-CN` 与其他启用语言的缓存保持原值
- **AND** `en-US` 中源码插件、动态插件扇区无需重新加载

#### Scenario: 动态插件启停仅清除该插件相关扇区
- **WHEN** 某个动态插件被启用、停用或升级
- **THEN** 系统仅清除涉及该插件的动态插件扇区缓存与合并视图
- **AND** 宿主与未受影响插件的翻译数据继续命中缓存
- **AND** 重新合并时只对该插件 ID 的资源做加载或移除

### Requirement: 运行时翻译包接口必须支持 ETag 协商
宿主系统 SHALL 在 `/i18n/runtime/messages` 接口响应中输出 `ETag` 头,值由当前语言与运行时翻译包版本(`bundleVersion`)派生且在版本变化时必然不同。系统 MUST 接收请求中的 `If-None-Match` 头,若值与当前响应 ETag 一致则返回 `304 Not Modified` 且不携带消息体。每次扇区缓存失效 MUST 触发 `bundleVersion` 自增,确保同一语言下的不同 bundle 内容拥有不同 ETag。

#### Scenario: 同一 bundle 在二次请求时返回 304
- **WHEN** 前端首次以 `Accept-Language: en-US` 请求运行时翻译包并保存返回的 `ETag`
- **AND** 后端在两次请求之间未发生任何缓存失效
- **AND** 前端在第二次请求中携带 `If-None-Match` 等于上次的 `ETag`
- **THEN** 后端返回 `304 Not Modified` 且不携带消息体

#### Scenario: 翻译资源变化后 ETag 一定变化
- **WHEN** 任意扇区(host / source-plugin / dynamic-plugin)发生缓存失效
- **THEN** `bundleVersion` 自增
- **AND** 同语言下次请求返回的 `ETag` 与之前不同
- **AND** 携带旧 `If-None-Match` 的请求返回 `200` 与最新消息体

### Requirement: 默认管理工作台必须按 ETag 持久化运行时翻译并接入鉴权链
默认管理工作台 SHALL 通过统一的 `requestClient` 调用运行时翻译包接口,使其参与鉴权、错误处理和降级链。前端 SHALL 把每次成功响应的 `{locale, etag, messages, savedAt}` 持久化到 `localStorage`,并在下次进入或语言切换时优先用持久化数据快速渲染,然后在后台带 `If-None-Match` 协商。持久化数据 MUST 设置不长于 7 天的 TTL,超过 TTL 时强制重新拉取。

#### Scenario: 二次进入页面时零网络切语言
- **WHEN** 用户已在某语言下成功加载过运行时翻译包并写入持久化缓存
- **AND** 用户在 7 天内重新打开页面或切换到该语言
- **THEN** 前端直接用持久化数据完成 `vue-i18n` 注入
- **AND** 后台异步带 `If-None-Match` 发起协商,命中 `304` 时不更新内存与持久化数据

#### Scenario: 持久化数据超过 TTL 时强制刷新
- **WHEN** 持久化条目的 `savedAt` 距离当前时间超过 7 天
- **THEN** 前端忽略持久化数据,发起带空 `If-None-Match` 或不带该头的请求
- **AND** 拉取成功后更新内存与持久化数据,刷新 `savedAt`

#### Scenario: 运行时翻译失败时降级到持久化兜底
- **WHEN** 运行时翻译包接口在网络异常、超时或服务端 5xx 时失败
- **AND** 持久化缓存中存在该语言的有效条目
- **THEN** 前端使用持久化条目完成渲染,不阻塞页面
- **AND** 前端通过统一降级通知机制告知用户翻译可能存在版本偏差
