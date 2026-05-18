## MODIFIED Requirements

### Requirement: 工作台头部租户标识与切换器
工作台头部 SHALL 在`multi-tenant`启用时展示当前租户标识；1:N用户的标识右侧增加切换下拉，展示用户可见的所有`active membership`租户。顶部租户切换下拉框 SHALL 保持与宿主头部控件一致的紧凑样式：固定宽度、单行截断、建筑图标、位于全局搜索入口左侧，并与搜索入口保留稳定间距。

#### Scenario: 1:N用户切换租户
- **WHEN** 用户 U 有 3 个`membership`
- **THEN** 头部下拉展示 3 个选项，当前租户带勾选标识
- **AND** 选择其他租户触发`/auth/switch-tenant`
- **AND** 客户端用新`token`重新拉菜单、权限和工作台数据
- **AND** 客户端强制进入新租户上下文的默认页面，不得继续停留在切换前页面

#### Scenario: 1:1用户或`single`模式
- **WHEN** 用户仅 1 个`membership`
- **THEN** 头部仅显示静态租户标识，无切换器

#### Scenario: 顶部租户切换器样式稳定
- **WHEN** `multi-tenant`已启用且管理工作台头部展示租户切换器
- **THEN** 下拉框保持固定宽度并对长租户名称单行截断
- **AND** 下拉框显示建筑图标
- **AND** 下拉框位于全局搜索入口左侧
- **AND** 下拉框与全局搜索入口之间保留稳定间距
