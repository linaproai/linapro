## ADDED Requirements

### Requirement: 邮件协议与连接配置不得进入宿主核心契约

系统 SHALL 将邮件协议实现、Connection/Account 权威存储与邮件管理面保留在 `linapro-mail-core` 及其协议插件中。`apps/lina-core/pkg/plugin` MUST NOT 新增邮件协议 SPI 实现包或邮件 Connection 领域表访问作为长期公共契约。宿主 notify 可依赖邮件 owner 能力完成 email 通道投递，但 MUST NOT 将 SMTP/IMAP/POP3 细节嵌入宿主通用模块。

#### Scenario: 邮件公开契约位于 owner 插件

- **WHEN** 源码插件需要类型化邮件发送或接收能力
- **THEN** 公开契约 MUST 来自 `linapro-mail-core` 的 `backend/cap` 路径
- **AND** MUST NOT 要求从 `lina-core/pkg/plugin/capability` 导入邮件协议实现

#### Scenario: 宿主保持通用边界

- **WHEN** 审查邮件相关代码归属
- **THEN** 邮件 Connection 表与协议客户端 MUST 不落在 `lina-core` 宿主业务表与核心 service 中作为权威实现
- **AND** 宿主插件生命周期仅提供通用全局 Hook 机制，MUST NOT 硬编码邮件 kind 冲突规则

### Requirement: 宿主不拥有邮件业务表与厂商拆分

系统 SHALL NOT 在宿主 `sys_*` 基线中新增邮件 Connection/Account 权威表。按邮箱厂商（Gmail/QQ/iCloud 等）差异 MUST 通过 mail-core 多连接/账号配置表达，MUST NOT 在宿主内按厂商拆分模块。

#### Scenario: 多厂商通过配置而非宿主模块

- **WHEN** 管理员需要同时支持不同邮箱服务商的 SMTP/IMAP
- **THEN** 系统 MUST 通过 mail-core 的 Connection/设置配置完成
- **AND** MUST NOT 要求在 `lina-core` 增加厂商专用模块或表
