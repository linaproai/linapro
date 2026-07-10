# go-static-lint-governance Specification

## Purpose
TBD - created by archiving change add-go-static-lint. Update Purpose after archive.
## Requirements
### Requirement: 仓库必须固定`golangci-lint`配置和版本

系统 SHALL 在仓库根目录维护`golangci-lint`配置、`golangci-lint`版本锁定文件和`staticcheck`版本锁定文件，用于定义`Go`静态检查规则、格式化器、生成代码排除和`nolint`治理策略。默认配置 MUST 明确排除不应手工维护的生成代码，MUST 不依赖开发者本机安装的未固定版本作为`CI`事实源。导入分组 formatter MAY 在后续完成全仓导入顺序规范化后启用为阻断门禁。

#### Scenario: 查看静态检查配置

- **WHEN** 开发者查看仓库根目录
- **THEN** 存在可被`golangci-lint`读取的配置文件
- **AND** 存在记录固定`golangci-lint`和`staticcheck`版本的仓库文件
- **AND** 配置包含适用于`lina-core`、`linactl`和官方插件`Go module`的格式化、生成代码排除和`nolint`治理策略

#### Scenario: 生成代码不要求手工整改

- **WHEN** 静态检查扫描 GoFrame 生成的`dao`、`do`、`entity`或其他带有生成标记的源码
- **THEN** 配置必须避免要求开发者手工修改这些生成文件
- **AND** 手写生产代码仍必须接受错误处理、静态分析和基础格式检查

### Requirement: `linactl`必须提供跨平台`Go`静态检查入口

系统 SHALL 通过`linactl lint.go`提供`Go`静态检查入口，并通过根`Makefile`和`make.cmd`薄包装暴露给开发者。命令实现 MUST 保持跨平台，不得把扫描、插件工作区准备或参数解析逻辑写入平台专属脚本、`Shell`管道或`Makefile`业务逻辑。

#### Scenario: 开发者运行宿主模式静态检查

- **WHEN** 开发者运行`make lint.go plugins=0`或等价`linactl lint.go plugins=0`
- **THEN** 命令在宿主`Go`工作区执行`golangci-lint`
- **AND** 扫描范围包含`apps/lina-core`和`hack/tools/linactl`
- **AND** 命令输出可审计的扫描模式、配置文件和检查结果

#### Scenario: 开发者在`Windows`运行静态检查入口

- **WHEN** 开发者在`Windows`环境运行`make.cmd lint.go plugins=0`
- **THEN** 命令通过`linactl`执行同一套静态检查逻辑
- **AND** 不要求`bash`、`sh`、`sed`、`awk`或其他`Unix`专属命令作为默认路径依赖

#### Scenario: 本地缺少锁定版本时自动安装

- **WHEN** 开发者运行`make lint.go plugins=0`或等价`linactl lint.go plugins=0`
- **AND** 本机`PATH`中的`golangci-lint`或`staticcheck`缺失，或版本不匹配对应仓库版本文件
- **THEN** 命令必须通过`go install`安装仓库锁定版本的缺失或不匹配工具
- **AND** 安装流程不得继承插件完整模式的临时`GOWORK`或`GOFLAGS`
- **AND** 后续静态检查和死代码检查必须使用锁定版本的二进制执行

#### Scenario: 环境初始化预安装静态检查工具

- **WHEN** 开发者运行`make env.setup`或等价`linactl env.setup`
- **THEN** 命令必须优先确保`golangci-lint`和`staticcheck`均为仓库锁定版本
- **AND** 缺失或版本不匹配时必须通过与`linactl lint.go`一致的`go install`路径安装
- **AND** 完成`Go`静态检查工具准备后再继续安装前端依赖和`Playwright`浏览器

### Requirement: 构建约束敏感的死代码检查必须按多目标归并

系统 SHALL 将死代码检查从单次`golangci-lint`扫描中拆分出来单独治理。`.golangci.yml` MUST NOT 启用独立`unused` linter；`linactl lint.go` MUST 使用仓库固定版本`staticcheck U1000`作为统一死代码门禁。普通包 MUST 在宿主默认目标下运行`U1000`；包含`wasip1`或`!wasip1`构建约束的包 MUST 额外归并宿主默认目标与`GOOS=wasip1 GOARCH=wasm` guest 目标结果，避免把仅在动态插件 guest 构建中使用的代码误报为未使用。

