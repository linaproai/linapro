## ADDED Requirements

### Requirement: 登录页按解析策略呈现
登录页 SHALL 根据当前生效的解析策略呈现不同 UI:
- subdomain 策略:不显示租户输入框(从 URL 自动解析)。
- header 策略:显示租户编码输入框(将作为登录前 `X-Tenant-Code` hint 发送,不得覆盖已签发正式 JWT 的 `TenantId`)。
- jwt/session/default 策略:不显示租户输入,凭 username/password 登录后由 select-tenant 决定。

#### Scenario: subdomain 模式登录
- **WHEN** 用户访问 `acme.app.com/login`
- **THEN** 登录页无租户输入
- **AND** 提交后 `/auth/login` 直接定位租户 acme

#### Scenario: 1:N 用户的挑选器
- **WHEN** 1:N 用户登录后 `/auth/login` 返回 pre_token + 租户列表
- **THEN** 前端在登录框内容区域切换为租户下拉挑选器(展示 code + name)
- **AND** 账号密码登录框与租户挑选器不得同时展示
- **AND** 挑选器说明文案应描述"选择本次要进入的租户"场景,不得继续提示填写账号密码
- **AND** 租户下拉控件应复用前端框架表单/选择器组件并与确认按钮保持清晰垂直间距
- **AND** 用户选定后调 `/auth/select-tenant`

### Requirement: prompt 模式下的可见性
当 ambiguous 行为 = `prompt` 时,挑选器 SHALL 提供清晰指引("您加入了多个组织,请选择登录目标");`reject` 模式下不展示挑选器,仅显示错误。

#### Scenario: prompt 模式
- **WHEN** 1:N 用户首次登录,服务端返回 `TENANT_REQUIRED`
- **THEN** 前端识别错误码,展示挑选器

### Requirement: 多租户禁用时退化
当 multi-tenant 未启用,登录页 SHALL 仅显示传统的 username/password,登录成功直接进入工作台。

#### Scenario: 单租户行为等价
- **WHEN** multi-tenant 未启用
- **THEN** 登录页与单租户时代视觉/交互一致
