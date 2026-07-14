# auth-account-recovery-and-register Specification

## Purpose
TBD - created by archiving change login-forgot-password-and-register. Update Purpose after archive.
## Requirements
### Requirement: 登录页提供忘记密码与创建账号入口

系统 SHALL 在标准登录页按 Vben 认证布局展示「忘记密码」与「创建账号」入口（受系统开关控制），并导航到对应认证子路由。创建账号入口必须位于「其他登录方式」区域之后。

#### Scenario: 从登录页进入忘记密码
- **当** 忘记密码开关开启且未认证用户在 `/auth/login` 点击忘记密码入口时
- **则** 浏览器导航到 `/auth/forget-password`
- **且** 页面展示忘记密码标题、邮箱输入与发送重置链接操作

#### Scenario: 从登录页进入创建账号
- **当** 创建账号开关开启且未认证用户在 `/auth/login` 点击创建账号入口时
- **则** 浏览器导航到 `/auth/register`
- **且** 页面展示创建账号标题、注册表单字段与注册操作

#### Scenario: 创建账号入口位于其他登录方式下方
- **当** 未认证用户访问 `/auth/login` 且创建账号开关开启时
- **则** 页面在「其他登录方式」区域（或该区域隐藏占位）之后展示「还没有账号? 创建账号」
- **且** 创建账号入口不出现在主登录按钮与外部/社交登录区域之间

### Requirement: 忘记密码与创建账号可由系统参数开关控制

系统 SHALL 通过宿主系统参数控制忘记密码与创建账号公开入口，默认均为开启。关闭后登录页隐藏对应入口，直接访问对应子路由时回退到标准登录页。

#### Scenario: 默认开启时展示入口
- **当** 系统参数未覆盖或值为 `true` 时
- **则** 登录页展示忘记密码入口与创建账号入口
- **且** 公开前端配置返回对应布尔字段为 `true`

#### Scenario: 关闭忘记密码后隐藏入口并回退路由
- **当** 系统参数 `sys.auth.forgetPasswordEnabled` 为 `false` 时
- **则** 登录页不显示忘记密码入口
- **且** 访问 `/auth/forget-password` 时重定向到 `/auth/login`

#### Scenario: 关闭创建账号后隐藏入口并回退路由
- **当** 系统参数 `sys.auth.registerEnabled` 为 `false` 时
- **则** 登录页不显示创建账号入口
- **且** 访问 `/auth/register` 时重定向到 `/auth/login`

### Requirement: 注册协议可弹窗阅读且正文可配置

系统 SHALL 在创建账号页提供隐私政策与服务条款链接；点击后以弹窗展示正文。正文由宿主系统参数 `sys.auth.privacyPolicy` 与 `sys.auth.termsOfService` 配置，并通过公开前端配置下发。

#### Scenario: 点击隐私政策打开弹窗
- **当** 用户在创建账号页点击隐私政策链接时
- **则** 系统展示包含隐私政策正文的弹窗

#### Scenario: 点击服务条款打开弹窗
- **当** 用户在创建账号页点击服务条款链接时
- **则** 系统展示包含服务条款正文的弹窗

#### Scenario: 管理员修改协议正文后注册页展示新内容
- **当** 管理员更新 `sys.auth.privacyPolicy` 或 `sys.auth.termsOfService` 后
- **则** 公开前端配置返回新正文
- **且** 注册页协议弹窗展示更新后的内容

### Requirement: 公开注册创建平台账号

系统 SHALL 在注册开关开启时通过 `POST /auth/register` 创建平台用户：唯一用户名与邮箱、密码 6–32、状态启用、分配内置 `user` 角色、不自动登录。

#### Scenario: 合法注册成功后可登录
- **当** 注册开关开启且用户提交合法用户名、邮箱与密码时
- **则** 系统创建账号并返回成功
- **且** 用户可使用该用户名密码登录

#### Scenario: 用户名冲突
- **当** 注册用户名已存在时
- **则** 系统返回用户名已存在错误且不创建账号

### Requirement: 邮件密码重置

系统 SHALL 在忘记密码开关开启且邮件通道可用时，通过 `POST /auth/forget-password` 受理重置请求，并向匹配邮箱发送含一次性令牌的链接；`POST /auth/reset-password` 消费令牌并更新密码。

#### Scenario: 合法邮箱提交重置请求
- **当** 忘记密码开关开启、邮件通道可用且邮箱格式合法时
- **则** 系统返回受理成功
- **且** 若账号存在则发送重置邮件

#### Scenario: 使用令牌重置密码
- **当** 用户持有未过期未使用的重置令牌并提交新密码时
- **则** 系统更新密码
- **且** 旧密码无法再登录、新密码可以登录

#### Scenario: 邮件通道不可用
- **当** 忘记密码开关开启但邮件投递实现未注册时
- **则** 系统返回密码找回暂时不可用错误

