## ADDED Requirements

### Requirement: 官方插件工作区必须可选

系统 SHALL 将 `apps/lina-plugins` 视为官方源码插件工作区的默认位置，但宿主基础开发、构建和测试入口不得要求该目录一定存在或一定包含插件代码。该目录可以不存在、为空目录，或作为 `git submodule` 挂载并包含官方插件代码。

#### Scenario: 插件工作区不存在时宿主命令可执行
- **WHEN** 工作区中不存在 `apps/lina-plugins`
- **THEN** 宿主后端编译和宿主后端单元测试仍可执行
- **AND** 失败不得来自 Go workspace 加载缺失插件模块或宿主编译期导入官方插件聚合包

#### Scenario: 插件工作区为空时宿主命令可执行
- **WHEN** 工作区中存在空目录 `apps/lina-plugins`
- **THEN** 宿主后端编译和宿主后端单元测试仍可执行
- **AND** 源码插件发现结果为空集合
- **AND** 宿主启动不得因为没有源码插件清单而失败

#### Scenario: 官方插件通过 submodule 恢复
- **WHEN** `apps/lina-plugins` 作为 submodule 初始化完成
- **THEN** 宿主可发现官方源码插件清单
- **AND** 官方源码插件后端、前端、manifest、SQL 和 E2E 继续使用 `apps/lina-plugins/<plugin-id>/` 目录约定

### Requirement: 宿主与官方源码插件编译期依赖必须解耦

宿主 SHALL 能在没有官方源码插件 Go module 的情况下完成默认编译。官方源码插件后端注册必须由显式的插件完整构建路径、构建标签、生成聚合文件或等价机制启用，不得由宿主默认入口无条件依赖 `lina-plugins` module。

#### Scenario: 默认宿主构建不导入官方插件聚合模块
- **WHEN** 开发者在未初始化 `apps/lina-plugins` 的工作区中构建 `apps/lina-core`
- **THEN** 构建不解析 `lina-plugins` module
- **AND** 构建不要求任何 `lina-plugin-*` module 存在

#### Scenario: 完整插件构建启用官方源码插件注册
- **WHEN** 开发者显式运行包含官方插件的完整构建或测试入口
- **THEN** 构建路径启用官方源码插件后端注册
- **AND** 如果 `apps/lina-plugins` submodule 未初始化，命令快速失败并提示初始化 submodule

### Requirement: 插件工作区状态必须可诊断

开发工具 SHALL 能区分插件工作区缺失、空目录、已初始化 submodule 和路径存在但结构无效四类状态，并在需要插件内容的命令中输出可操作诊断。

#### Scenario: 需要插件内容但 submodule 未初始化
- **WHEN** 开发者运行官方插件单元测试、插件 E2E、动态插件 wasm 构建或完整插件构建
- **AND** `apps/lina-plugins` 不存在或为空
- **THEN** 命令失败并说明当前缺少官方插件工作区
- **AND** 错误提示包含 `git submodule update --init --recursive`

#### Scenario: 宿主命令遇到缺失插件工作区
- **WHEN** 开发者运行 host-only 宿主命令
- **AND** `apps/lina-plugins` 不存在或为空
- **THEN** 命令继续执行
- **AND** 诊断最多以信息级输出当前处于 host-only 模式

### Requirement: Host-only 与 Plugin-full 验证必须分离

系统 SHALL 提供可重复执行的 host-only 验证和 plugin-full 验证。Host-only 验证必须在插件工作区不存在和为空两种状态下覆盖宿主后端、宿主前端和宿主 E2E；plugin-full 验证必须在 submodule 初始化后覆盖官方插件后端单元测试和插件 E2E。

#### Scenario: Host-only 验证覆盖插件目录缺失状态
- **WHEN** `apps/lina-plugins` 被临时移出工作区
- **THEN** host-only 验证运行宿主后端单元测试
- **AND** host-only 验证运行宿主前端类型检查或构建
- **AND** host-only 验证运行不依赖官方源码插件内容的宿主 E2E

#### Scenario: Host-only 验证覆盖插件目录为空状态
- **WHEN** `apps/lina-plugins` 是空目录
- **THEN** host-only 验证运行同一组宿主验证
- **AND** 插件管理或插件发现相关宿主测试断言源码插件集合为空时行为正确

#### Scenario: Plugin-full 验证覆盖 submodule 状态
- **WHEN** `apps/lina-plugins` submodule 初始化完成
- **THEN** plugin-full 验证运行所有官方插件 Go 单元测试
- **AND** plugin-full 验证运行所有官方插件自有 E2E
- **AND** plugin-full 验证运行动态插件 wasm 构建或等价插件产物验证

### Requirement: 官方插件 submodule 迁移必须保持默认路径

官方插件仓库 SHALL 作为单个 submodule 挂载到 `apps/lina-plugins`，除非后续 OpenSpec 变更明确批准路径迁移。主仓库文档、CI 和开发命令必须围绕该默认挂载路径描述初始化与验证流程。

#### Scenario: 克隆主仓库后初始化官方插件
- **WHEN** 开发者克隆主仓库并需要完整官方插件能力
- **THEN** 文档提供 submodule 初始化命令
- **AND** 初始化后 `apps/lina-plugins` 下出现官方插件目录
- **AND** 现有插件目录结构规范继续适用

#### Scenario: 主仓库只运行宿主能力
- **WHEN** 开发者克隆主仓库但不初始化 submodule
- **THEN** 文档说明可运行 host-only 开发和测试入口
- **AND** 文档说明插件专属命令需要初始化 submodule
