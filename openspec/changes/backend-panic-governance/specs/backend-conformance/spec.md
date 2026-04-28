## ADDED Requirements

### Requirement: 运行期错误不得以 panic 替代显式错误处理
后端生产代码 SHALL 仅在启动期、初始化期、不可回滚关键链路、`Must*` 语义构造函数或未知 panic 重新抛出场景中使用 `panic`；普通请求、导入导出、动态插件输入、运行时配置读取和可恢复资源处理路径 MUST 通过显式 `error` 返回、统一错误响应或受控降级处理。

#### Scenario: 启动期不可恢复错误使用 fail-fast
- **WHEN** 后端在进程启动、驱动注册、命令树初始化或源码插件静态注册阶段发现不可恢复错误
- **THEN** 代码 MAY 使用 `panic` 让进程快速失败
- **AND** 该 panic 调用点 MUST 位于 allowlist 并说明保留原因

#### Scenario: 普通业务请求返回错误
- **WHEN** 普通 HTTP 请求、文件导入导出、Excel 生成或资源关闭遇到可处理错误
- **THEN** 服务或控制器 MUST 通过 `error` 返回让统一错误处理链路生成响应
- **AND** 不得使用 `panic` 代替错误返回

#### Scenario: 动态插件输入校验失败
- **WHEN** 动态插件产物、清单、hostServices 声明或授权输入不合法
- **THEN** 宿主 MUST 返回带上下文的校验错误
- **AND** 不得因为插件提供的动态输入触发生产代码 panic

#### Scenario: 运行时配置异常值显式返回
- **WHEN** 受保护运行时配置在读取快照时出现解析错误
- **THEN** 后端 MUST 通过显式 `error` 返回或统一错误响应暴露该配置异常
- **AND** 写入路径仍 MUST 保持严格校验，防止正常管理入口保存非法值

#### Scenario: 新增 panic 被静态检查约束
- **WHEN** 开发者在后端生产 Go 代码中新增 `panic` 调用
- **THEN** 自动化检查 MUST 要求该调用点匹配 allowlist
- **AND** allowlist 条目 MUST 标注所属类别和保留理由
