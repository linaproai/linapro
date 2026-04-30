## ADDED Requirements

### Requirement: Built-in system parameter names and default copy must be localized in English
系统参数页面 SHALL 对框架内置参数名称、说明和默认展示内容按当前语言本地化，避免英文环境仍显示中文内置系统文案。

#### Scenario: Login and IP blacklist parameters display English metadata
- **WHEN** 管理员在 `en-US` 环境下打开系统参数页面
- **THEN** `sys.login.blackIPList`、`sys.auth.loginSubtitle`、`sys.auth.pageDesc`、`sys.auth.pageTitle` 的名称和说明显示英文
- **AND** 页面不显示 `用户登录-IP 黑名单列表`、`登录展示-登录副标题`、`登录展示-页面说明`、`登录展示-页面标题`

#### Scenario: Built-in public frontend copy can project English display content
- **WHEN** 系统参数列表在 `en-US` 环境下展示框架默认登录页标题、说明或副标题
- **THEN** 展示内容使用英文投影或英文默认值
- **AND** 编辑详情仍保留治理所需的稳定 `configKey` 和真实保存值，不把展示文案写回错误字段

#### Scenario: Config localization resources stay complete
- **WHEN** 新增或修改内置系统参数翻译键
- **THEN** `zh-CN`、`en-US`、`zh-TW` 运行时语言资源保持键集合一致
- **AND** 缺失翻译检查不得出现新增参数键缺失

### Requirement: Built-in system parameters must be editable but not deletable
系统参数 SHALL 标识系统内置记录，并禁止删除系统内置记录，同时继续允许管理员修改内置参数的可编辑字段和值。

#### Scenario: Built-in system parameter delete action is disabled
- **WHEN** 管理员查看系统参数列表中的系统内置参数
- **THEN** 该行删除按钮置灰且不会触发删除确认
- **AND** 鼠标悬停删除按钮时展示系统内置数据不支持删除的提示
- **AND** 编辑按钮仍可打开编辑表单

#### Scenario: Backend rejects built-in system parameter deletion
- **WHEN** 调用端绕过前端直接请求删除系统内置参数
- **THEN** 后端 SHALL 返回结构化业务错误并保留记录
- **AND** 非内置系统参数仍可按既有权限与校验规则删除