#### Scenario: guest 专属字段不被宿主目标误判

- **WHEN** 某个字段、函数或桥接路径只在`wasip1` guest 构建中被引用
- **AND** 宿主默认目标下该符号没有读路径
- **THEN** 多目标死代码检查不得将其报告为未使用
- **AND** `golangci-lint`不得通过独立`unused` linter 报告该符号

#### Scenario: 所有目标都未使用的代码仍然阻断

- **WHEN** 某个符号在宿主默认目标和`GOOS=wasip1 GOARCH=wasm` guest 目标中都没有使用路径
- **THEN** 多目标死代码检查必须报告该符号
- **AND** `linactl lint.go`返回非零退出码阻断本地和`CI`门禁

### Requirement: 静态检查必须覆盖官方插件完整工作区

系统 SHALL 支持`plugins=1`插件完整模式的`Go`静态检查。插件完整模式 MUST 复用官方插件工作区检查和临时`go.work`生成机制，覆盖宿主、工具、源码插件、动态插件构建相关`Go module`和必要的官方插件聚合模块。官方插件工作区缺失时，显式`plugins=1` MUST 失败并给出初始化提示。

#### Scenario: 插件完整模式扫描所有官方插件模块

- **WHEN** 官方插件工作区已初始化且开发者运行`make lint.go plugins=1`
- **THEN** 命令生成或复用插件完整模式的临时`go.work`
- **AND** 静态检查覆盖宿主模块、`linactl`模块和官方插件工作区下的`Go module`
- **AND** 输出标识当前执行的是插件完整模式

#### Scenario: 插件工作区缺失时拒绝插件完整模式

- **WHEN** 官方插件工作区缺失或未包含插件清单，且开发者运行`make lint.go plugins=1`
- **THEN** 命令失败
- **AND** 错误消息说明官方插件工作区状态和初始化方式

### Requirement: `CI`必须将`Go`静态检查作为质量门禁

系统 SHALL 在`GitHub Actions`中提供可复用的`Go`静态检查工作流，并接入现有主验证套件。主`CI`和发布验证默认 MUST 运行宿主模式和插件完整模式静态检查。`CI`执行路径 MUST 使用仓库固定的`golangci-lint`和`staticcheck`版本，并通过`make`或`linactl`入口运行和安装，保证本地与`CI`一致。

#### Scenario: 主`CI`运行静态检查

- **WHEN** 分支触发主`CI`
- **THEN** 验证套件运行`Go`静态检查作业
- **AND** 宿主模式和插件完整模式静态检查失败都会阻断`CI`
- **AND** 作业不使用`only-new-issues`作为长期默认豁免策略

#### Scenario: 发布验证运行静态检查

- **WHEN** 发布标签触发发布验证
- **THEN** 发布验证复用同一静态检查工作流
- **AND** 静态检查必须在镜像发布和`GitHub Release`创建前通过

### Requirement: 静态检查治理必须有文档和验证记录

系统 SHALL 在开发工具文档和规则说明中记录`Go`静态检查入口、参数、版本升级方式、跨平台边界、插件模式覆盖和失败处理方式。实现任务 MUST 记录跨平台影响、测试策略、`i18n`影响判断、缓存一致性无影响判断、数据权限无影响判断和实际验证命令。

#### Scenario: 开发者查看工具文档

- **WHEN** 开发者查看`linactl`或仓库开发工具说明文档
- **THEN** 文档说明如何运行宿主模式和插件完整模式`Go`静态检查
- **AND** 文档说明`golangci-lint`和`staticcheck`版本由仓库锁定
- **AND** 文档说明自动修复入口如存在必须由开发者显式触发

#### Scenario: 审查静态检查变更

- **WHEN** `lina-review`审查本变更实现
- **THEN** 审查结论必须包含开发工具跨平台影响和验证方式
- **AND** 审查结论必须记录无运行时`i18n`资源影响、无缓存一致性影响、无数据权限影响和无运行期服务依赖影响

