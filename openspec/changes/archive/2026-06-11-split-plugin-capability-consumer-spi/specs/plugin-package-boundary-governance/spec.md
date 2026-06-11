# plugin-package-boundary-governance 规范增量

## ADDED Requirements

### Requirement: capability 普通契约与 SPI 子包边界必须受治理验证

系统 SHALL 要求`pkg/plugin/capability/**`普通生产契约保持无 GoFrame HTTP 和数据库 query builder 依赖。除路径段以`spi`结尾的源码插件 provider SPI 子包外，`capability/**`非测试生产代码 MUST NOT import `github.com/gogf/gf/v2/database/gdb`或`github.com/gogf/gf/v2/net/ghttp`。该约束 MUST 由随`go test`执行的治理测试持续验证。

#### Scenario: 治理测试验证普通 capability 不导入 gdb

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `capability/**`中非`*spi`子包的非测试源文件不存在`github.com/gogf/gf/v2/database/gdb` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 治理测试验证普通 capability 不导入 ghttp

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `capability/**`中非`*spi`子包的非测试源文件不存在`github.com/gogf/gf/v2/net/ghttp` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: SPI 子包允许宿主接缝类型

- **WHEN** `tenantspi`或`orgspi`需要表达数据库 scope helper、request resolver 或 provider runtime
- **THEN** 对应 SPI 子包可以 import `gdb`或`ghttp`
- **AND** 该豁免不得扩散到父级`tenantcap`、`orgcap`或其他普通能力包

### Requirement: pluginbridge 不得依赖源码插件 Provider SPI

系统 SHALL 将`pkg/plugin/pluginbridge`限定为动态插件 ABI、transport、公开协议和动态插件专属 guest SDK。`pluginbridge/**`非测试生产代码 MUST NOT import `pkg/plugin/capability/**`下路径段以`spi`结尾的源码插件 provider SPI 子包。该约束 MUST 由随`go test`执行的治理测试持续验证。

#### Scenario: 动态插件 bridge 不导入 tenantspi

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `pluginbridge/**`非测试生产源文件不存在`pkg/plugin/capability/tenantcap/tenantspi` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 动态插件 bridge 不导入 orgspi

- **WHEN** 执行`pkg/plugin`的 import 边界治理测试
- **THEN** `pluginbridge/**`非测试生产源文件不存在`pkg/plugin/capability/orgcap/orgspi` import
- **AND** 违规时测试失败并指出违规文件和违规 import 路径

#### Scenario: 测试代码跨边界验证豁免

- **WHEN** `capability`、`pluginhost`或`pluginbridge`的`_test.go`文件为治理测试或集成验证 import SPI、`gdb`或`ghttp`
- **THEN** 治理测试不将其判定为违规
- **AND** 豁免仅适用于测试代码，不适用于任何生产源文件

