## ADDED Requirements

### Requirement: 插件升级编排必须归属 upgrade 子组件

系统 SHALL 将源码插件升级状态、动态插件升级 preview、升级 execute、失败诊断、release 提升和缓存发布编排归属到`internal/service/plugin/internal/upgrade`。`internal/service/plugin`根门面 MUST 只保留公开契约、平台治理守卫、必要锁入口协调、输入轻量校验和委托，不得继续承载迁移后的升级长流程。

#### Scenario: 根门面执行升级操作

- **WHEN** 调用方通过插件根服务执行 source 或 dynamic 插件升级
- **THEN** 根门面先执行平台上下文和治理守卫
- **AND** 根门面将升级业务编排委托给`upgrade`子组件
- **AND** 根门面不得直接导入`sourceupgrade`或`runtimeupgrade`包完成升级流程

#### Scenario: upgrade 子组件执行升级状态机

- **WHEN** `upgrade`子组件处理 source 或 dynamic 升级
- **THEN** 它通过构造函数注入的 catalog、store、lifecycle、runtime、integration、dependency、i18n、locker、cache publisher 和 topology 窄契约完成编排
- **AND** 不通过 package-level service locator、构造后 setter 或反向持有插件根门面完成调用

### Requirement: 插件服务不得保留 sourceupgrade 与 runtimeupgrade 平行包

系统 SHALL 在统一升级编排落地后删除`internal/sourceupgrade`和`internal/runtimeupgrade`平行包。源码插件升级和动态插件升级只允许通过`internal/upgrade`及其内部职责文件表达差异；静态治理测试 MUST 阻断生产代码重新导入旧平行包或重新创建旧目录。

#### Scenario: 生产代码重新导入旧升级包

- **WHEN** 插件服务生产 Go 文件导入`internal/sourceupgrade`或`internal/runtimeupgrade`
- **THEN** 静态治理测试失败
- **AND** 调用方必须改为依赖`internal/upgrade`的窄契约或同包内部 helper

#### Scenario: 旧升级目录重新出现

- **WHEN** `apps/lina-core/internal/service/plugin/internal/sourceupgrade`或`apps/lina-core/internal/service/plugin/internal/runtimeupgrade`目录重新出现
- **THEN** 静态治理测试失败
- **AND** 新增升级逻辑必须放入`internal/upgrade`

### Requirement: 插件升级测试必须覆盖统一编排边界

系统 SHALL 为统一升级编排提供职责明确的后端单元测试和静态边界测试。测试必须覆盖 source upgrade status、dynamic preview、source execute、dynamic execute、失败诊断、治理守卫单次执行、缓存发布和旧升级包清零。

#### Scenario: 验证 source 与 dynamic 共用失败诊断

- **WHEN** source 或 dynamic 插件升级在 SQL、callback、release switch 或缓存发布阶段失败
- **THEN** 测试验证失败诊断使用统一 phase 和 message key 语义
- **AND** 有效 release 与派生缓存权威来源保持正确

#### Scenario: 验证根门面不再承载升级长流程

- **WHEN** 开发者修改插件根门面升级文件
- **THEN** 静态边界测试检查根门面不导入旧升级包、不再通过公开 source upgrade 方法再入、不直接拼装 runtime upgrade preview 纯函数
