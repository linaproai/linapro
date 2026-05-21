## ADDED Requirements

### Requirement: Go 单元测试必须保留并发安全检测覆盖
Go 单元测试主路径 SHALL 保留 `-race` 对并发敏感代码的检测能力。降低总体耗时的实现 MUST 优先减少重复重型 fixture、真实链路执行和无测试包扫描成本，不得以无说明地移除主路径 race 覆盖作为默认优化手段。

### Requirement: 真实 dynamic Wasm 样例执行必须限定在 smoke 边界
Go 单元测试 SHALL 区分普通逻辑测试和真实 dynamic Wasm 样例链路测试。普通插件 runtime、catalog、integration 和 lifecycle 单测 MUST 优先使用 synthetic artifact、fake executor、轻量 host service 替身或测试辅助生成的 artifact；真实 bundled dynamic Wasm 样例执行 MUST 收敛为少量 smoke 覆盖。

### Requirement: 插件测试 fixture 必须复用且保持测试隔离
插件相关 Go 测试 SHALL 通过共享 helper 或包级 fixture 复用不可变基础资源，减少重复 artifact 写入、manifest 同步、插件安装启用、runtime cache 刷新和治理表清理。每个测试 MUST 仍保持自包含、顺序无关和数据隔离。

### Requirement: Go 测试入口必须避免完整执行无测试包
`linactl test.go` SHALL 在执行测试前生成测试计划，区分包含 `_test.go` 的包、仅需编译 smoke 的包和无需执行单测的包。默认单测入口 MUST 避免对大量无测试文件的包执行完整 `go test ./...` 主路径。

### Requirement: Go 单元测试运行必须输出可审计耗时摘要
Go 单元测试入口 SHALL 输出足够的执行计划和耗时摘要，使开发者能够持续识别最慢 module、测试包数量、无测试包数量、race 状态和真实 smoke 边界。

### Requirement: PostgreSQL 测试基础设施不得产生无关 health check 错误
Go 单元测试 CI 的 PostgreSQL service health check SHALL 使用明确的数据库用户和数据库名，避免使用 runner 默认用户触发无关认证错误。
