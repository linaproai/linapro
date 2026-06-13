## ADDED Requirements

### Requirement: 插件生命周期编排必须归属 lifecycle 子组件

系统 SHALL 将插件安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子编排归属到`internal/service/plugin/internal/lifecycle`。`internal/service/plugin`根门面 MUST 只保留公共契约、平台治理守卫、输入轻量校验、必要锁/缓存入口协调和委托，不得直接承载迁移后的生命周期长状态机。

#### Scenario: 根门面执行生命周期写操作

- **WHEN** 调用方通过插件根服务执行 Install、Uninstall、Enable、Disable 或 UpdateStatus
- **THEN** 根门面先执行平台上下文和入口治理守卫
- **AND** 根门面将业务编排委托给 lifecycle 子组件
- **AND** 根门面不得直接访问插件治理 DAO 或执行 SQL migration 细节

#### Scenario: lifecycle 编排执行状态机

- **WHEN** lifecycle 子组件处理安装、卸载或状态变更
- **THEN** 它通过构造函数注入的 catalog、store、runtime、integration、migration、dependency、i18n 和 cache publisher 窄契约完成编排
- **AND** 不通过 package-level service locator、构造后 setter 或反向持有插件根门面完成调用

### Requirement: 插件 SQL migration executor 必须独立于 lifecycle 编排

系统 SHALL 将插件生命周期 SQL 文件执行、mock data 执行、uninstall SQL 执行和 migration ledger 维护归属到独立`internal/service/plugin/internal/migration`组件。lifecycle 编排 MUST 只通过窄接口调用 migration executor，不得把 SQL 文件执行细节和生命周期状态机混在同一组件主流程中。

#### Scenario: 安装流程执行 SQL

- **WHEN** lifecycle 编排需要执行插件 install SQL 或 mock-data SQL
- **THEN** lifecycle 调用 migration 组件的窄接口
- **AND** migration 组件保持事务、方言转译和账本一致性语义
- **AND** lifecycle 编排函数只表达安装步骤和错误处理语义

#### Scenario: 动态插件卸载执行 SQL

- **WHEN** lifecycle 编排需要执行动态插件 uninstall SQL
- **THEN** migration 组件负责 SQL asset 读取、方言转译、执行和 ledger 处理
- **AND** lifecycle 负责卸载状态机、资源清理和回滚/收尾调用顺序

### Requirement: 插件列表投影必须由单一投影构建入口维护

系统 SHALL 为插件管理列表、管理摘要、详情和依赖快照提供单一投影构建入口。该入口 MUST 明确输入 mode、当前页或目标插件范围，并统一处理 manifest snapshot、store governance projection、runtime item、host service authorization、dependency summary、租户供应策略和 i18n 展示字段的批量装配。

#### Scenario: 构建管理摘要列表

- **WHEN** 插件管理列表需要摘要投影
- **THEN** 系统通过统一投影构建入口选择 summary mode
- **AND** 复用同一批 manifest、store、runtime 和 dependency 快照
- **AND** 不为摘要列表复制一条独立 manifest 扫描和 runtime 合并流水线

#### Scenario: 构建插件详情

- **WHEN** 插件管理需要单个插件详情
- **THEN** 系统通过统一投影构建入口选择 detail mode 并限制目标插件范围
- **AND** 返回字段保持与列表和同步构建路径一致
- **AND** 不循环调用单项详情接口装配列表、摘要或依赖快照

### Requirement: 插件服务根门面不得直接访问治理 DAO

系统 SHALL 阻断迁移完成后的`internal/service/plugin`根门面生产代码直接访问`internal/dao`、`internal/model/do`或`internal/model/entity`完成插件治理读写。插件治理持久化 MUST 通过`store`，生命周期 SQL MUST 通过`migration`，列表投影 MUST 通过投影构建入口。

#### Scenario: 根门面新增治理读写

- **WHEN** 开发者在插件根包新增或修改生产 Go 代码
- **THEN** 静态治理测试检查根门面生产代码不得导入`lina-core/internal/dao`、`lina-core/internal/model/do`或`lina-core/internal/model/entity`
- **AND** 需要治理数据时必须改用 store、lifecycle、migration 或投影组件的窄契约

#### Scenario: 测试和启动装配例外

- **WHEN** 测试 fixture、治理扫描或启动装配需要接触多个组件
- **THEN** 该访问必须保留在测试、装配或验证边界
- **AND** 不得进入普通业务运行路径

### Requirement: 插件生命周期重复 helper 必须收敛并受长度治理

系统 SHALL 收敛 lifecycle veto 汇总、dynamic decision 汇总、卸载收尾、动态插件启用资格判断和 decision/err 处理等同构逻辑。本变更迁移或重写后的生命周期编排函数和投影函数 MUST 使用命名明确的窄函数表达步骤，单个迁移后业务函数 SHOULD 不超过 60 行；超过时必须在设计或任务记录中说明不可拆分原因。

#### Scenario: 汇总生命周期 veto

- **WHEN** source plugin lifecycle decision 或 dynamic plugin lifecycle decision 需要返回阻断原因
- **THEN** 系统使用同一套汇总 helper 或稳定等价抽象
- **AND** 不再维护两份逐行同构的计数和本地化逻辑

#### Scenario: 拆分 UpdateStatus 四象限

- **WHEN** lifecycle 编排处理 source/dynamic 与 enable/disable 组合
- **THEN** 每个组合分支进入命名明确的窄函数
- **AND** 主流程只表达分派、事务边界和错误语义
