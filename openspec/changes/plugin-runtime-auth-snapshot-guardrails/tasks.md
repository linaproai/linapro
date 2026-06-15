## 1. 权限与会话边界确认

- [ ] 1.1 静态确认动态路由认证中所有角色 DAO、DO、Entity 直连点，并记录迁移目标 owner。
- [ ] 1.2 对齐现有`session.Store.TouchOrValidate`语义，确认本变更不新增 session 有效性缓存或写回节流机制。
- [ ] 1.3 记录影响分析：数据权限、缓存一致性、`i18n`、开发工具跨平台、测试策略、DI 来源和`apps/lina-core/pkg/plugin`README 是否需要同步。

## 2. Role 访问投影契约

- [ ] 2.1 在`role`模块发布动态路由访问投影窄契约，返回权限、角色名、数据范围、unsupported 标记和超管标记。
- [ ] 2.2 让访问投影复用 token access snapshot、`permission-access`修订号、租户维度和 fail-closed 策略。
- [ ] 2.3 将插件 runtime 构造函数改为显式注入访问投影契约，删除动态路由认证对角色治理表的直接访问。
- [ ] 2.4 补充权限命中、权限拒绝、租户隔离、权限拓扑变化、freshness 不可确认 fail-closed 和返回对象隔离测试。

## 3. 动态路由身份快照迁移

- [ ] 3.1 将动态路由身份快照构建改为消费`role`访问投影和共享`session.Store`。
- [ ] 3.2 验证登出、强制下线、token 撤销、session 过期、租户不匹配和 Redis hot state 失败时动态路由均拒绝执行。
- [ ] 3.3 验证身份快照中的数据范围和 host service context 与宿主受保护 API 同源一致。

## 4. Guest 执行护栏

- [ ] 4.1 增加 guest 全局并发、按插件并发、获取超时和内存页上限配置读取与校验。
- [ ] 4.2 在动态路由、生命周期、hook、cron discovery 和 cron job guest 执行入口接入同一护栏。
- [ ] 4.3 定义资源繁忙、资源耗尽和非法配置的稳定`bizerr`错误码、message key、英文 fallback 和必要`i18n`资源。
- [ ] 4.4 补充并发上限、按插件隔离、超时释放、panic/错误释放许可和非法配置测试。

## 5. Host call 与 datahost 微优化

- [ ] 5.1 在`ExecuteBridge`请求内构建一次 host service 授权快照并传递给 host call context，保持每次调用的 service/method/resource 校验。
- [ ] 5.2 为 datahost 表契约缓存设计按插件、表名和迁移状态的键，并在插件 SQL 生命周期成功提交后按插件失效。
- [ ] 5.3 补充授权收缩、系统型调用、DDL 后 schema 刷新、缓存不可用回源和数据权限边界测试。
- [ ] 5.4 如果 datahost 表契约缓存实施风险高于收益，保留 host call 授权快照优化并在任务记录中说明后移原因。

## 6. 验证与审查

- [ ] 6.1 运行覆盖`role`、`plugin/internal/runtime`、`wasm`、`datahost`和启动装配变更包的`go test <changed-package> -count=1`。
- [ ] 6.2 涉及构造函数、路由绑定或启动装配变更时，运行`cd apps/lina-core && go test ./internal/cmd -count=1`或等价启动绑定测试。
- [ ] 6.3 运行`openspec validate plugin-runtime-auth-snapshot-guardrails --strict`。
- [ ] 6.4 执行`lina-review`，审查数据权限等价性、缓存一致性七要素、DI 来源检查、错误本地化和测试覆盖。
