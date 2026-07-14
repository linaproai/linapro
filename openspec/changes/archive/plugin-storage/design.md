## Context

宿主对象存储与文件中心分层如下：

- 插件可见契约：`storagecap.Service`（logical path + 插件/租户作用域）
- 后端扩展契约：`storagecap.Provider` + `storagecap.Provide` + `ResolveProvider`
- 内置后端：local provider，委托宿主内部 `storage.Service`（namespace `plugins`）
- 文件中心：`Files()` / `sys_file`，走宿主 `storage.Service` namespace `files`，不经 `storagecap.Provider`

历史上存在三类缺口：

1. **领域语义不清**：`Storage()` 与 `Files()` 在 capability 面并存，示例与文档易把插件附件误接到文件中心或宿主本地路径。
2. **动态插件大对象写入脆弱**：guest 侧 `io.ReadAll` 后单次 `storage.put`，受 guest 内存与 envelope 约束。
3. **云后端未交付**：规范预留「源码插件可注册 OSS/S3 等」，但无官方 provider 插件与运维配置面。

菜单侧：宿主在菜单种子中维护稳定一级目录（`dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`），插件用 `parent_key` 语义挂载，空目录由导航投影隐藏。Storage 领域 owner 是宿主；配置入口应复用运维设置面，而不是仿照 plugin-owned 域目录（如「授权登录」）再建一层。

## Goals / Non-Goals

**Goals:**

- 固化 `Storage()` / `Files()` 职责边界，示例与 README 与契约一致
- 动态插件 `Storage().Put` 支持有界内存分片上传，路径授权与会话绑定完整
- 交付 COS / OSS / OBS / 七牛 / AWS / Azure / S3 兼容等官方 source provider 插件，业务调用面仍仅 `Storage()`
- 云插件配置页挂载到既有 `setting`（系统设置），交互与密钥治理对齐授权登录类设置页
- 保持唯一 active provider 语义（启用状态驱动，不新增主配置 active 项）

**Non-Goals:**

- 文件中心 / `Files()` / `sys_file` 对象内容上云
- 预签名直传、CDN 公网 URL 进入 `storagecap` 契约
- 跨 provider 迁移工具；按租户选择不同云后端
- 新建 `linapro-storage-core` 或仅为菜单存在的领域壳插件
- 独立 `storage` 一级宿主稳定目录
- 动态插件实现 Provider（SDK 与进程内注册要求 source）
- 重命名 `Storage()` / `Files()` 公开方法名

## Decisions

### 1. Storage 与 Files 边界

**决策**：不重命名公开入口。`Storage()` 拥有插件对象内容生命周期（保存、下载、替换、删除、列出、卸载清理）。`Files()` 仅做宿主文件中心投影与存在性不泄露校验。`Storage()` 不生成 `sys_file` 记录；`Files()` 不新增对象存储方法。`linapro-demo-source` 附件改走 `Storage()`。

### 2. 动态插件分片上传

**决策**：保留 `storage.put` 单次路径；新增 `put.init` / `put.chunk` / `put.commit` / `put.abort`。guest SDK 按输入大小与可知性自动选择。上传会话由 WASM storage host service 管理，commit 时通过 `storagecap.Service.Put` 流式写入最终 logical path。分片各步授权与 `storage.put` 等价，path 匹配最终插件 logical path；upload ID 绑定插件 ID、logical path 与会话状态。不对最终对象大小设动态 host service 固定上限。

### 3. Provider 选择：唯一可服务自动选中

以 `storagecap.ResolveProvider` 为准：

| 可服务已注册 provider 插件数 | 行为 |
|---|---|
| 0 | 内置 local |
| 1 | 该插件 |
| ≥2 | `CodeStorageProviderConflict`，不静默挑选、不回退 local |

「可服务」= 插件平台级 provider 启用（`IsProviderEnabled`），**不是**「配置字段已填完整」。  
插件已启用但配置无效时：仍为 active，对象操作返回明确错误，**不得**静默回退 local。配置页须用 Alert 说明。

不再通过宿主主配置 active provider plugin ID 选择后端；运维通过插件管理启停恰好一个云插件选择后端。

### 4. 配置入口挂系统设置，不设 storage 一级目录 / storage-core

**决策**：云插件 `parent_key: setting`；不新增 `menu_key=storage` 一级目录，也不引入 `linapro-storage-core`。

**理由**：云存储配置属于运维设置面，与字典、参数、文件管理等同属「系统设置」语义；避免一级导航膨胀。Storage 领域能力仍是 core-owned，仅配置入口复用既有宿主目录。

一级侧栏 sort：`extension=10`、`developer=11`；`sort=9` 留给授权登录等域目录。云存储配置子菜单在「系统设置」下相对顺序（插件 `sort`）：COS 10 / OSS 20 / OBS 30 / 七牛 35 / AWS 40 / Azure 50 / S3 60。

### 5. 插件边界（厂商 + 协议）

