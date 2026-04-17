## ADDED Requirements

### Requirement: 重置密码默认值由参数设置驱动
系统 SHALL 支持通过参数设置中的 `sys.user.initPassword` 控制用户管理页面“重置密码”弹窗中的默认密码回填值。

#### Scenario: 打开重置密码弹窗时回填当前初始密码参数
- **WHEN** 管理员打开某个用户的“重置密码”弹窗且 `sys.user.initPassword` 已配置
- **THEN** 密码输入框默认显示当前 `sys.user.initPassword` 的值

#### Scenario: 初始密码参数读取失败时仍可继续重置密码
- **WHEN** 管理员打开“重置密码”弹窗但 `sys.user.initPassword` 无法读取
- **THEN** 弹窗仍然正常打开
- **AND** 密码输入框不自动回填任何默认值
