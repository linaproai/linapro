## ADDED Requirements

### Requirement: Dynamic route permission buttons must mount under their owning plugin menu
动态插件路由声明生成的按钮权限菜单 SHALL 挂载到所属动态插件页面菜单或插件主菜单下，不得在菜单管理中形成与业务入口脱节的漂浮按钮集合。

#### Scenario: Dynamic plugin route buttons are children of plugin menu
- **WHEN** `plugin-demo-dynamic` 启用并同步动态路由权限
- **THEN** 对应按钮权限出现在该插件页面菜单或插件主菜单的子节点下
- **AND** 菜单管理列表不显示大量以 `Dynamic Route Permission:plugin-demo-dynamic:` 为前缀且缺少合理父菜单语义的顶层或错挂按钮

#### Scenario: Dynamic plugin buttons remain assignable
- **WHEN** 管理员在角色授权树中查看动态插件权限
- **THEN** 动态路由按钮仍随所属插件菜单参与授权
- **AND** 插件停用后授权关系按现有插件菜单治理规则暂时失效但不丢失

### Requirement: Menu tree expandable rows must be clickable
菜单管理树形列表 SHALL 为可展开的目录或菜单行提供清晰的可点击鼠标形状，并允许点击节点标题区域展开或折叠。

#### Scenario: Expandable menu row pointer and click
- **WHEN** 管理员将鼠标移动到存在子节点的目录或菜单行标题区域
- **THEN** 鼠标形状变为可点击指针
- **AND** 点击该区域会展开或折叠该节点

#### Scenario: Redundant title icon hint is removed
- **WHEN** 管理员查看菜单管理列表标题区域
- **THEN** 标题后的额外展开提示图标不再展示
- **AND** 树节点自身的展开交互仍保持可发现和可操作
