## ADDED Requirements

### Requirement: 动态插件清单可声明结构化宿主服务策略

系统 SHALL 允许动态插件在`plugin.yaml`中仅声明结构化`hostServices`策略，用于描述需要的宿主 service、method、资源申请和治理参数；宿主内部 capability 分类必须根据这些声明自动推导，而不是要求作者重复维护顶层`capabilities`字段。其中`storage`服务当前通过`resources.paths`声明逻辑路径申请，`data`服务当前通过`resources.tables`声明数据表申请。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者编写动态插件清单
- **THEN** 清单可以声明`hostServices`元数据
- **AND** 每个声明至少包含 service、method 集合以及资源申请或策略参数
- **AND** 清单不再需要单独声明顶层`capabilities`
- **AND** 构建器对未知 service、未知 method 和非法策略直接报错

#### Scenario: 宿主读取宿主服务策略快照

- **WHEN** 宿主查看一个动态插件的 manifest 快照或 release 快照
- **THEN** 宿主可以恢复该插件声明的宿主服务策略
- **AND** 管理员可以据此审查插件计划访问的宿主能力范围

#### Scenario: 插件声明资源申请而非宿主底层连接

- **WHEN** 开发者在清单中声明宿主服务依赖
- **THEN** 对`storage`服务，插件只声明稳定的逻辑路径或路径前缀`resources.paths`
- **AND** 对`network`服务，插件只声明 URL 模式列表
- **AND** 对`data`服务，插件在`resources`节点下声明需要访问的表名列表`tables`
- **AND** 对`cache`、`lock`和`notify`等低优先级服务，当前仍可继续使用逻辑`resourceRef`规划，其中分别表示缓存命名空间、逻辑锁名和通知通道标识
- **AND** 插件清单不得固化数据库连接、宿主文件绝对路径、缓存地址或密钥明文
- **AND** 真实资源绑定由宿主安装流程或管理员配置完成

### Requirement: 宿主服务资源申请纳入插件治理资源索引

系统 SHALL 将动态插件声明的宿主服务资源申请统一纳入`sys_plugin_resource_ref`治理资源索引；该表用于承载 release 级别的插件治理资源投影，而不只是镜像某个名为`resourceRef`的作者侧字段。对`storage`记录逻辑路径申请，对`network`记录 URL 模式申请，对`data`记录表名申请，对`cache`、`lock`、`notify`等低优先级服务继续记录逻辑资源引用。

#### Scenario: 安装或升级动态插件同步治理资源索引

- **WHEN** 宿主安装或升级一个声明了宿主服务资源的动态插件
- **THEN** 宿主将这些资源申请同步为插件资源归属记录
- **AND** 资源类型能够区分`host-storage`、`host-upstream`、`host-data-table`、`host-cache`、`host-lock`和`host-notify-channel`
- **AND** 这些记录可以参与审计、卸载和回滚治理

#### Scenario: 卸载或回滚动态插件更新治理资源索引

- **WHEN** 宿主卸载一个动态插件或将其回滚到旧 release
- **THEN** 宿主同步更新对应的宿主服务资源申请记录
- **AND** 当前 release 不再使用的逻辑路径、URL 模式、低优先级服务逻辑`resourceRef`或数据表声明不得继续保留为生效态

#### Scenario: 激活 release 时恢复逻辑引用绑定

- **WHEN** 宿主激活一个动态插件 release
- **THEN** 宿主根据 release 快照恢复资源申请的最终授权状态
- **AND** 运行时后续只按该快照解释宿主服务调用

### Requirement: 资源型宿主服务申请在安装或启用时需要宿主确认授权

系统 SHALL 在动态插件安装或启用阶段展示所有资源型宿主服务权限申请，并由宿主管理员确认最终授权结果。

#### Scenario: 安装时展示宿主服务权限申请

- **WHEN** 宿主准备安装一个声明了资源型 hostServices 的动态插件
- **THEN** 宿主展示插件申请的 service、method、资源标识（如`path`、URL 模式、`resourceRef`或`table`）及其治理参数摘要
- **AND** 当申请项属于`data` service 且宿主可解析表级说明时，宿主同时展示表名对应的人类可读说明，避免管理员只能依赖裸表名判断用途
- **AND** 管理员可以基于该清单审查插件计划访问的宿主资源范围

#### Scenario: 启用时确认或收窄宿主服务授权

- **WHEN** 宿主准备启用一个声明了资源型 hostServices 的动态插件 release
- **THEN** 宿主允许管理员批准、收窄或拒绝这些资源申请
- **AND** 宿主将最终确认结果持久化为当前 release 的授权快照
- **AND** 运行时后续只按这份最终快照解释宿主服务调用
