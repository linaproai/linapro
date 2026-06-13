## ADDED Requirements

### Requirement: 插件生命周期编排下沉后必须保持治理语义

系统 SHALL 在将插件生命周期编排迁入 lifecycle 子组件后保持现有安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子的治理语义。平台上下文守卫、依赖检查、反向依赖阻断、host service authorization、SQL migration、资源引用同步、菜单权限同步、hook 分发、runtime state 同步和缓存失效不得因迁移遗漏或改变顺序。

#### Scenario: 插件安装语义保持

- **WHEN** 管理员安装 source 或 dynamic 插件
- **THEN** 系统仍执行依赖检查、manifest 校验、host service authorization、SQL migration、registry/release 写入、资源和菜单同步、hook 分发和缓存失效
- **AND** 任一步骤失败时保持迁移前的回滚或失败返回语义

#### Scenario: 插件卸载语义保持

- **WHEN** 管理员卸载插件
- **THEN** 系统仍在副作用前检查反向依赖和 lifecycle veto
- **AND** 保持 runtime 停止、资源引用清理、菜单权限清理、uninstall SQL、storage cleanup 和缓存失效顺序
- **AND** force 与 purge storage 语义不因 lifecycle 子组件迁移改变

#### Scenario: 启动自动启用语义保持

- **WHEN** 宿主启动并处理`plugin.autoEnable`
- **THEN** 系统仍只对配置目标插件执行自动安装/启用
- **AND** 保持 startup auto-enable 的依赖检查、mock-data 策略、租户供应策略和启动统计语义

### Requirement: 插件生命周期业务控制参数必须显式传递

系统 SHALL 将生命周期普通业务控制参数作为方法参数、options 或稳定输入结构显式传递。安装 mock data、startup auto-enable 标记、依赖检查结果和类似控制语义 MUST NOT 通过普通请求 context key 隐式改变生命周期行为。仅用于一次启动编排的只读大快照 MAY 继续通过 context 传递，但不得改变生命周期业务语义。

#### Scenario: 安装 mock data

- **WHEN** 调用方请求安装插件并选择加载 mock-data
- **THEN** 该选择通过 install options 或等价显式输入传入 lifecycle
- **AND** lifecycle 不通过 context key 判断是否加载 mock-data

#### Scenario: 启动自动启用

- **WHEN** startup bootstrap 触发插件自动安装或启用
- **THEN** startup 语义通过显式 options 或内部启动入口传入 lifecycle
- **AND** 普通 HTTP 安装请求无法通过伪造 context key 获得 startup-only 行为

### Requirement: 租户生命周期钩子必须通过 lifecycle 子组件编排

系统 SHALL 由 lifecycle 子组件编排租户删除、租户插件禁用和新租户供应相关的插件 lifecycle precondition 与 notification。根门面和 tenant capability adapter MUST 只依赖窄接口，不得复制租户生命周期扫描和 veto 汇总逻辑。

#### Scenario: 租户删除前检查插件 veto

- **WHEN** 租户能力在删除租户前请求插件 precondition
- **THEN** lifecycle 子组件扫描需要参与租户删除 veto 的已启用插件
- **AND** 使用统一 veto 汇总返回结构化错误
- **AND** 根门面不复制该扫描逻辑

#### Scenario: 新租户供应插件

- **WHEN** 新租户创建后需要按平台策略供应插件
- **THEN** lifecycle 子组件读取 auto-enable for new tenants 策略并执行供应编排
- **AND** 供应完成后通过统一插件变化发布入口失效派生缓存
