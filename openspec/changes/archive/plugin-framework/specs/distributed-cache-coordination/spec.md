## Requirements

### Requirement: runtime revision controller 必须属于缓存协调边界

系统 SHALL 将运行时缓存 revision controller 作为缓存协调组件能力维护，而不是作为插件领域包的非 internal 子包暴露。插件 runtime、插件管理读模型、runtime reconciler 和 i18n runtime bundle 等消费方 MUST 从缓存协调边界导入 revision controller，并按各自 domain/scope 创建实例。

### Requirement: 插件生命周期缓存失效必须通过单一变化发布入口

系统 SHALL 为插件同步、动态包上传、安装、卸载、启用、禁用、源码升级、动态升级、租户供应策略更新和启动自动启用等插件治理变化提供单一插件变化发布入口。该入口 MUST 复用 `plugin-runtime` revision controller，统一失效 runtime 派生缓存、插件管理读模型、frontend bundle、i18n runtime bundle 和 WASM 相关派生状态。

### Requirement: 集群和单机缓存后端选择必须在拓扑感知构造边界完成

系统 SHALL 在 HTTP 启动期根据拓扑显式创建缓存后端。单机模式使用 SQL table provider；集群模式使用 coordination KV provider。生产路径 MUST NOT 依赖包级默认 provider 隐式选择后端。
