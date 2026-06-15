## ADDED Requirements

### Requirement: Files 领域必须只表示宿主文件中心投影

系统 SHALL 将`Files()`领域能力限定为宿主文件中心资源投影和可见性校验能力。`Files()`方法 MUST 基于宿主文件领域 owner 的数据权限、租户边界和存在性不泄露策略返回文件投影或执行可见性确认，不得承担插件私有对象存储的内容读写生命周期。

#### Scenario: 插件读取宿主文件投影

- **WHEN** 源码插件或动态插件需要展示用户已上传到宿主文件中心的文件名称、MIME、大小或业务场景
- **THEN** 插件必须调用`Files().BatchGet`读取当前上下文可见的文件投影
- **AND** 响应必须使用`filecap.FileProjection`等领域 DTO
- **AND** 响应不得向插件暴露宿主文件中心`DAO`、`DO`、`Entity`、本地绝对路径或存储 provider 私有 key

#### Scenario: 插件校验宿主文件可见性

- **WHEN** 插件业务命令引用一批宿主文件中心文件 ID
- **THEN** 插件必须在命令执行前调用`Files().EnsureVisible`或等价领域校验
- **AND** 任一文件不存在、不可见或越过租户和数据权限边界时，命令必须整体拒绝
- **AND** 错误不得区分目标文件是真实不存在还是当前调用方不可见

#### Scenario: 动态插件声明文件领域服务

- **WHEN** 动态插件需要读取或校验宿主文件中心投影
- **THEN** 插件必须在`plugin.yaml hostServices`中声明`service: files`和所需文件领域方法
- **AND** 该声明不得使用`resources.paths`表达对象存储路径
- **AND** 宿主分发必须进入`filecap.Service`，不得转发到`storagecap.Service`

#### Scenario: 插件私有附件不使用 Files

- **WHEN** 插件需要保存、下载、删除、列出或清理插件私有附件对象
- **THEN** 插件必须使用`Storage()`领域能力
- **AND** 插件不得通过`Files()`领域方法把插件私有对象伪装为宿主文件中心资源
