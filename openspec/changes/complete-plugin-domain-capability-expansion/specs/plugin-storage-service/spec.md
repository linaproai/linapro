## ADDED Requirements

### Requirement: Storage 必须支持插件私有对象批量元数据
系统 SHALL 提供`Storage.BatchStat`和动态`storage.stat.batch`，按路径集合批量返回当前插件和租户作用域内的对象元数据。请求 MUST 限制路径数量、单路径长度和总字节数，响应 MUST 使用领域元数据 DTO。

#### Scenario: 批量读取对象元数据
- **WHEN** 插件请求多个 storage 路径的元数据
- **THEN** 系统一次性返回可见路径的元数据和不透明缺失集合
- **AND** 不得逐路径调用单对象 stat 作为常规实现

### Requirement: Storage 必须支持有界游标列表
系统 SHALL 提供`Storage.ListCursor`和动态`storage.list.cursor`，按前缀、cursor 和 limit 返回插件私有对象列表。实现 MUST 不提供无界前缀遍历。

#### Scenario: 游标读取下一页对象
- **WHEN** 插件携带 prefix、cursor 和 limit 请求 storage 列表
- **THEN** 系统返回当前页对象元数据和 next cursor
- **AND** 每页数量不得超过领域上限

### Requirement: Storage 必须支持批量删除明确路径集合
系统 SHALL 提供`Storage.DeleteMany`和动态`storage.delete.batch`，只删除当前插件和租户作用域下明确路径集合。删除缺失对象 SHOULD 作为 no-op 成功处理，除非 provider 返回不可恢复错误。

#### Scenario: 批量删除私有对象
- **WHEN** 插件提交多个明确 storage 路径删除
- **THEN** 系统在资源授权和插件/租户作用域内删除这些路径
- **AND** 不接受宿主物理路径或无边界前缀删除

### Requirement: Storage provider 选择不得依赖主配置项
系统 SHALL 不通过主框架配置文件选择`Storage`provider。宿主 MUST 在当前已注册 storage provider 插件中选择唯一一个平台可服务插件；没有可服务 provider 插件时 MUST 使用内置本地文件 provider。

#### Scenario: 未安装或未启用 storage provider 插件
- **WHEN** 插件调用`Storage`
- **AND** 当前没有可服务的 storage provider 插件
- **THEN** 系统使用内置本地文件 provider 处理请求

#### Scenario: 唯一 provider 插件可服务
- **WHEN** 插件调用`Storage`
- **AND** 当前恰好一个已注册 storage provider 插件处于平台可服务状态
- **THEN** 系统使用该 provider 插件处理请求

#### Scenario: 多个 provider 插件同时可服务
- **WHEN** 插件调用`Storage`
- **AND** 当前多个已注册 storage provider 插件处于平台可服务状态
- **THEN** 系统拒绝本次 storage 请求并返回`CodeStorageProviderConflict`
- **AND** 不得静默选择任意一个 provider 或回退到本地文件 provider
