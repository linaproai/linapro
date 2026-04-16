# plugin-storage-service Specification

## Purpose
TBD - created by archiving change dynamic-plugin-host-service-extension. Update Purpose after archive.
## Requirements
### Requirement: 动态插件通过逻辑存储空间访问文件

系统 SHALL 为动态插件提供受隔离的存储服务，插件只能通过宿主授权的逻辑存储空间访问文件，不能直接指定宿主物理路径。

#### Scenario: 插件写入授权存储空间

- **WHEN** 插件调用存储服务写入文件
- **THEN** 请求必须指向该插件已授权的逻辑存储空间或对象引用
- **AND** 宿主按插件隔离目录或对象前缀保存文件
- **AND** 宿主返回文件标识、大小和元数据摘要

#### Scenario: 插件读取授权存储对象

- **WHEN** 插件调用存储服务读取文件
- **THEN** 宿主仅允许读取当前插件被授权访问的逻辑对象
- **AND** 宿主不得向 guest 暴露宿主物理文件路径

#### Scenario: 插件尝试访问未授权路径

- **WHEN** 插件尝试通过路径拼接、目录穿越或未授权`resourceRef`访问文件
- **THEN** 宿主拒绝该调用
- **AND** 宿主不暴露宿主真实文件系统结构

### Requirement: 宿主存储服务实施大小、类型和公开性治理

系统 SHALL 对动态插件的文件读写操作实施大小限制、类型限制和可见性治理。

#### Scenario: 宿主校验文件写入约束

- **WHEN** 插件向某个逻辑存储空间写入或覆盖文件
- **THEN** 宿主根据该空间策略校验最大大小、允许类型和覆盖规则
- **AND** 不符合策略的请求被拒绝

#### Scenario: 插件请求对外暴露文件

- **WHEN** 插件请求生成文件的对外访问地址
- **THEN** 宿主仅对声明为公开或允许签名访问的逻辑存储空间返回可访问地址
- **AND** 私有存储空间不得返回永久公开链接

