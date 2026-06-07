## ADDED Requirements

### Requirement: 插件服务根包必须保持 facade 职责清晰

系统 SHALL 将`apps/lina-core/internal/service/plugin`根包作为插件宿主服务的稳定 facade，根包主文件只保留公共契约、核心类型、构造、轻量校验和必要 wiring。复杂实现流程 SHOULD 放在同包职责明确的非主文件或既有`internal/<subcomponent>`中；当移动会造成循环依赖或只新增透传抽象时，系统 MAY 保持同包实现，但必须通过文件名和测试组织体现职责。

#### Scenario: 根包新增复杂实现

- **WHEN** 开发者为插件根服务新增生命周期、runtime upgrade、列表投影、host service 或缓存相关复杂流程
- **THEN** 实现文件名必须体现具体职责
- **AND** 不得把具体流程直接塞回`plugin.go`主文件

#### Scenario: 子组件移动会造成无意义转发

- **WHEN** 某段逻辑移动到`internal/<subcomponent>`只会产生新增接口和参数透传
- **THEN** 系统可以保持同包实现
- **AND** 通过窄函数、清晰命名和关联测试控制复杂度

### Requirement: 插件服务测试必须按被测职责组织

系统 SHALL 让插件服务单元测试文件与被测源码或明确主题关联。大测试文件 SHOULD 按 lifecycle、runtime upgrade、management list、startup auto-enable、tenant governance、host service 和测试 fixture 等职责拆分。共享 helper MUST 放在职责明确的测试支撑文件或`internal/testutil`，并由当前测试显式调用。

#### Scenario: 根包测试 helper 被多个测试文件复制

- **WHEN** 同一测试 helper 被多个根包测试文件重复实现
- **THEN** helper 应收敛到根包`*_test.go`支撑文件或`internal/testutil`
- **AND** 不得为了复用 helper 扩大生产代码导出面

#### Scenario: 大测试文件包含多个无关主题

- **WHEN** 一个`*_test.go`文件同时覆盖多个不同源码职责
- **THEN** 测试应按职责拆分到关联文件
- **AND** 每个测试仍必须自行构造依赖、数据和清理逻辑
