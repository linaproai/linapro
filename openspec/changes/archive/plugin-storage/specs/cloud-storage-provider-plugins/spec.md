## Requirements

### Requirement: 系统必须提供云对象存储 provider 源码插件（厂商 + 协议）

系统 SHALL 交付以下官方 source 插件，并分别通过 `storagecap.Provide(pluginID, factory)` 注册 `storagecap.Provider` 实现：

- `linapro-storage-cos`：腾讯云 COS（厂商 SDK）
- `linapro-storage-oss`：阿里云 OSS（厂商 SDK）
- `linapro-storage-obs`：华为云 OBS（厂商 SDK）
- `linapro-storage-qiniu`：七牛云 Kodo（厂商 SDK；region 可选，留空时 MUST 根据 AccessKey 与 bucket 自动探测；endpoint 可选下载域名）
- `linapro-storage-aws`：AWS S3（厂商插件；配置面向 AWS，region 必填，默认由 SDK 解析官方 endpoint，MUST NOT 将自定义 endpoint / path-style 作为主配置项）
- `linapro-storage-azure`：Azure Blob Storage（厂商插件；配置面向 Azure account/container/shared key，MUST NOT 将 Azure 并入 S3 协议插件）
- `linapro-storage-s3`：S3 兼容协议插件（MinIO、R2、Ceph RGW 等；MUST 要求自定义 endpoint，MUST 支持 path-style；region 可选，缺省时使用签名占位默认值）

上述插件 MUST 为 `type: source`、`scope_nature: platform_only`、平台级全局安装模式，MUST NOT 将云 SDK 依赖引入 `lina-core` 宿主模块路径。

#### Scenario: 通用 S3 协议插件要求 endpoint

- **WHEN** 管理员配置 `linapro-storage-s3`
- **THEN** 设置 MUST 包含 endpoint 与 bucket 等协议字段
- **AND** region 可为可选；留空时运行时 MUST 使用默认签名 region（如 `us-east-1`）
- **AND** MUST NOT 将菜单与展示名绑定为「AWS S3」

#### Scenario: AWS 厂商插件使用 AWS 心智配置

- **WHEN** 管理员配置 `linapro-storage-aws`
- **THEN** 设置 MUST 包含 access key、secret、region、bucket
- **AND** 管理面 MUST NOT 以主表单暴露自定义 endpoint / force path-style（第一期）

#### Scenario: 华为云 OBS 厂商插件配置

- **WHEN** 管理员配置 `linapro-storage-obs`
- **THEN** 设置 MUST 包含 access key、secret、region、bucket
- **AND** endpoint 可选；留空时运行时 MUST 使用华为云默认 OBS endpoint 规则

#### Scenario: 七牛云 Kodo 厂商插件配置

- **WHEN** 管理员配置 `linapro-storage-qiniu`
- **THEN** 设置 MUST 包含 access key、secret、bucket
- **AND** region 可为可选；留空时运行时 MUST 根据 AccessKey 与 bucket 自动探测机房
- **AND** endpoint 可为可选自定义下载域名

#### Scenario: Azure Blob 厂商插件配置

- **WHEN** 管理员配置 `linapro-storage-azure`
- **THEN** 设置 MUST 包含 account name、account key、container
- **AND** endpoint 可选；留空时运行时 MUST 使用 `https://{account}.blob.core.windows.net/`
- **AND** MUST NOT 要求 S3 语义字段（region 作为主必填项、force path-style）

#### Scenario: 仅启用 COS 插件时使用 COS 后端

- **WHEN** `linapro-storage-cos` 已安装且为唯一可服务 storage provider 插件
- **AND** 业务插件调用 `storagecap.Service.Put` 写入对象
- **THEN** 对象内容 MUST 写入 COS 配置中的目标桶
- **AND** 调用方 MUST 仍只使用 logical path，不得感知 bucket 或云侧 object key 拼接细节

#### Scenario: 未启用任何云插件时回退本地

- **WHEN** 全部云 storage provider 插件均未处于可服务状态
- **THEN** `storagecap.Service` MUST 使用宿主内置 local provider
- **AND** 不得因未安装云插件而拒绝 Storage 调用

#### Scenario: 同时启用多个云插件时冲突

- **WHEN** 两个或以上云 storage provider 插件同时可服务
- **THEN** `storagecap.Service` 对象操作 MUST 失败并返回 `CodeStorageProviderConflict`（或等价稳定业务码）
- **AND** MUST NOT 静默选择其中一个 provider
- **AND** MUST NOT 静默回退 local

### Requirement: Provider 必须实现完整对象操作契约

每个云 storage provider 插件 MUST 实现 `storagecap.Provider` 的完整方法集，包括 `Put`、`Get`、`Delete`、`DeleteMany`、`List`、`ListCursor`、`Stat` 和 `BatchStat`。Provider MUST 只接收并解释 scoped object key，MUST NOT 接收或解析动态插件 `hostServices` 授权快照、插件 logical path 或租户/插件隔离规则（隔离由宿主 adapter 负责）。

#### Scenario: 批量元数据与游标列表

