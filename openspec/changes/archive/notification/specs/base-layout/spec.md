## MODIFIED Requirements

### Requirement: 顶部导航栏
系统 SHALL 在顶部显示导航栏，包含用户信息、消息通知和操作入口。

#### Scenario: 显示用户信息
- **WHEN** 用户登录后查看顶部导航栏
- **THEN** 显示当前登录用户的用户名或昵称

#### Scenario: 显示消息通知铃铛
- **WHEN** 用户登录后查看顶部导航栏
- **THEN** 在用户头像左侧显示消息通知铃铛图标
- **THEN** 有未读消息时显示数量徽标

#### Scenario: 退出登录
- **WHEN** 用户点击顶部导航栏的退出按钮
- **THEN** 清除 Token，停止消息轮询，跳转到登录页面
