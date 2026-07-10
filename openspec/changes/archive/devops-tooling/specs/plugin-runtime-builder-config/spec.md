# 动态插件 Builder 配置收敛

## Purpose
将动态插件 builder 专用 hook、resource 与 lifecycle timeout 配置统一到插件根 hack/config.yaml 的 wasm 配置，作为源码目录形态的唯一配置来源。

## Requirements

### Requirement: 动态插件后端 builder 配置必须统一维护在插件根配置

系统 SHALL 将动态插件 builder 专用的后端 hook 和 backend resource 声明统一维护在插件根`hack/config.yaml`的`wasm.hooks`和`wasm.resources`下。构建工具和宿主本地动态插件目录加载逻辑 MUST 使用该配置作为源码目录形态的唯一配置来源。系统 MUST NOT 从`backend/hooks/*.yaml`、`backend/resources/*.yaml`或其他`backend/*/*.yaml`文件读取、扫描或检测 builder 配置。

#### Scenario: 从插件根配置打包 hook 和 resource

- **WHEN** 动态插件根`hack/config.yaml`声明`wasm.hooks`和`wasm.resources`
- **THEN** `build-wasm`将这些声明打包到 runtime artifact 对应的 hook 和 resource sections
- **AND** 宿主解析该 artifact 后可以从 artifact sections 水合 hook 和 backend resource 契约

#### Scenario: 宿主加载本地动态插件目录时使用同一配置来源

- **WHEN** 宿主需要从本地动态插件源码目录加载期望 manifest
- **AND** 插件根`hack/config.yaml`声明`wasm.hooks`和`wasm.resources`
- **THEN** 宿主 manifest catalog 使用同一配置水合 hook 和 backend resource 契约
- **AND** 不扫描或检测`backend/hooks`、`backend/resources`或其他`backend/*/*.yaml`路径

#### Scenario: hook duration 使用带单位字符串

- **WHEN** 插件根`hack/config.yaml`在`wasm.hooks`中声明`timeout`或`sleep`
- **THEN** 该值必须是带单位的时间长度字符串，例如`50ms`或`3s`
- **AND** 构建工具将其转换为 runtime artifact 中的毫秒字段
- **AND** 不接受`timeoutMs`或`sleepMs`作为配置输入


### Requirement: 动态插件生命周期声明必须作为自动发现契约的可选覆盖

系统 SHALL 使用 controller 自动发现生成动态插件生命周期契约，并且仅允许插件根`hack/config.yaml`的`wasm.lifecycle.timeouts`为已发现生命周期 handler 声明超时时长。该配置 MUST 使用带单位的时间长度字符串，例如`3s`、`2m`或`50ms`，并由构建工具转换为 artifact 契约中的毫秒字段。系统 MUST NOT 从`backend/lifecycle/*.yaml`或源码常量读取、扫描或检测生命周期 builder 配置。

#### Scenario: 插件根配置覆盖生命周期超时

- **WHEN** 动态插件 controller 暴露合法`BeforeInstall`生命周期方法
- **AND** 插件根`hack/config.yaml`声明`wasm.lifecycle.timeouts.BeforeInstall=3s`
- **THEN** `build-wasm`生成`BeforeInstall`生命周期契约
- **AND** 该契约的 timeout 使用配置声明的`3000`毫秒

#### Scenario: 配置声明不存在的方法

- **WHEN** 插件根`hack/config.yaml`声明`wasm.lifecycle.timeouts.BeforeInstall=3s`
- **AND** 动态插件 controller 未暴露合法`BeforeInstall`handler
- **THEN** `build-wasm`构建失败
- **AND** 错误信息指向该 lifecycle timeout 找不到对应 handler

#### Scenario: 配置声明非法时长

- **WHEN** 插件根`hack/config.yaml`声明不带单位的`wasm.lifecycle.timeouts.BeforeInstall=3000`
- **THEN** `build-wasm`构建失败
- **AND** 错误信息指向 lifecycle timeout 必须使用带单位的时间长度字符串