- **WHEN** 宿主 adapter 调用 provider 的 `BatchStat` 或 `ListCursor`
- **THEN** 实现 MUST 返回契约约定的 found/missing 与 next cursor 语义
- **AND** MUST NOT 对桶执行无界全量遍历作为生产路径

#### Scenario: 覆盖与存在冲突

- **WHEN** 调用方以 `Overwrite=false` 写入已存在对象 key
- **THEN** provider MUST 映射为 `storagecap` 对象已存在错误（或等价稳定码）
- **AND** MUST NOT 静默覆盖

### Requirement: 云存储插件必须提供管理配置页与 settings API

每个云 storage provider 插件 SHALL 提供管理后台配置页面与受权限保护的 settings API（查询与保存）。配置页 MUST 挂载到宿主 `setting` 目录（`parent_key: setting`）。页面布局 MUST 与授权登录类设置页保持一致：外层卡片、顶部说明 Alert、水平表单、密钥掩码与「已配置可留空保持」语义。

#### Scenario: 保存并掩码回显密钥

- **WHEN** 管理员在配置页填写访问密钥并保存
- **THEN** 系统 MUST 将密钥持久化到平台级 `sys_config`（或宿主 SysConfig 缝）
- **AND** 再次查询 settings 时密钥 MUST 仅以掩码形式返回
- **AND** 响应 MUST 标明密钥已配置

#### Scenario: 空密钥保持原值

- **WHEN** 密钥已配置
- **AND** 管理员保存时提交空密钥或掩码占位
- **THEN** 系统 MUST 保持原密钥不变

#### Scenario: 权限控制

- **WHEN** 用户不具备对应插件的 settings 查看或更新权限
- **THEN** 系统 MUST 拒绝 settings 查询或保存
- **AND** 对应用户 MUST NOT 在导航中看到无权限的配置菜单（或进入后被拒绝）

### Requirement: 配置页必须说明唯一启用与 fail-closed 语义

每个云存储配置页 MUST 向管理员展示明确说明：全局 Storage 后端同时只允许一个 storage provider 插件可服务；多开将导致 Storage 调用冲突失败；插件已启用但配置不完整时 MUST NOT 静默回退本地存储。

#### Scenario: 页面展示运维提示

- **WHEN** 管理员打开任一云存储配置页
- **THEN** 页面 MUST 显示关于唯一启用与配置不完整 fail-closed 的说明信息

### Requirement: 云存储插件必须支持连通性探测

每个云 storage provider 插件 SHALL 提供连通性探测能力（独立 API 或 settings 子动作），在不持久化脏数据的前提下验证当前表单或已存配置能否访问目标桶。探测失败 MUST 返回可读错误，MUST NOT 修改已持久化密钥以外的无关状态。

#### Scenario: 配置正确时探测成功

- **WHEN** 管理员提交该插件合法的必填配置（厂商插件含 region；协议插件含 endpoint）与密钥并触发测试连接
- **THEN** 系统 MUST 对目标桶执行只读探测并返回成功

#### Scenario: 测试连接失败以弹窗展示原因

- **WHEN** 管理员在任一云存储插件配置页点击测试连接且探测失败
- **THEN** 页面 MUST 以弹窗（或等价模态提示）展示完整可读失败原因，便于排查
- **AND** MUST NOT 仅依赖页面顶部常驻 Alert 区域展示该错误详情
- **AND** 成功路径的短 toast 提示 MAY 保留

#### Scenario: 配置错误时探测失败

- **WHEN** 管理员使用错误密钥或错误 bucket 触发测试连接
- **THEN** 系统 MUST 返回失败结果
- **AND** MUST NOT 将探测失败解释为静默成功

### Requirement: 配置不完整时对象操作必须失败且不回退本地

当某一云 storage provider 插件为唯一可服务 provider，但其运行时配置缺失或无效时，系统 SHALL 使 `storagecap.Service` 对象读写返回明确错误，MUST NOT 自动改用内置 local provider。

#### Scenario: 唯一云插件启用但未配置 bucket

- **WHEN** 仅该云插件可服务
- **AND** bucket 或密钥未配置
- **AND** 业务调用 `Storage().Put` 或 `Get`
- **THEN** 调用 MUST 失败并返回可诊断错误
- **AND** 对象 MUST NOT 被写入本地磁盘 provider

### Requirement: 云存储插件菜单图标必须归属插件侧

云存储插件菜单图标 MUST 优先使用 Iconify 集合图标；若需自定义品牌 SVG，MUST 放在插件 `frontend/icons` 并由工作台构建期注册，MUST NOT 将业务品牌 SVG 落入宿主 `packages/icons`。

#### Scenario: 自定义品牌图标不污染宿主 icon 包

- **WHEN** COS/OSS 等插件需要厂商品牌矢量图标且 Iconify 无合适图元
- **THEN** SVG MUST 位于对应插件 `frontend/icons`
- **AND** 侧栏图标标识 MUST 可在构建期解析
- **AND** 宿主 `apps/lina-vben/packages/icons` MUST NOT 新增该业务品牌 SVG
