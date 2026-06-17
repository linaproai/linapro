## Requirements

### Requirement: 动态普通领域 host service 协议名必须与领域目录一致

系统 SHALL 要求动态插件普通领域 `hostServices.service` 协议名与已发布的动态领域目录名称保持一致。集合型领域 MUST 使用复数领域名。`i18n` 不属于动态插件可声明的普通领域 host service；动态插件多语言资源由宿主统一管理。

### Requirement: 动态插件普通领域 host service 必须覆盖源码插件普通领域能力

系统 SHALL 让动态插件通过 `hostServices` 获得已发布动态普通领域能力的覆盖。动态插件协议 MUST NOT 暴露 `AdminServices` 目录、数据库查询构造器、`DAO/DO/Entity`、HTTP 请求对象、宿主内部 service 或 `i18n` 运行时翻译服务。

### Requirement: 动态插件 i18n 资源必须由宿主管理

系统 SHALL 允许动态插件继续通过 `manifest/i18n` 交付多语言资源，但运行时资源发现、合并、缓存、失效和前端语言包分发 MUST 由宿主统一管理。动态插件后端 MUST NOT 通过 host service 读取 locale、翻译消息或检索 message key。

### Requirement: WASM host service runtime 必须由实例持有

系统 SHALL 让动态插件 WASM host service 分发器通过显式 runtime 实例读取运行期依赖。WASM host service 的 domain capability directory、插件配置 factory、宿主配置 service、manifest service factory 和其他共享宿主能力 MUST 来自启动期构造并注入的实例，不得通过包级 `Configure*` 函数、包级 `atomic.Pointer` 快照或包级默认实例作为生产调用事实源。

### Requirement: Host call 授权快照可以请求内复用但不得改变治理语义

系统 SHALL 允许 WASM host service handler 在同一次 guest 执行中复用已构建的 host service 授权快照。复用 MUST 仅降低快照装配成本，不得改变当前 active release 授权来源、service/method/resource 校验、数据权限、租户边界、审计字段或错误 envelope。

### Requirement: WASM host service 公共 helper 必须归属公共层

系统 SHALL 将跨领域复用的 WASM host service helper 归属到 `wasm` 公共 host service 文件或 `hostservicedispatch` 公共层。具体领域文件 MUST 只承载该领域 service/method 的 transport 适配，不得承载 `CapabilityContext` 构造或 registry 公共响应逻辑。
