## Context

宿主已具备完整对象存储领域分层：

- 插件可见契约：`storagecap.Service`（logical path + 插件/租户作用域）
- 后端扩展契约：`storagecap.Provider` + `storagecap.Provide` + `ResolveProvider`
- 内置后端：local provider，委托宿主内部 `storage.Service`（namespace `plugins`）
- 文件中心：独立走 `storage.Service` namespace `files`，不经 `storagecap.Provider`

当前没有任何云厂商 provider 插件；规范中已预留「源码插件可注册 OSS/S3 等」，但未交付。

菜单侧：宿主在 `006-menu-role-management.sql` 维护稳定一级目录（`monitor`、`extension`、`content` 等），插件用 `parent_key` 语义挂载，**空目录由导航投影隐藏**。`linapro-extlogin-core` 的「授权登录」是 plugin-owned 域目录，不适合作为 Storage 的模板——Storage 的领域 owner 是宿主。

## Goals / Non-Goals

**Goals:**

- 交付 COS / OSS / OBS / 七牛 / AWS / Azure / S3 兼容等官方 source provider 插件，业务调用面仍仅 `Storage()`
- 宿主预留 `storage`（存储管理）稳定目录，云插件配置页挂载其下
- 配置页交互与密钥治理对齐授权登录子页
- 保持唯一 active provider 语义（启用状态驱动，不新增主配置 active 项）

**Non-Goals:**

- 文件中心 / `Files()` / `sys_file` 对象内容上云
- 预签名直传、CDN 公网 URL 进入 `storagecap` 契约
- 跨 provider 迁移工具
- 按租户选择不同云后端
- 新建 `linapro-storage-core` 或任何仅为菜单存在的领域壳插件
- 动态插件实现 Provider（SDK 与进程内注册要求 source）

## Decisions

### 1. 父目录归宿主，不设 storage-core

**决策**：在宿主菜单种子增加 `menu_key=storage`，`type=D`，名称「存储管理」；云插件 `parent_key: storage`。

**理由**：Storage 是 core-owned；与 `monitor` 挂载监控插件同构。避免卸载某一云插件带走父目录。

**替代**：`linapro-storage-core` 只建目录——过重且误导「领域引擎在插件」。某云插件兼 owner——卸载副作用不可接受。

### 2. 菜单 sort

**决策**：

| menu_key | sort |
|---|---|
| storage | 10 |
| extension | 11（由 10 顺延） |
| developer | 12（由 11 顺延） |

授权登录（插件 sort=9）保持不变，位于「存储管理」之前。侧栏顺序：授权登录 → **存储管理** → **扩展中心** → 开发中心。
### 3. 插件边界（厂商 + 协议）

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

### 4. Provider 选择语义（沿用现实现）

以代码 `storagecap.ResolveProvider` 为准：

- 0 个可服务已注册 provider 插件 → local
- 1 个 → 该插件
- ≥2 个 → `CodeStorageProviderConflict`，不静默挑选、不回退 local

「可服务」= 插件平台级 provider 启用（`IsProviderEnabled`），**不是**「配置字段已填完整」。  
插件已启用但配置无效时：仍为 active，对象操作返回明确错误，**不得**静默回退 local。配置页须用 Alert 说明。

历史规范中「主配置 active provider ID」与「唯一可服务自动选中」并存张力：实现与 README 已采用后者；本变更以自动唯一选中 + 冲突拒绝为准，不新增主配置项。

### 5. 配置持久化与管理面

对标 `linapro-oidc-google` settings：

- 平台级 `sys_config`（`tenant_id=0`），key 前缀 `plugin.<plugin-id>.*`
- GET 返回投影：密钥掩码 + `*Configured`；PUT 空密钥保持原值
- 权限：`linapro-storage-*:settings:view` / `settings:update`
- 页面：`Card` + `Alert` + 水平 `Form`（label≈180px），无重复 Card 标题
- 第一期提供「测试连接」（Head/List bucket 一类只读探测）

