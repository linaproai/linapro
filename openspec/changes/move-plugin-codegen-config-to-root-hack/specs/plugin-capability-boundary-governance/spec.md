## MODIFIED Requirements

### Requirement: 插件生产代码不得依赖宿主核心表实现

系统 SHALL 禁止源码插件和动态插件生产代码生成或直接查询宿主核心`sys_*`表、响应宿主私有缓存快照或宿主内部 service 实现。宿主核心数据 MUST 由对应领域 owner 通过领域能力、`pluginhost.Services`或动态`hostServices`协议发布。Go 语言`internal`目录规则已经阻断的宿主`DAO/DO/Entity`导入和类型使用不作为治理扫描规则重复检查。

#### Scenario: 源码插件生成宿主核心表 DAO

- **WHEN** 插件根`hack/config.yaml`声明生成`sys_user`、`sys_role`、`sys_dict_data`、`sys_online_session`、`sys_plugin`或其他宿主核心表
- **THEN** 治理验证失败
- **AND** 插件必须改为依赖对应领域能力契约

#### Scenario: 插件生产代码直接查询宿主表

- **WHEN** 插件生产代码调用`g.DB().Model("sys_*")`、`shared.TableSysUser`或等价直接表入口
- **THEN** 治理验证失败
- **AND** 变更不得通过审查，除非该调用位于测试、Mock、安装 SQL 或迁移治理例外边界内

#### Scenario: 运行通用插件规范检查

- **WHEN** 开发者执行`make plugins.check`
- **THEN** 系统扫描`apps/lina-plugins`下所有包含`plugin.yaml`的插件目录
- **AND** 输出插件规范检查结果，发现违规时以非零状态退出

#### Scenario: 旧插件代码生成配置路径被拒绝

- **WHEN** 插件目录存在`backend/hack/config.yaml`
- **THEN** `make plugins.check`失败
- **AND** 错误消息提示将代码生成配置迁移到插件根`hack/config.yaml`

#### Scenario: 已有 DAO 生成物但缺少根配置被拒绝

- **WHEN** 插件目录存在`backend/internal/dao`生成物但缺少插件根`hack/config.yaml`
- **THEN** `make plugins.check`失败
- **AND** 错误消息提示补齐可重生成的代码生成配置

## ADDED Requirements

### Requirement: 源码插件编译接入必须由工具自动聚合

系统 SHALL 通过构建工具扫描官方插件工作区并自动生成源码插件后端聚合模块。插件作者 MUST 只维护插件自身`plugin.yaml`、插件 Go module 和`backend/plugin.go`等插件内资源，不得在`apps/lina-plugins`根目录维护手写`lina-plugins.go`、根`go.mod`或根`go.sum`来接入宿主编译。

#### Scenario: 源码插件通过自动聚合接入宿主编译

- **WHEN** 源码插件需要随宿主编译
- **THEN** 插件作者维护该插件目录内的`plugin.yaml`、`go.mod`和`backend/plugin.go`
- **AND** 构建工具在插件完整模式下自动生成`lina-plugins`聚合模块
- **AND** 插件作者不需要修改`apps/lina-plugins`根目录的 Go 文件或 Go module 文件
