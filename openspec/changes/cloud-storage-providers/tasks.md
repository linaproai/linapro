## 1. 宿主存储管理目录

- [x] 1.1 在宿主菜单 SQL 种子/幂等迁移中新增 `menu_key=storage`、`type=D` 一级目录（展示名存储管理），`sort=10`；将 `extension` 顺延为 `sort=11`、`developer` 顺延为 `sort=12`
- [x] 1.2 更新宿主 `manifest/i18n` 与 packed 双语 `menu.json`，增加 `storage` 标题（中文「存储管理」/ 英文 `Storage`）
- [x] 1.3 核对菜单投影空目录隐藏对 `storage` 生效；补充或更新相关单测/治理断言
- [x] 1.4 记录 i18n 影响（宿主菜单）；确认无数据权限/缓存新语义时在任务记录中写明无影响判断

## 2. 云存储插件脚手架（三家共用模式）

- [x] 2.1 新建 `linapro-storage-cos` source 插件骨架（plugin.yaml、embed、go.mod、Makefile、README 双语、i18n 骨架）
- [x] 2.2 新建 `linapro-storage-oss` source 插件骨架（同上）
- [x] 2.3 新建 `linapro-storage-s3` source 插件骨架（同上；声明支持自定义 endpoint / path-style）
- [x] 2.4 三插件 `plugin.yaml` 菜单：`parent_key: storage`，配置页 + settings-update 按钮权限；sort 分别为 10/20/30
- [x] 2.5 三插件在 `init` 中调用 `storagecap.Provide(pluginID, factory)` 注册（factory 可先返回未配置错误实现，保证可编译链接）

## 3. Settings 后端与配置持久化

- [x] 3.1 为 COS 实现 settings service（sys_config key、投影掩码、空密钥保持）与 GET/PUT API + 权限标签
- [x] 3.2 为 OSS 实现同等 settings 后端
- [x] 3.3 为 S3 实现同等 settings 后端（含 forcePathStyle 等 S3 字段）
- [x] 3.4 三插件平台级 sys_config 种子 SQL（tenant_id=0）及卸载清理策略
- [x] 3.5 实现连通性探测 API/动作（Head/List bucket）；失败不写脏配置；权限与 settings 对齐
- [x] 3.6 补充 settings 与掩码行为单测

## 4. Settings 前端页面

- [x] 4.1 COS 配置页：对齐授权登录子页布局（Card + Alert + 水平 Form + 密钥 InputPassword + 测试连接 + 保存）
- [x] 4.2 OSS 配置页：同上
- [x] 4.3 S3 配置页：同上（含 path-style / endpoint 字段）
- [x] 4.4 三插件页面 i18n（菜单、表单、Alert 唯一启用/fail-closed 说明、成功失败提示）

## 5. Provider 云 SDK 实现

- [x] 5.1 COS：实现完整 `storagecap.Provider`（Put/Get/Delete/DeleteMany/List/ListCursor/Stat/BatchStat），错误码映射
- [x] 5.2 OSS：实现完整 Provider
- [x] 5.3 S3：实现完整 Provider（自定义 endpoint + forcePathStyle）
- [x] 5.4 factory 从 settings 懒加载客户端；配置不全时操作失败且不回退 local
- [x] 5.5 Provider 单测（mock SDK 或可注入 fake）；覆盖覆盖冲突、missing、batch/cursor 语义
- [x] 5.6 记录 DI：settings/sys_config 与 Provider 依赖的 owner、创建位置、是否复用宿主启动共享实例

## 6. 运行时选择与冲突验证

- [x] 6.1 验证 0 云插件启用 → local；1 个启用且配置有效 → 云；≥2 启用 → conflict
- [x] 6.2 验证唯一云插件启用但配置无效 → 明确错误、不写 local
- [x] 6.3 用 demo 或最小调用路径冒烟 `Storage().Put/Get` 经云 provider（可用集成标签/本地 MinIO 测 S3）

