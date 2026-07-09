## Requirements

### Requirement: 插件清单处理必须与治理副作用分离

系统 SHALL 将插件清单扫描、解析、校验和访问视为清单事实源读取能力。清单处理路径不得隐藏写入治理表或触发菜单同步、权限同步、资源引用同步、hook 分发、运行时节点状态同步或缓存失效。需要治理写入或副作用时，调用方 MUST 在显式编排入口中按顺序调用清单、存储和副作用能力。

### Requirement: 插件治理写入后的副作用调用点必须可追踪

系统 SHALL 在插件显式同步、安装、卸载、启用、禁用、升级、动态包上传和租户供应策略更新等治理路径中保留可追踪的副作用调用点。每个菜单同步、资源引用同步、hook 分发、运行时节点状态同步和缓存失效操作 MUST 由当前编排入口显式触发，且错误处理语义必须与治理写入顺序一致。

### Requirement: 插件清单读取必须复用有界读模型缓存

系统 SHALL 为插件清单事实源提供内部读模型缓存。源码插件 manifest、动态插件 desired manifest、release manifest 与 release YAML 快照 MUST 按各自权威数据源和不可变性分开缓存。缓存读取路径 MUST 不写入插件治理表，不触发副作用。

### Requirement: 插件清单缓存必须支持按插件显式失效

系统 SHALL 在插件同步、动态包上传、安装、卸载、启用、禁用、升级、active release 切换或源码插件同步成功后，按 `pluginID` 或 artifact 路径失效对应清单缓存。全局失效只允许用于全局配置或无法确定影响范围的治理事件。

### Requirement: 插件生命周期编排下沉后必须保持治理语义

系统 SHALL 在将插件生命周期编排迁入 lifecycle 子组件后保持现有安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子的治理语义。平台上下文守卫、依赖检查、反向依赖阻断、host service authorization、SQL migration、资源引用同步、菜单权限同步、hook 分发、runtime state 同步和缓存失效不得因迁移遗漏或改变顺序。

### Requirement: 插件生命周期业务控制参数必须显式传递

系统 SHALL 将生命周期普通业务控制参数作为方法参数、options 或稳定输入结构显式传递。安装 mock data、startup auto-enable 标记、依赖检查结果和类似控制语义 MUST NOT 通过普通请求 context key 隐式改变生命周期行为。

### Requirement: 租户生命周期钩子必须通过 lifecycle 子组件编排

系统 SHALL 由 lifecycle 子组件编排租户删除、租户插件禁用和新租户供应相关的插件 lifecycle precondition 与 notification。根门面和 tenant capability adapter MUST 只依赖窄接口，不得复制租户生命周期扫描和 veto 汇总逻辑。
