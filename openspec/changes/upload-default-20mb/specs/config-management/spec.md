## ADDED Requirements

### Requirement: 默认上传大小必须统一为20MB
系统 SHALL 将 `sys.upload.maxSize` 的平台默认值统一设为 `20`，并让数据库初始化、配置模板和运行时上传默认回退都以该值为基线，除非管理员显式修改该参数。

#### Scenario: 宿主初始化写入 20MB 默认值
- **WHEN** 管理员执行宿主初始化 SQL
- **THEN** `sys_config` 中 `sys.upload.maxSize` 的默认值必须为 `20`
- **AND** 配置管理读取到的该内建参数默认值也必须为 `20`

#### Scenario: 未覆写配置时运行时默认上限为 20MB
- **WHEN** 宿主在未被管理员覆写上传大小配置的情况下处理 `multipart` 上传请求
- **THEN** 文件上传校验必须按 20MB 上限执行
- **AND** 因默认上限触发的友好错误提示必须返回“文件大小不能超过20MB”等价语义

### Requirement: 默认上传大小的多来源配置必须保持一致
系统 SHALL 保证 `sys.upload.maxSize` 的数据库种子值、配置模板默认值和宿主静态回退值保持一致，避免不同启动路径出现不同的默认上传上限。

#### Scenario: 使用默认模板启动宿主
- **WHEN** 操作者使用宿主默认 `config.template.yaml` 生成运行配置且未单独修改上传大小
- **THEN** 宿主读取到的默认上传大小必须为 20MB
- **AND** 该默认值必须与宿主初始化 SQL 中的 `sys.upload.maxSize` 默认值一致