| 插件 ID | 后端 |
|---|---|
| `linapro-storage-cos` | 腾讯云 COS 官方 SDK（厂商） |
| `linapro-storage-oss` | 阿里云 OSS 官方 SDK（厂商） |
| `linapro-storage-obs` | 华为云 OBS 官方 SDK（厂商） |
| `linapro-storage-qiniu` | 七牛云 Kodo 官方 SDK（厂商：region 可选自动探测；endpoint 语义为可选下载域名） |
| `linapro-storage-aws` | AWS SDK for Go v2，官方 S3（厂商：region 必填；主表单不暴露 endpoint/path-style） |
| `linapro-storage-azure` | Azure SDK for Go，Blob Storage（厂商：account + container + shared key；可选自定义 endpoint） |
| `linapro-storage-s3` | AWS SDK for Go v2 作协议客户端（endpoint 必填、path-style；region 可选默认 `us-east-1`） |

均为 `type: source`、`distribution: managed`、`scope_nature: platform_only`、`default_install_mode: global`。无插件间依赖。AWS 与 S3 兼容可共用 SDK，配置面必须分离。Azure Blob 非 S3 API，不得并入 S3 协议插件。

### 6. 配置持久化与管理面

对标 OIDC/授权登录类 settings：

- 平台级 `sys_config`（`tenant_id=0`），key 前缀 `plugin.<plugin-id>.*`
- GET 返回投影：密钥掩码 + `*Configured`；PUT 空密钥保持原值
- 权限：`linapro-storage-*:settings:view` / `settings:update`
- 页面：`Card` + `Alert` + 水平 `Form`（label≈180px），无重复 Card 标题
- 连通性探测：Head/List bucket 一类只读探测；失败以 Modal.error（或等价弹窗）展示完整原因，不单靠页顶常驻 Alert；失败不写脏配置
- 探测权限第一期复用 `settings:view`，保存用 `settings:update`

字段要点：

- COS/OSS/OBS：AccessKey/Secret、Region、Bucket、可选 Endpoint、可选 Path prefix
- 七牛 Kodo：AccessKey/Secret、Bucket 必填；Region 可选（留空自动探测）；Endpoint 可选（下载域名）；可选 Path prefix
- AWS 厂商：`accessKey` / `secret` / `region` / `bucket` / 可选 `pathPrefix`
- Azure 厂商：`accountName` / `accountKey` / `container` / 可选 `endpoint` / 可选 `pathPrefix`
- S3 协议：`accessKey` / `secret` / `endpoint` / `bucket` / `forcePathStyle` / 可选 `region` / 可选 `pathPrefix`

### 7. Provider 实现边界

- 只处理 adapter 传入的 scoped object key
- 实现完整 `Provider` 方法集（含 `ListCursor`、`BatchStat`、`DeleteMany`）
- 错误映射到 `storagecap` 稳定码（path invalid / exists / unavailable 等）
- 大文件：第一期依赖 SDK 对流式/分片上传的封装；不扩展 `storagecap` multipart 契约
- 注册方式：`storagecap.Provide(pluginID, factory)` 包级注册；不在此引入 `pluginhost.Providers().ProvideStorage`

### 8. 插件菜单图标边界

- **优先** Iconify 集合图标（`simple-icons:*`、`carbon:*` 等），写在 `plugin.yaml`
- 无合适图元时，自定义单色 SVG 放在插件 `frontend/icons/*.svg`；工作台构建期注册为 `svg:<plugin-id>-<stem>`
- **禁止**把业务品牌 SVG 写入宿主 `apps/lina-vben/packages/icons`
- 侧栏菜单图标不走 `/x-assets`（x-assets 面向公开静态托管；菜单矢量需启动期可用、可主题着色）

### 9. 与文件中心隔离

云 provider 仅承接 `storagecap` 路径（插件私有对象）。  
`file` 模块继续使用宿主 `storage.Service` 本地根；不将 host-internal storage 抽象为云后端。

业务侧上传（文件管理/图片等）在解析 active Storage provider 冲突或不可用时必须 fail-closed：不得因哈希复用跳过 Put 而表现为成功。

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 管理员同时启用多个云插件导致全站 Storage 失败 | 配置页强提示；文档与 E2E 说明 |
| 启用但未配齐密钥，写入失败且不回退 local | Alert + 测试连接；错误码可读 |
| 切换 provider 后旧对象不可见 | 文档明确无自动迁移 |
| 多云 SDK 依赖体积与许可 | 仅插件 go.mod 引入；宿主不引用 |
| 列表/游标语义与本地实现细微差异 | Provider 契约单测锁定分页与 missing 语义 |
| 动态插件分片会话泄漏 | abort/超时清理；会话绑定插件与 path |

## Migration Plan

1. 边界与分片能力随宿主发布；示例插件改用 `Storage()`。
2. 升级宿主：移除历史 `storage` 一级目录（若存在）；云插件 `parent_key: setting`。
3. 安装并配置恰好一个云插件 → 启用 → `Storage()` 流量切到云。
4. 回滚：禁用/卸载云插件 → 自动回退 local（仅影响新读写；云上对象不会自动拉回本地）。
5. 无强制数据迁移步骤。

## Open Questions（已收敛默认）

- 测试连接权限：第一期探测复用 `settings:view`，保存用 `settings:update`。
- 角色默认授权：超级管理员自动拥有新权限；默认 admin 角色按宿主插件菜单同步惯例处理。
