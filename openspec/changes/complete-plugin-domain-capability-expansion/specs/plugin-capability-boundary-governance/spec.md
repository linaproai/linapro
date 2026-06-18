## ADDED Requirements

### Requirement: 剩余领域能力不得扩大宿主核心数据边界
系统 SHALL 要求本变更新增的普通领域能力只返回领域 DTO、值对象、批量结果、分页结果、状态对象或结构化错误。新增方法 MUST NOT 暴露宿主核心表、官方能力插件内部表、`DAO/DO/Entity`、`*gdb.Model`、`*ghttp.Request`、原始 SQL、内部缓存快照、provider SPI 或物理存储路径。

#### Scenario: 插件读取组织档案
- **WHEN** 插件调用组织档案批量能力
- **THEN** 系统返回`orgcap`领域投影
- **AND** 响应不得包含 provider 内部模型、组织插件表实体或数据库查询条件

#### Scenario: 插件读取 storage 元数据
- **WHEN** 插件调用 storage 批量元数据方法
- **THEN** 系统返回插件私有对象领域元数据
- **AND** 响应不得包含宿主物理路径、bucket 密钥或 provider 私有配置

### Requirement: 管理命令和 Provider SPI 不得通过剩余阶段动态发布
系统 SHALL 禁止将本变更解释为动态插件管理能力授权、Provider SPI 授权或宿主内部治理授权。用户、租户、组织、插件、会话、任务、通知、storage、cache、manifest、`AI`的创建、更新、状态变更、授权关系变更和 provider 配置读取 MUST 继续留在管理面或宿主内部治理面，除非另开设计。

#### Scenario: 动态插件声明管理动作
- **WHEN** 动态插件试图通过本变更新增方法执行插件安装、用户授权、会话吊销或 provider 配置读取
- **THEN** 系统不得把这些动作纳入普通动态方法
- **AND** 如需开放必须另行设计管理授权、审计和失败语义

### Requirement: 动态 data 服务不得访问宿主核心表或官方能力插件内部表
系统 SHALL 要求本变更的`data.batch_get`只允许访问当前插件自有表和已授权资源表。动态插件声明`data`服务访问`sys_user`、`sys_role`、`sys_dict_data`、`sys_online_session`、`sys_plugin`或官方能力插件内部表时，治理校验 MUST 继续失败。

#### Scenario: data batch_get 声明核心表
- **WHEN** 动态插件声明`data.batch_get`访问`sys_user`
- **THEN** 安装、启用或运行时授权校验失败
- **AND** 插件必须改为使用`users.batch_get`、`users.resolve.batch`或其他领域能力
