# go-unit-test-execution-efficiency Specification

## Purpose
TBD - created by archiving change optimize-go-unit-test-runtime. Update Purpose after archive.
## Requirements
### Requirement: Go 单元测试必须保留并发安全检测覆盖
Go 单元测试主路径 SHALL 保留 `-race` 对并发敏感代码的检测能力。降低总体耗时的实现 MUST 优先减少重复重型 fixture、真实链路执行和无测试包扫描成本，不得以无说明地移除主路径 race 覆盖作为默认优化手段。

#### Scenario: 默认 Go 单元测试仍启用 race
- **WHEN** CI 或开发者运行仓库默认 Go 单元测试入口
- **THEN** 执行计划必须明确显示 race 检测处于启用状态
- **AND** 插件运行时、缓存、协调、session、锁、认证、租户和权限相关包必须继续被 race 检测覆盖

#### Scenario: 调整 race 覆盖需要显式说明
- **WHEN** 某个测试入口、包集合或 smoke 入口不启用 race
- **THEN** 任务记录或命令输出必须说明该入口的用途、覆盖边界和不启用 race 的原因
- **AND** 不得用该入口替代默认 Go 单元测试质量门禁

### Requirement: 真实 dynamic Wasm 样例执行必须限定在 smoke 边界
Go 单元测试 SHALL 区分普通逻辑测试和真实 dynamic Wasm 样例链路测试。普通插件 runtime、catalog、integration 和 lifecycle 单测 MUST 优先使用 synthetic artifact、fake executor、轻量 host service 替身或测试辅助生成的 artifact；真实 bundled dynamic Wasm 样例执行 MUST 收敛为少量 smoke 覆盖。

#### Scenario: 普通 runtime 逻辑测试使用轻量 artifact
- **WHEN** 测试验证 dynamic plugin manifest 解析、路由分发、cron 声明、授权判断或错误映射
- **THEN** 测试必须使用 synthetic artifact、fake executor 或测试辅助生成的轻量 artifact
- **AND** 测试不得为了验证纯逻辑路径重复执行真实 bundled dynamic Wasm 样例

#### Scenario: 真实 bundled Wasm smoke 覆盖保留
- **WHEN** 测试需要证明真实 bundled dynamic Wasm 样例与宿主 bridge 能够闭环
- **THEN** 测试必须标识为 smoke 或集成型单元测试
- **AND** 同一测试包中真实样例 artifact 构建或加载应复用一次性 fixture，避免每个用例重复构建

### Requirement: 插件测试 fixture 必须复用且保持测试隔离
插件相关 Go 测试 SHALL 通过共享 helper 或包级 fixture 复用不可变基础资源，减少重复 artifact 写入、manifest 同步、插件安装启用、runtime cache 刷新和治理表清理。每个测试 MUST 仍保持自包含、顺序无关和数据隔离。

#### Scenario: 插件生命周期测试声明差异场景
- **WHEN** 测试需要安装、启用、禁用、卸载、上传、同步或升级插件
- **THEN** 测试应复用共享 fixture 准备基础服务、artifact 模板和清理辅助
- **AND** 测试只声明当前场景差异所需的 plugin id、release、tenant 或授权输入

#### Scenario: 共享 fixture 不污染测试状态
- **WHEN** 多个测试复用同一 fixture
- **THEN** 每个测试必须使用唯一业务标识或隔离作用域
- **AND** 每个测试必须通过 `t.Cleanup` 或等价机制清理自身创建的数据库行、文件和缓存状态

### Requirement: Go 测试入口必须避免完整执行无测试包
`linactl test.go` SHALL 在执行测试前生成测试计划，区分包含 `_test.go` 的包、仅需编译 smoke 的包和无需执行单测的包。默认单测入口 MUST 避免对大量无测试文件的包执行完整 `go test ./...` 主路径。

#### Scenario: 发现含测试文件的包
- **WHEN** `linactl test.go` 准备执行某个 Go module 的测试
- **THEN** 它必须发现该 module 内包含 `_test.go` 的包集合
- **AND** 它必须仅对这些包执行单元测试命令，除非调用方显式要求完整 `./...` 模式

#### Scenario: 无测试包仍可获得编译保障
- **WHEN** module 中存在没有 `_test.go` 但仍需要编译检查的生产包
- **THEN** 测试入口必须提供轻量编译 smoke 或在执行计划中明确说明由其他编译门禁覆盖
- **AND** 不得把无测试包伪装成已执行单元测试

### Requirement: Go 单元测试运行必须输出可审计耗时摘要
Go 单元测试入口 SHALL 输出足够的执行计划和耗时摘要，使开发者能够持续识别最慢 module、测试包数量、无测试包数量、race 状态和真实 smoke 边界。

#### Scenario: 测试启动报告执行计划
- **WHEN** 开发者或 CI 启动 `linactl test.go`
- **THEN** 输出必须包含模块数量、待测试包数量、无测试包数量、race 状态和 verbose 状态
- **AND** 插件完整模式下必须标识官方插件 workspace 是否参与执行

#### Scenario: 测试结束报告耗时
- **WHEN** `linactl test.go` 完成所有模块测试
- **THEN** 输出必须包含每个 module 的耗时摘要
- **AND** 如果存在 smoke 或集成型单元测试分类，输出必须能区分这些分类的耗时

### Requirement: PostgreSQL 测试基础设施不得产生无关 health check 错误
Go 单元测试 CI 的 PostgreSQL service health check SHALL 使用明确的数据库用户和数据库名，避免使用 runner 默认用户触发无关认证错误。

#### Scenario: PostgreSQL health check 使用测试数据库用户
- **WHEN** GitHub Actions 启动 Go 单元测试 PostgreSQL service
- **THEN** health check 命令必须使用 `postgres` 用户和 `linapro` 数据库进行 readiness 检查
- **AND** service 日志不得持续出现由默认 `root` 用户探测导致的 `role "root" does not exist` 错误