COS/OSS/OBS 公共字段：AccessKey/Secret、Region、Bucket、可选 Endpoint、可选 Path prefix。  
七牛 Kodo：AccessKey/Secret、Bucket 必填；Region 可选（`z0`/`z1`/`z2`/`cn-east-2`/`na0`/`as0`，留空自动探测）；Endpoint 可选（自定义下载域名）；可选 Path prefix。  
AWS 厂商：`accessKey` / `secret` / `region` / `bucket` / 可选 `pathPrefix`。  
Azure 厂商：`accountName` / `accountKey` / `container` / 可选 `endpoint`（默认 `https://{account}.blob.core.windows.net/`）/ 可选 `pathPrefix`。  
S3 协议：`accessKey` / `secret` / `endpoint` / `bucket` / `forcePathStyle` / 可选 `region`（默认签名占位）/ 可选 `pathPrefix`。

### 6. Provider 实现边界

- 只处理 adapter 传入的 scoped object key
- 实现完整 `Provider` 方法集（含 `ListCursor`、`BatchStat`、`DeleteMany`）
- 大文件：第一期依赖 SDK 对流式/分片上传的封装；不扩展 `storagecap` multipart 契约

### 6.1 插件菜单图标边界

- **优先** Iconify 集合图标（`simple-icons:*`、`carbon:*` 等），写在 `plugin.yaml`，不改宿主。
- Iconify 无合适图元时，自定义单色 SVG 放在插件 `frontend/icons/*.svg`；工作台构建期注册为 `svg:<plugin-id>-<stem>`。
- **禁止**把业务品牌 SVG 写入宿主 `apps/lina-vben/packages/icons/src/svg/icons/`。
- 侧栏菜单图标不走 `/x-assets`：x-assets 面向公开静态文件托管，带 version 与运行时 HTTP；菜单矢量 icon 需要启动期可用、可主题着色。
- 错误映射到 `storagecap` 稳定码（path invalid / exists / unavailable 等）

### 7. 与文件中心隔离

云 provider 仅承接 `storagecap` 路径（插件私有对象）。  
`file` 模块继续使用宿主 `storage.Service` 本地根；本变更不抽象 host-internal storage 为云后端。

### 8. 注册方式

**决策**：继续使用现有 `storagecap.Provide(pluginID, factory)` 包级注册（与当前正式扩展点一致）。

**不在本变更**引入 `pluginhost.Providers().ProvideStorage`，避免扩大宿主 provider facade 范围；若后续统一收口，另开变更。

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 管理员同时启用多个云插件导致全站 Storage 失败 | 配置页强提示；可选后续加 core 状态页；E2E/文档说明 |
| 启用但未配齐密钥，写入失败且不回退 local | Alert + 测试连接；错误码可读 |
| 切换 provider 后旧对象不可见 | 文档明确无自动迁移；第一期 out of scope |
| 三家 SDK 依赖体积与许可 | 仅插件 go.mod 引入；宿主不引用 |
| 列表/游标语义与本地实现细微差异 | 以 Provider 契约单测锁定分页与 missing 语义 |
| 宿主菜单 SQL 与已有库 sort 冲突 | 幂等 INSERT + 对 developer 的 sort 更新需幂等迁移策略 |

## Migration Plan

1. 升级宿主：执行菜单种子/迁移，写入 `storage` 目录，调整 `extension`/`developer` sort；已有环境无云插件时侧栏不出现空目录。
2. 安装并配置恰好一个云插件 → 启用 → `Storage()` 流量切到云。
3. 回滚：禁用/卸载云插件 → 自动回退 local（仅影响新读写；云上对象不会自动拉回本地）。
4. 无强制数据迁移步骤。

## Open Questions

- 测试连接是否需要独立 permission（默认复用 `settings:update` 或 `settings:view`）：建议第一期复用 `settings:view` 做探测、`settings:update` 做保存。
- 角色默认授权：超级管理员自动拥有新权限；是否给默认 admin 角色种子绑定三插件 settings——实现时按宿主插件菜单同步惯例处理。
