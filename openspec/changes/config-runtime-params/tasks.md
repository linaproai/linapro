## 1. OpenSpec 文档与规范治理

- [x] 1.1 创建 `config-runtime-params` active change，并补齐 `proposal.md`、`design.md`、`tasks.md`
- [x] 1.2 为 `config-management`、`user-auth`、`online-user`、`user-management` 编写本次变更的 delta specs
- [x] 1.3 回收此前提前写入 `openspec/specs/` 基线规范的内容，恢复“active change 先行、归档再同步基线”的流程

## 2. 内置运行时参数契约与元数据

- [x] 2.1 梳理当前宿主已真实接入且适合运行时生效的 5 项内置参数，并在配置服务中建立统一注册表与校验规则
- [x] 2.2 在 `007-config-management.sql` 中为 5 项内置运行时参数补齐名称、默认值、备注说明并支持 upsert
- [x] 2.3 在参数管理服务中实现内置参数的创建/更新/导入值校验，以及防改键名、防删除保护

## 3. 宿主运行时行为接入

- [x] 3.1 让 `sys.jwt.expire`、`sys.session.timeout`、`sys.upload.maxSize`、`sys.login.blackIPList` 通过宿主配置服务参与运行时读取
- [x] 3.2 在登录与鉴权链路中接入登录 IP 黑名单和实时会话超时校验，确保参数改动能够即时生效
- [x] 3.3 在用户重置密码弹窗中通过参数查询接口读取 `sys.user.initPassword` 作为默认回填值

## 4. 测试与审查

- [x] 4.1 补充 Go 单测，覆盖运行时参数格式校验、JWT/会话/上传配置覆盖、登录黑名单拦截、会话超时校验与内置参数保护
- [x] 4.2 更新或新增 E2E 用例：`TC0010-user-reset-pwd.ts`、`TC0049-config-crud.ts`、`TC0079-config-runtime-params.ts`
- [x] 4.3 运行相关验证命令（Go tests、ESLint、Playwright），并根据设计与实现完成一次遗漏审查

## Feedback

- [x] **FB-1**: 为内置运行时参数补充多实例部署下的本地快照缓存与共享 revision 同步机制，避免热点链路每次读取都直查 `sys_config`
- [x] **FB-2**: 为运行时参数缓存同步补充 Go 单测，覆盖缓存命中、revision 变更后重载，以及 `sysconfig` 更新/导入触发缓存刷新
- [x] **FB-3**: 为运行时参数缓存相关关键逻辑补充维护性注释，说明本地缓存、自愈清理与后台 watcher 的协作关系
- [x] **FB-4**: 优化 JWT 配置读取链路，避免鉴权热点路径每次请求都重复构造配置对象并读取静态配置
- [x] **FB-5**: 为静态配置缓存实现补充源码注释，说明 once 初始化、缓存重置与返回副本的维护意图
- [x] **FB-6**: 审查配置服务热点调用链路，继续优化登录黑名单与上传配置读取，避免高频请求重复构造配置对象或重复解析规则
- [x] **FB-7**: 基于“全新项目无历史债务”的约束，审查运行时参数 SQL 版本拆分策略，将独立 `014` 种子数据并回合适的历史 SQL 文件
- [x] **FB-8**: 审查 `cloneUserAccessContext` 的切片克隆实现，评估并优化热点缓存复制时的内存分配方式
- [x] **FB-9**: 审查 `internal/service` 下所有组件的 `Service` 接口方法注释完整性，并增加守卫测试保证后续方法缺少注释时会失败
- [x] **FB-10**: 为项目 SQL 管理规范补充“脚本必须满足幂等执行”要求，明确建表、改表、种子数据写入等语句的安全重复执行约束
- [x] **FB-11**: 审查并增强 `openspec-review` 技能，使其显式包含 SQL 文件变更时的规范审查要求与报告项
- [x] **FB-12**: 批量审查并修复 `apps/lina-core/manifest/sql/*.sql` 现有脚本的幂等性问题，确保根目录版本 SQL 均可安全重复执行
- [x] **FB-13**: 下线 `sys.user.initPassword` 内置参数，同步移除重置密码弹窗默认回填、SQL 种子数据与相关测试覆盖
- [x] **FB-14**: 将品牌、登录页展示和后台主题/布局相关的安全前端参数纳入受保护系统参数清单，并补齐 SQL 种子数据与格式校验
- [x] **FB-15**: 提供公开白名单前端配置接口，并让登录页、浏览器标题及后台 preferences 在启动阶段消费这些系统参数
- [x] **FB-16**: 在参数设置保存/导入后即时刷新公开前端配置，并补充 Go 单测与 Playwright E2E 覆盖登录页展示和主题切换行为
- [x] **FB-17**: 修复浏览器本地 preferences 缓存覆盖 `sys.ui.theme.mode` 导致后台主题在当前浏览器不生效的问题
- [x] **FB-18**: 下线 `sys.ui.theme.primaryColor` 公开前端参数，回收后端契约、前端消费与测试链路以简化系统管理复杂度
- [x] **FB-19**: 为插件管理页增加“详情”按钮和详情弹窗，展示插件基础治理信息与宿主服务明细
- [x] **FB-20**: 新增 `hack/tests/e2e/plugin/TC0078-plugin-detail-dialog.ts`，覆盖详情弹窗展示与空状态提示
- [x] **FB-21**: 收敛 OpenSpec 文档语言治理：新建迭代文档跟随用户上下文语言生成，归档文档与同步主规范统一使用英文，并通过 `openspec/config.yaml` 与项目规范落地
- [x] **FB-22**: 为 `lina-core` 增加结构化日志开关配置，启用后使用 GoFrame 官方 `HandlerJson` 输出单行 JSON 日志
- [x] **FB-23**: 收敛 `ghttp.Server` 日志与业务日志到同一日志出口，文件输出时统一写入相同日志文件
- [x] **FB-24**: 修复页面刷新时启动 Loading 画面中的 `LinaPro` 品牌文字字体闪变问题，确保启动阶段与应用渲染后的字体保持一致
