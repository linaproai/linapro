# Tasks

## Summary

- [x] `sys_file` 模型与本地 Storage：去重、`scene` 收敛、路径简化与历史 `t/...` 兼容
- [x] 文件管理 API/前端：上传列表下载删除、场景与后缀、通用上传组件、公告/头像接入
- [x] 统一对象后端：`ResolveProvider`、`files/` 键、本地 NamespaceFiles 路由、云未命中回退、`engine` 写入与 DI
- [x] 客户端直连：`storagecap` DirectAccess 契约与稳定码；宿主 resolving 探测与 local 降级；文件中心 init/complete/abort 会话与 API；官方七云 Provider 直连；前端 uploadApi 双模式
- [x] 对象 Multipart：`storagecap` 可选 Multipart 会话与探测；宿主 `NamespaceFiles` 透传；proxy chunked API；direct multipart part-url + complete(parts)；自动策略 channel×encoding；`maxSize` 默认 200、`multipartThresholdMB` 默认 100 等配置校验
- [x] 官方云 Multipart：s3/aws/cos/oss/obs memory backend 与实现覆盖；未实现接口即 SupportsMultipart=false
- [x] 前端四策略：uploadApi 自动 direct/proxy × single/multipart；业务组件 props 不变；直传失败不静默降级
- [x] 动态插件 WASM put.commit 走 Provider Multipart：**延期**（guest API 不变，temp-file+Put 仍可用；一期优先文件中心与 s3/aws 闭环）
- [x] 验证：宿主/插件 Go 单测（storagecap/storage/config/file、s3/aws/cos/oss/obs objstore）、cmd/httpstartup 编译、local 环境 multipart/E2E 关键路径、`openspec validate --strict`、`lina-review`（CI 补齐 go.work host-only、packed 默认值断言与 apidoc i18n）
- [x] 治理：i18n（API apidoc、bizerr、SQL 配置 seed 与插件 CORS/分片文案双语）；数据权限（init/complete/chunked 继承 `system:file:upload` 与租户会话绑定）；缓存（会话进程内，无新业务缓存语义）；DI（`file.New` 自建会话 store，无新运行期依赖 owner；storage resolving 复用启动期 runtime）；跨平台开发工具无影响；第一期不扩展 `filecap`
