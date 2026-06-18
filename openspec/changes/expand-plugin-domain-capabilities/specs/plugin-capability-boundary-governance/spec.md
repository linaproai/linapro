## ADDED Requirements

### Requirement: 领域能力补充不得扩大宿主核心数据边界

系统 SHALL 要求`expand-plugin-domain-capabilities`新增的普通领域能力只返回领域 DTO、值对象、批量结果或结构化错误。新增方法 MUST NOT 暴露宿主核心表、官方能力插件内部表、`DAO/DO/Entity`、`*gdb.Model`、`*ghttp.Request`、原始 SQL、内部缓存快照或 provider SPI。

#### Scenario: 用户批量解析返回领域投影

- **WHEN** 插件调用用户批量解析方法
- **THEN** 系统返回`usercap.UserProjection`等领域投影和不透明`MissingIDs`
- **AND** 响应不得包含`sys_user`实体、数据库自增主键模型、数据权限 SQL 或内部缓存字段

#### Scenario: 当前会话返回最小投影

- **WHEN** 插件调用当前会话投影方法
- **THEN** 系统返回`sessioncap.Projection`或等价领域 DTO
- **AND** 响应不得暴露 session store 私有缓存结构、token 存储后端或撤销标记内部实现

### Requirement: 阶段一不得开放管理命令或动态数据核心表访问

系统 SHALL 禁止将阶段一领域能力扩展解释为动态插件管理能力授权、Provider SPI 授权或动态`data`服务访问宿主核心表的授权。用户、权限、字典和会话的创建、更新、状态变更、吊销、授权关系变更和 provider 配置读取 MUST 继续留在管理面或宿主内部治理面。

#### Scenario: 动态插件声明管理动作

- **WHEN** 动态插件试图通过本变更新增方法执行用户状态变更、会话吊销、字典刷新或权限授权关系变更
- **THEN** 系统不得把这些动作纳入阶段一动态普通方法
- **AND** 如需开放必须另行设计管理授权、审计、数据权限和失败语义

#### Scenario: 动态 data 服务声明核心表

- **WHEN** 动态插件声明`data`服务访问`sys_user`、`sys_role`、`sys_dict_data`、`sys_online_session`或`sys_plugin`
- **THEN** 治理校验必须继续失败
- **AND** 插件必须改为依赖本变更发布的普通领域能力或后续正式发布的领域能力
