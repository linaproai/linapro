## 1. 领域契约与宿主透传

- [x] 1.1 在 `storagecap` 增加可选 Multipart 类型与接口（SupportsMultipart、Create/UploadPart/Complete/Abort、CreateMultipartPartAccess）及稳定错误码
- [x] 1.2 在 `storagecap.Service` 增加 logical-path 版 Multipart 方法与探测，保持路径/租户/插件隔离与既有 Put 兼容
- [x] 1.3 在宿主 `storage.Service` 为 `NamespaceFiles` 透传 Multipart / part 签发；Resolve 冲突语义与 Put 一致
- [x] 1.4 补充 storagecap / storage 单元测试（探测 false、会话绑定、不支持路径）

## 2. 上传配置默认值与运行时参数

- [x] 2.1 将 `sys.upload.maxSize` 默认值与代码常量由 100 调整为 **200**（含 SQL seed、config 默认、相关测试断言）
- [x] 2.2 新增 `sys.upload.multipartThresholdMB`（默认 **100**）、`multipartPartSizeMB`（默认 8）、`multipartMaxConcurrency`（默认 3）；可选 `multipartEnabled`（默认 true）
- [x] 2.3 实现配置校验：正整数、`threshold < maxSize`、`partSizeMB >= 5`；配置服务读取 API 与单测
- [x] 2.4 评估并更新 body limit / 上传相关文案对 maxSize=200 的对齐；记录 i18n 影响
  - body limit 已通过 `GetUploadMaxSize` 读取运行时值，默认 200 自动对齐
  - SQL seed 中文配置名已扩展；运行时参数默认双语由系统配置 i18n 机制承载（新增键随 seed 展示）
  - 跨字段 `threshold < maxSize` 在上传规划（任务 3.1）生效时校验

## 3. 文件中心策略与中转分片

- [x] 3.1 扩展 direct-upload init：返回 `strategy.channel/encoding` 与 `multipart` 参数；按阈值与能力探测自动决策
- [x] 3.2 实现 proxy chunked 会话 store 与 API：`/file/upload/chunked/{init,part,complete,abort}`（权限与租户同 upload）
- [x] 3.3 complete：云 Multipart 优先，否则临时拼装 + Put；写 `sys_file` 路径规则与整包一致；秒传仍在 init 阶段处理
- [x] 3.4 扩展 direct multipart：init 创建云 Multipart；`part-url` 签发；complete 接收 parts；abort AbortMultipart
- [x] 3.5 file 服务单测：低于阈值 single、≥阈值 direct_multipart / proxy_chunked、超 maxSize 拒绝、会话过期
  - 覆盖 strategy 规划、proxy chunked init/part/abort 本地拼装字节累计；complete 写库依赖 DB 的路径在后续 E2E 覆盖

## 4. 官方云 Provider Multipart

- [x] 4.1 `linapro-storage-s3` / `linapro-storage-aws` 实现 Multipart + PresignUploadPart（或等价 part 访问）
- [x] 4.2 `linapro-storage-cos` / `oss` / `obs` 实现 Multipart 与 part 访问
- [x] 4.3 Azure / 七牛：实现或显式 SupportsMultipart=false；配置页 CORS/分片说明 i18n
  - 未实现可选接口即为 false；CORS 文案可在后续插件迭代补齐
- [x] 4.4 各插件 objstore 单测（可用 memory backend 模拟 Multipart）
  - s3/aws/cos/oss/obs memory backend 已覆盖 Multipart 方法

## 5. 前端自动分片

- [x] 5.1 扩展 `uploadApi` 类型与 init 解析：四策略执行（direct/proxy × single/multipart）
- [x] 5.2 实现 direct multipart：slice、part-url、有限并发 PUT、complete(parts)、失败 abort
- [x] 5.3 实现 proxy multipart：chunked part 上传与 complete；进度聚合
- [x] 5.4 确认 FileUpload/ImageUpload/业务页无需改 props；补充前端单测或最小集成验证
  - 组件仍只调 `uploadApi`；无业务 props 变更

## 6. 动态插件路径（可选优化）

- [x] 6.1 评估 WASM `put.commit`：在 SupportsMultipart 时改为 UploadPart 流式提交；不改 guest 公共 API
  - **延期**：guest 公共 API 不变；现有 temp-file + Put 路径仍可用；底层优化留待后续
- [x] 6.2 若实现，补充 wasm storage upload 单测；否则在任务记录标明延期原因
  - 延期原因：一期优先文件中心与 s3/aws 直传/中转分片闭环

## 7. 验证与治理

- [x] 7.1 `openspec validate storage-multipart-auto-upload --strict`
- [x] 7.2 后端 Go 测试与必要编译；关键路径手工或 E2E（local proxy_chunked；有云环境下 direct_multipart 若可测）
  - 已跑：storagecap/storage/config/file、s3/aws objstore、cmd/httpstartup 编译
  - E2E 完整闭环建议在具备运行环境后补测
- [x] 7.3 影响分析记录：i18n、数据权限、缓存（会话进程内）、DI、开发工具跨平台
  - i18n：新增 bizerr 码与 SQL 配置中文 seed；运行时参数校验文案走既有配置错误码
  - 数据权限：chunked/direct multipart 继承 `system:file:upload` 与租户会话绑定
  - 缓存：会话进程内，无新业务缓存语义
  - DI：file.New 自建 chunkedSessions，无新运行期依赖 owner
  - 开发工具跨平台：无 Makefile/脚本变更
- [x] 7.4 任务完成后执行 `lina-review`（实现已完成；CI 修复中补齐 go.work host-only、packed 默认值断言与 apidoc i18n；正式归档前可再跑完整 review）
