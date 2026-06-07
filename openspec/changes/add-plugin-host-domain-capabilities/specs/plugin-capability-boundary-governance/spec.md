## ADDED Requirements

### Requirement: 插件生产代码不得依赖宿主核心表实现

系统 SHALL 禁止源码插件和动态插件生产代码生成或直接查询宿主核心`sys_*`表、响应宿主私有缓存快照或宿主内部 service 实现。宿主核心数据 MUST 由对应领域 owner 通过领域能力、`pluginhost.Services`或动态`hostServices`协议发布。Go 语言`internal`目录规则已经阻断的宿主`DAO/DO/Entity`导入和类型使用不作为治理扫描规则重复检查。

#### Scenario: 源码插件生成宿主核心表 DAO

- **WHEN** 插件`backend/hack/config.yaml`声明生成`sys_user`、`sys_role`、`sys_dict_data`、`sys_online_session`、`sys_plugin`或其他宿主核心表
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

### Requirement: 源码插件和动态插件必须共享领域能力语义

系统 SHALL 要求源码插件和动态插件访问同一宿主领域能力时共享领域 owner、输入输出`DTO`、领域`ID`类型、数据权限、缓存一致性、错误语义和`i18n`标签语义。动态插件`hostServices`handler 只能作为 transport 适配层，不得成为与源码插件平行的领域语义 owner。

#### Scenario: 两类插件读取用户投影

- **WHEN** 源码插件和动态插件分别读取用户基础投影
- **THEN** 二者最终进入同一`usercap.Service`语义
- **AND** 返回字段、缺失语义、数据权限边界和错误码保持一致

#### Scenario: 动态插件新增领域方法

- **WHEN** 宿主为动态插件新增一个领域`host service method`
- **THEN** 该方法必须映射到领域能力接口或受控领域适配器
- **AND** 不得只在`pluginbridge`或 WASM handler 中定义一套绕过源码插件语义的业务规则
