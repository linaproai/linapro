## ADDED Requirements

### Requirement: 标准构建必须区分 Host-only 与包含官方插件的完整构建

发布构建 SHALL 支持 host-only 构建和包含官方源码插件 submodule 的完整构建。Host-only 构建不得要求 `apps/lina-plugins` 存在；完整构建在需要官方源码插件后端、前端页面或动态插件产物时必须验证 submodule 已初始化。

#### Scenario: Host-only 发布构建
- **WHEN** 构建环境未初始化 `apps/lina-plugins` submodule
- **AND** 运维人员运行 host-only 构建入口
- **THEN** 构建产出不包含官方源码插件页面和官方插件后端注册
- **AND** 构建不得因缺少官方插件工作区失败

#### Scenario: 完整发布构建缺少 submodule
- **WHEN** 运维人员运行包含官方插件的完整构建入口
- **AND** `apps/lina-plugins` 不存在或为空
- **THEN** 构建快速失败
- **AND** 错误提示初始化官方插件 submodule

#### Scenario: 完整发布构建包含官方源码插件
- **WHEN** `apps/lina-plugins` submodule 初始化完成
- **AND** 运维人员运行完整发布构建入口
- **THEN** 构建包含官方源码插件后端注册和前端页面
- **AND** 动态插件产物按既有输出目录约定生成