## 7. E2E 与文档

- [x] 7.1 宿主/插件 E2E：未装云插件时无「存储管理」；安装至少一家后出现目录与子菜单（遵循 lina-e2e TC 编号约定，模块本地 TC001 起）
- [x] 7.2 至少一家插件 E2E：settings 保存、密钥掩码回显、无权限拒绝
- [x] 7.3 更新三插件 README 双语：唯一启用策略、配置项、与 Storage 领域关系、out of scope
- [x] 7.4 如需更新 `pkg/plugin` README 中 Storage provider 运维说明，按 documentation 规范同步中英

## 8. 校验与审查准备

- [x] 8.1 运行 `openspec validate cloud-storage-providers --strict`
- [x] 8.2 运行本变更相关 Go 编译/单测与新增 E2E（或记录环境阻断）
- [x] 8.3 影响分析复核：i18n、数据权限（平台 settings）、缓存（无新权威缓存）、开发工具跨平台（无则记录无影响）
- [x] 8.4 任务全部完成后触发 `lina-review` 再进入归档流程

## Feedback

- [x] **FB-1**: COS 菜单图标误用 QQ 企鹅，OSS 审查并改用阿里云品牌图标
- [x] **FB-2**: 存储管理目录排序调整到扩展中心之上
- [x] **FB-3**: 新增 linapro-storage-aws 厂商插件；将 linapro-storage-s3 改为通用 S3 协议插件（配置字段与产品语义分离）
- [x] **FB-4**: COS 菜单使用腾讯云品牌图标；S3 插件/菜单命名去掉「兼容」字样（对象存储-S3 / S3存储）
- [x] **FB-5 / P0**: 新增 `linapro-storage-obs`（华为云 OBS 厂商）与 `linapro-storage-azure`（Azure Blob 厂商）；补齐国内三强与非 S3 国际后端；配置页/Provider/单测/README/i18n 对齐既有云存储插件模式
- [x] **FB-6**: 连接测试失败时顶部 Toast 展示过长 SDK 错误；改为短 Toast + 页内可关闭/可复制 Alert 展示详情
- [x] **FB-7**: 连接测试失败 Alert 标题「连接测试失败」字号过大，改为与页面正文一致的字号
- [x] **FB-8**: 启用腾讯云 COS 后文件管理上传失败，报错「重置文件读取位置失败」
- [x] **FB-9**: 云存储插件品牌 SVG 不得落在宿主 `packages/icons`；改为插件 `frontend/icons` + 工作台构建期注册，Iconify 优先
- [x] **FB-10**: 多云冲突时文件管理「文件上传」因哈希复用跳过 Put 仍成功；应与图片上传一样 fail-closed

## 9. P0 扩展：华为 OBS 与 Azure Blob

- [x] 9.1 新建 `linapro-storage-obs` source 插件（骨架、settings、Provider、前端配置页、i18n、README）
- [x] 9.2 新建 `linapro-storage-azure` source 插件（Account/Container 配置模型、Provider、前端配置页、i18n、README）
- [x] 9.3 OBS 使用官方 `huaweicloud-sdk-go-obs`；Azure 使用官方 `azblob` Shared Key
- [x] 9.4 菜单 `parent_key: storage`，sort：COS 10 / OSS 20 / OBS 30 / 七牛 35 / AWS 40 / Azure 50 / S3 60；唯一启用与 fail-closed 文案
- [x] 9.5 更新 proposal/design/specs 与插件清单 README；运行插件单测

## 10. 七牛云 Kodo

- [x] 10.1 新建 `linapro-storage-qiniu` source 插件（settings、Provider、前端、i18n、README）
- [x] 10.2 使用官方 `qiniu/go-sdk/v7`；region 可选自动探测；endpoint 语义为下载域名
- [x] 10.3 菜单 sort=35（华为 OBS 之后、AWS 之前）；运行单测

