## ADDED Requirements

### Requirement: 动态插件 Storage Put 必须支持有界内存分片上传

系统 SHALL 允许动态插件通过`Storage().Put`写入大文件或未知大小输入，并在 guest SDK 内部按输入大小选择单次`storage.put`或分片上传。分片上传 MUST 使用`put.init`、`put.chunk`、`put.commit`和`put.abort`host service 方法完成传输；宿主 MUST 在 commit 时把已接收临时对象作为`io.Reader`交给`storagecap.Service.Put`写入最终 logical path。系统 MUST NOT 对最终对象大小设置动态 host service 固定上限；单个 chunk payload 可以有固定上限以保护 host call 内存边界。

#### Scenario: 小文件通过单次 host call 写入

- **WHEN** 动态插件调用`Storage().Put`且 guest SDK 能确认输入大小不超过单次上传阈值
- **THEN** guest SDK 可以继续使用`storage.put`一次性提交请求
- **AND** 宿主必须通过`storagecap.Service.Put`写入最终 logical path

#### Scenario: 大文件通过分片写入

- **WHEN** 动态插件调用`Storage().Put`且输入大小超过单次上传阈值
- **THEN** guest SDK 必须先调用`put.init`创建上传会话
- **AND** 按顺序调用`put.chunk`传输分片
- **AND** 调用`put.commit`提交总大小并写入最终 logical path
- **AND** 宿主在分片阶段不得把完整对象常驻内存

#### Scenario: 未知大小 reader 通过分片写入

- **WHEN** 动态插件调用`Storage().Put`且输入 reader 的总大小未知
- **THEN** guest SDK 必须使用分片上传路径
- **AND** 不得为了判断大小而先把完整 reader 读入内存

#### Scenario: 分片上传失败后清理会话

- **WHEN** 动态插件分片上传在任一 chunk 或 commit 阶段失败
- **THEN** guest SDK 必须尽力调用`put.abort`
- **AND** 宿主必须删除对应临时上传文件并释放会话状态

### Requirement: 动态插件 Storage 分片上传必须保持路径授权和会话绑定

系统 SHALL 对`put.init`、`put.chunk`、`put.commit`和`put.abort`执行与`storage.put`等价的 service、method 和`storage.resources.paths`授权校验。授权 path MUST 匹配最终插件 logical path，而不是 upload ID、provider object key、宿主物理路径或文件中心 ID。宿主 MUST 将 upload ID 绑定到当前插件 ID、最终 logical path 和上传会话状态，并在后续 chunk、commit 和 abort 中校验一致性。

#### Scenario: 插件声明分片上传方法

- **WHEN** 动态插件需要通过`Storage().Put`写入大文件或未知大小 reader
- **THEN** 插件必须在`plugin.yaml hostServices`的`service: storage`声明`put.init`、`put.chunk`、`put.commit`和`put.abort`
- **AND** 宿主授权快照必须包含这些方法后，guest SDK 才能完成分片上传流程

#### Scenario: 分片上传访问授权路径

- **WHEN** 动态插件已获授权访问 storage path `reports/`
- **AND** 插件通过分片上传写入 logical path `reports/demo.bin`
- **THEN** `put.init`、`put.chunk`、`put.commit`和`put.abort`都必须按`reports/demo.bin`执行路径授权
- **AND** 授权通过后才能访问上传会话或写入最终对象

#### Scenario: 分片上传访问未授权路径

- **WHEN** 动态插件只获授权访问 storage path `reports/`
- **AND** 插件尝试通过分片上传写入 logical path `secrets/demo.bin`
- **THEN** 宿主必须在进入上传会话或`storagecap.Service.Put`之前拒绝调用
- **AND** provider 不得收到该写入请求

#### Scenario: Upload ID 被挪用到其他路径

- **WHEN** 动态插件使用为`reports/a.bin`创建的 upload ID
- **AND** 后续 chunk、commit 或 abort 请求声明 logical path `reports/b.bin`
- **THEN** 宿主必须拒绝该请求
- **AND** 不得把临时对象提交到任一最终 logical path

#### Scenario: 分片 offset 不连续

- **WHEN** 动态插件提交的 chunk offset 不等于会话当前累计大小
- **THEN** 宿主必须拒绝该 chunk
- **AND** 不得把乱序或重叠内容写入临时对象
