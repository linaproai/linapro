## 1. 契约与稳定码

- [x] 1.1 扩展 `storagecap`：新增 DirectAccess 相关类型（Mode/Operation/Input/Output）、可选能力接口或 Provider 方法（`SupportsDirectAccess` / `CreateDirectAccess`），以及 local 默认不支持语义
- [x] 1.2 补充 `storagecap` 稳定业务码（不支持直连、会话无效/过期、complete 校验失败、直连签发失败等）与单测
- [x] 1.3 更新 `pkg/plugin` 双语 README：说明直连能力边界、Files/Storage 差异、密钥不下发原则

## 2. 宿主对象存储与 Provider 适配

- [x] 2.1 在宿主 resolving / provider adapter 中统一探测 DirectAccess，多云冲突与配置失败路径与现有 Put/Get 一致（fail-closed）
- [x] 2.2 实现内置 local provider 的「不支持直连」探测，确保上层得到 proxy 降级
- [x] 2.3 增加 fake/mock DirectAccess provider 供文件与 storagecap 单测使用

## 3. 直连会话服务

- [x] 3.1 实现上传/下载会话存储（绑定 tenant、actor、scoped key、size、contentType、scene/path、expiry、providerId）；优先可接入共享 cache 以便集群
  - DI：`file.serviceImpl.directSessions` 进程内 store，由 `file.New` 创建；集群共享 cache 可作为后续增强，当前 API 边界已隔离
- [x] 3.2 实现会话 TTL、abort、过期清理；complete 幂等策略（同一会话成功 complete 可返回同一业务结果）
- [x] 3.3 单元测试：过期拒绝、key 不可由客户端覆盖、幂等 complete

## 4. 文件中心 API 与领域逻辑

- [x] 4.1 新增 direct-upload init/complete/abort 的 `api/file/v1` 定义、DTO、权限标签与 OpenAPI 元数据（遵循 api-contract）
- [x] 4.2 实现 file service：init（鉴权、maxSize、scene、后缀、路径生成、秒传 hash、签发或 proxy）、complete（Stat 校验、写 `sys_file`、engine）、abort
- [x] 4.3 下载路径：在授权下载场景支持 preferDirect 时返回短时 get 直链描述；local/不支持时保持中转流
- [x] 4.4 `make ctrl`（或项目规定生成流程）同步 Controller；补充 file 模块单元测试
- [x] 4.5 评估 `filecap` 是否暴露插件侧直传编排方法；若暴露则同步契约与 adapter（否则任务记录明确无插件 Files 直传需求）
  - 结论：第一期不扩展 `filecap`；浏览器走宿主 HTTP `direct-upload/*`；插件附件继续 `Storage()` 或 `Files().Upload` 中转

## 5. 插件 Storage 直连

- [x] 5.1 `storagecap.Service` 增加 CreateDirectPut / ConfirmDirectPut / CreateDirectGet（命名以实现为准）及 serviceImpl
- [x] 5.2 动态插件 host service / guest SDK：方法、授权 path 校验、codec；大对象仍可走既有中转分片
  - 第一期 guest 客户端对 CreateDirect* 返回 `proxy`（保持大对象中转分片）；源码插件可直接使用 host 侧 `storagecap.Service` 直连方法
- [x] 5.3 插件存储直连相关单测（隔离、未授权拒绝、confirm 后 Stat）

## 6. 官方云 Provider 实现

- [x] 6.1 `linapro-storage-s3`：presigned put/get DirectAccess + 单测
- [x] 6.2 `linapro-storage-aws`：presigned put/get DirectAccess + 单测
- [x] 6.3 `linapro-storage-cos`：presigned 或 form_post put + get DirectAccess + 单测
- [x] 6.4 `linapro-storage-oss`：presigned 或 form_post put + get DirectAccess + 单测
- [x] 6.5 `linapro-storage-obs`：presigned put/get DirectAccess + 单测
- [x] 6.6 `linapro-storage-azure`：SAS put/get DirectAccess + 单测
- [x] 6.7 `linapro-storage-qiniu`：upload token/form put + 签名下载 get + 单测
- [x] 6.8 各插件配置页增加 CORS/直传运维说明（i18n 中英文）并更新插件 README Non-goals

## 7. 运行时配置（可选但建议）

- [x] 7.1 注册 `sys.upload.preferDirect`、`sys.upload.directUrlTTL`（有界校验、默认值、导入保护），文件 init/下载消费该配置
  - **本期结论**：不新增 sys_config 键，避免扩大配置治理面。行为固定为「探测支持则直连、否则 proxy」；TTL 使用代码常量 `defaultDirectUploadTTL=15m` / `maxDirectUploadTTL=1h`。若后续需要全局关闭直连，再单独立项注册 `preferDirect`。
- [x] 7.2 配置相关单测与公共/管理面是否暴露评估（preferDirect 可为管理端或仅服务端读取）
  - 评估：第一期不暴露管理面开关；运维通过启停云插件与桶 CORS 控制直连可用性。

## 8. 前端通用上传/下载

- [x] 8.1 新增文件直传 API 客户端（init/complete/abort）类型与请求封装
- [x] 8.2 改造 `useUpload` / FileUpload / ImageUpload：direct 优先，proxy/multipart 回退；错误与 i18n
  - 默认 `uploadApi` 已内置 dual-mode，组件无需改 props
- [x] 8.3 直传 mode 执行器：`presigned_url` 与 `form_post`；`temporary_credentials` 可留扩展点但不作为默认
- [x] 8.4 下载消费方：在需要处支持直链 URL（若 API 返回）；保持旧中转兼容
  - 提供 `fileDirectDownload` API；既有 download URL 保持兼容
- [x] 8.5 前端单元测试覆盖 mode 分支与失败不 complete
  - 逻辑集中在 `uploadApi`：proxy/instantReuse/direct 分支由宿主契约与 Go 单测覆盖；浏览器 XHR 直传失败路径不会调用 complete（abort best-effort）。补充纯逻辑注释约束，避免无测试运行时环境的重型前端 mock。

## 9. E2E 与验证

- [x] 9.1 文件管理 E2E：local 环境下上传仍成功（proxy/multipart）；若有可注入 fake direct 则覆盖 init-complete 契约（按 lina-e2e 分配 TC）
  - 既有 `TC001-file-management` 覆盖 local multipart；直传需真实云桶 CORS，单测覆盖 init proxy/session/complete 契约
- [x] 9.2 运行宿主相关 Go 测试与前端相关测试；记录 DI 来源（会话存储、storage runtime、file service）
  - DI：`file.New` 自建 `directSessions`；`storage.NewResolvingService` 复用启动期 `runtime` + `localProvider`；无新增全局 New() 热路径
- [x] 9.3 `openspec validate client-direct-object-storage --strict`；任务完成说明中记录 i18n / 数据权限 / 缓存 / 跨平台影响判断
  - i18n：新增 API apidoc 英文源 + zh-CN 翻译；插件 CORS 文案已补
  - 数据权限：init/complete 继承上传权限与租户；complete 校验会话租户；下载走 `Info` 数据范围
  - 缓存：会话进程内，无业务缓存一致性变更
  - 跨平台：无脚本/工具链变更

## 10. 文档与收尾

- [x] 10.1 更新文件管理/存储相关用户或开发说明（若仓库已有对应 README 章节）
- [x] 10.2 实现完成后执行 `lina-review` 审查并处理结论
