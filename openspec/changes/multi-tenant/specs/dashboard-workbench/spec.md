## ADDED Requirements

### Requirement: 工作台头部租户标识与切换器
工作台头部 SHALL 在 multi-tenant 启用时展示当前租户标识(租户 name + code);1:N 用户的标识右侧增加切换下拉,展示用户可见的所有 active membership 租户。

#### Scenario: 1:N 用户切换租户
- **WHEN** 用户 U 有 3 个 membership
- **THEN** 头部下拉展示 3 个选项,当前租户带勾选标识
- **AND** 选择其他租户触发 `/auth/switch-tenant`
- **AND** 客户端用新 token 重新拉菜单/权限/工作台数据
- **AND** 客户端强制进入新租户上下文的默认页面,不得继续停留在切换前页面

#### Scenario: 1:1 用户(或 single 模式)
- **WHEN** 用户仅 1 个 membership
- **THEN** 头部仅显示静态租户标识,无切换器

### Requirement: 平台管理员的特殊头部样式
平台管理员视图(`bizctx.TenantId=0`)SHALL 在头部展示醒目"平台管理员"标识(色条/标签),与租户视图视觉区分;impersonation 模式下展示"代为操作租户 X"提示条。

#### Scenario: impersonation 视觉提示
- **WHEN** 平台管理员开始 impersonation 租户 A
- **THEN** 工作台顶部出现红色提示条"您正以租户 A 视角操作,点击退出"
- **AND** 该提示条位于顶部"切换租户"下拉框左侧
- **AND** 开始 impersonation 后客户端强制进入租户默认页面,不得继续停留在平台页面
- **AND** 退出按钮调 `/platform/tenants/A/end-impersonate`
- **AND** 客户端恢复平台管理员原 token、清空当前租户上下文、重新拉取菜单/权限/工作台数据并强制进入平台默认页面

### Requirement: 多租户禁用时隐藏租户 UI
当 multi-tenant 未启用,工作台头部 SHALL 不出现任何租户标识或切换器;UI 退化为单租户。

#### Scenario: 未启用时无租户头部元素
- **WHEN** multi-tenant 未启用,用户登录工作台
- **THEN** 工作台头部不展示租户标识、切换器或 impersonation 提示
- **AND** 与单租户时代视觉一致
