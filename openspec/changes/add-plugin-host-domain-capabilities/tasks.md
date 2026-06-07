## 1. 变更基线与影响记录

- [x] 1.1 重新读取本变更命中的项目规则和 OpenSpec 产物，确认本迭代不考虑旧生产接口兼容。
- [x] 1.2 扫描所有官方源码插件、动态插件示例、`pluginhost`、`pluginbridge`和`capability`包，列出宿主`sys_*`、宿主`DAO/DO/Entity`、旧领域接口、旧`host service`方法和动态`data`核心表授权的生产使用清单。
- [x] 1.3 为每一处直接宿主数据访问建立迁移映射，记录目标领域能力、领域 owner、调用方、数据权限边界、缓存影响和测试覆盖方式。
- [x] 1.4 在任务执行记录中完成影响分析，至少覆盖`i18n`、缓存一致性、数据权限、开发工具跨平台、测试策略、SQL/DAO、API 契约和 DI 来源；确认无影响的项目必须明确记录无影响判断。

## 2. 公共领域契约

- [x] 2.1 定义插件可见的`CapabilityContext`，覆盖`pluginID`、actor、tenant、调用来源、系统调用标识、授权快照、资源或投影标识和审计信息。
- [x] 2.2 定义领域命名`ID`类型和动态协议字符串编码/解码规则，避免在插件契约中暴露数据库自增主键类型。
- [x] 2.3 定义通用批量读取结果、`MissingIDs`语义、`Ensure*`命令前置校验语义和结构化业务错误边界。
- [x] 2.4 定义领域投影的稳定值、`labelKey`和可选`label`语义，明确 locale 来源和未启用`i18n`插件的消费边界。
- [x] 2.5 定义高频领域方法的分页、数量上限、字段投影、排序、过滤和批量装配公共约束。
- [x] 2.6 为公共契约补充 Go 注释、错误码和编译期接口断言，确保公开导出面只包含插件协议和领域能力需要的标识符。

## 3. 领域能力接口分层

- [x] 3.1 建立`usercap`、`authzcap`、`dictcap`、`filecap`、`sessioncap`、`configcap`、`notifycap`、`plugincap`、`jobcap`和`infracap`的`Service`、`AdminService`和必要的`ScopeService`接口。
- [x] 3.2 按统一模型重整现有`orgcap`、`tenantcap`和`ai`接口，删除旧生产入口并迁移到新的领域方法命名和返回投影。
- [x] 3.3 确认普通插件消费面不暴露`ScopeService`、`*gdb.Model`、宿主`DAO/DO/Entity`、HTTP 请求对象、私有缓存或宿主内部 service 实例。
- [x] 3.4 为每个领域方法标注调用职责、输入输出语义、数据权限边界、缓存边界、错误语义和批量性能边界。
- [x] 3.5 为新增或修改的运行期依赖记录 DI 来源检查，说明 owner、创建位置、传递路径、共享实例或共享后端策略。

## 4. 源码插件宿主服务目录

- [x] 4.1 扩展`pluginhost.Services`，提供普通领域服务目录和`Services.Admin()`完整类型化管理目录。
- [x] 4.2 调整源码插件注册、provider、hook、cron 和生命周期接入，确保业务对象通过构造函数接收最窄领域接口。
- [x] 4.3 为生命周期、hook、provider 回调和定时任务创建宿主系统 actor，并禁止插件自行伪造敏感调用上下文。
- [x] 4.4 在源码插件管理方法路径中补齐审计元数据，记录插件`ID`、actor、租户、来源、目标资源、结果状态和审计原因。
- [x] 4.5 运行覆盖`pluginhost`目录和受影响源码插件注册路径的 Go 编译门禁。

## 5. 动态插件`hostServices`协议

- [x] 5.1 扩展`plugin.yaml hostServices`声明结构，支持语言无关的领域`service`、领域`method`、资源或投影范围和管理方法声明。
- [x] 5.2 实现安装或启用阶段的领域方法授权确认和运行时授权快照生成，未发布、未声明或未授权的方法不得进入领域逻辑。
- [x] 5.3 扩展 WASM host service 分发器，在每次领域调用中构造`CapabilityContext`并转发到领域`Service`或`AdminService`适配器。
- [x] 5.4 更新动态插件 guest SDK 或等价代理目录，使用字符串编码领域`ID`、结构化错误和统一批量缺失语义。
- [x] 5.5 为请求型调用、系统型调用、未知方法、未授权方法、越权目标和结构化错误补充协议测试。
- [x] 5.6 运行覆盖`pluginbridge`、WASM host service 和 guest 协议包的 Go 编译门禁。

## 6. 动态`data`服务收窄

- [x] 6.1 在 manifest 校验、安装授权和运行时授权快照中校验`data`服务表归属，只允许当前插件自有表或宿主明确标记为该插件自有资源的表。
- [x] 6.2 永久拒绝动态`data`服务访问宿主核心`sys_*`表和官方能力插件内部表，并确保拒绝结果不会写入运行时授权快照。
- [x] 6.3 保留插件自有表访问的字段白名单、分页上限、排序、软删除、租户、数据权限、事务和审计治理，不放宽原始 SQL 禁止规则。
- [x] 6.4 为自有表成功访问、核心表拒绝、官方能力插件表拒绝和事务越界补充单元测试或集成测试。

## 7. 宿主领域适配器实现

- [x] 7.1 实现`usercap`读取、搜索、批量投影、可见性校验和管理动作，查询阶段注入租户与数据权限过滤。
- [x] 7.2 实现`authzcap`权限、角色、用户角色、菜单或按钮投影、授权关系管理和聚合统计，避免通过总数或候选项泄露不可见数据。
- [x] 7.3 实现`orgcap`组织、部门、岗位、成员候选、树形读取和范围注入能力，树形接口具备根节点、深度、分页或最大节点数边界。
- [x] 7.4 实现`tenantcap`租户投影、成员关系、租户可见性、系统调用边界和管理动作。
- [x] 7.5 实现`dictcap`字典类型、字典项、标签解析、批量标签投影和`i18n`标签返回语义。
- [x] 7.6 实现`filecap`文件投影、下载或引用校验、业务场景边界和目标可见性校验。
- [x] 7.7 实现`sessioncap`在线会话投影、批量读取、状态管理和会话缓存失效。
- [x] 7.8 实现`configcap`运行时配置读取、批量投影、管理变更和配置缓存失效。
- [x] 7.9 实现`notifycap`通知投影、发送或状态变更管理、目标用户可见性和审计。
- [x] 7.10 实现`plugincap`插件状态、资源引用、动态路由、授权快照和插件管理动作。
- [x] 7.11 实现`jobcap`任务、任务日志、调度状态、执行动作和聚合统计能力。
- [x] 7.12 实现`infracap`和重整后的`ai`基础设施或`AI`能力投影、状态报告和调用边界。
- [x] 7.13 为每个领域适配器补充单元测试，覆盖可见读取、不可见缺失、命令拒绝、批量上限、数据权限过滤和无`N+1`依据。

## 8. 缓存一致性与事务后失效

- [x] 8.1 建立关键运行时数据共享修订号抽象，覆盖权限、角色关系、租户成员、插件状态、插件资源、动态路由、字典、组织树、运行时配置和`hostConfig`。
- [x] 8.2 在领域管理写路径中实现事务提交成功后的幂等失效，事务回滚不得发布不可恢复的缓存更新或失效事件。
- [x] 8.3 实现单机模式本地缓存和本地失效分支，仍复用共享修订号抽象。
- [x] 8.4 实现集群模式共享后端、事件广播、分布式缓存或等价协调分支，禁止退化为仅当前节点可见的本地状态。
- [x] 8.5 为缓存后端不可用、修订号推进、重复失效、回源重建和能力不可用错误补充测试或审查证据。

## 9. 官方插件迁移

- [x] 9.1 修改任一`apps/lina-plugins/<plugin-id>/`文件前，先读取该插件根目录`AGENTS.md`普通文件或符号链接目标；无法读取时停止修改该插件目录。
- [x] 9.2 迁移`linapro-content-notice`生产代码，移除宿主`sys_*`生成和直接访问，改用领域能力完成用户、通知或字典投影。
- [x] 9.3 迁移`linapro-org-core`生产代码，移除宿主核心表直接访问，改用`orgcap`、`usercap`、`authzcap`或相关领域能力。
- [x] 9.4 迁移`linapro-tenant-core`生产代码，移除宿主核心表直接访问，改用`tenantcap`、`usercap`、`authzcap`或相关领域能力。
- [x] 9.5 迁移`linapro-monitor-online`生产代码，移除`sys_online_session`等宿主表生成和访问，改用`sessioncap`。
- [x] 9.6 迁移`linapro-monitor-operlog`生产代码，移除用户、字典或权限相关宿主表直查，改用`usercap`、`dictcap`和必要领域投影。
- [x] 9.7 迁移`linapro-monitor-loginlog`生产代码，移除用户、字典或会话相关宿主表直查，改用`usercap`、`dictcap`和`sessioncap`。
- [x] 9.8 对扫描发现的其他官方插件生产访问逐一迁移，更新插件`backend/hack/config.yaml`生成范围并重新生成插件 DAO。
- [x] 9.9 运行受影响插件后端包的 Go 编译门禁，确认插件生产代码不再导入宿主核心表生成工件。

## 10. SQL、DAO 与初始化资源

- [x] 10.1 判断本变更是否需要新增宿主`manifest/sql/{序号}-add-plugin-host-domain-capabilities.sql`，若需要则只写入幂等 DDL 和 Seed DML。
- [x] 10.2 判断插件迁移是否需要插件自身安装 SQL、卸载 SQL 或 Mock 数据调整，确保安装 SQL、卸载 SQL、Mock 数据目录和数据分类正确。
- [x] 10.3 涉及数据库结构、索引、唯一约束或查询辅助字段时，同步评估租户、数据权限、软删除、列表、聚合、树形和批量关联查询索引。
- [x] 10.4 执行`make db.init`、`make dao`或本仓库等价跨平台入口，生成结果不得手动修改`DAO/DO/Entity`。
- [x] 10.5 在任务记录中说明 SQL 幂等性、数据分类、自增主键写入、软删除语义和索引性能验证结果；若无 SQL 影响则明确记录。

## 11. API、文档元数据与`i18n`

- [x] 11.1 涉及 HTTP API 或 DTO 变更时，按 RESTful 语义、时间字段 Unix 毫秒、`g.Meta`、`dc`、`eg`和权限标签要求更新 API 源定义并重新生成控制器骨架。
- [x] 11.2 新增或修改 API 文档源文本、运行时错误、菜单、路由、按钮、表单、表格、字典或标签文案时，按宿主和插件`i18n`启用边界维护运行时语言包与`apidoc`资源。
- [x] 11.3 为领域能力返回的`labelKey`和可选`label`补充测试，确认已启用`i18n`和未启用`i18n`插件的边界符合规范。
- [x] 11.4 运行`make i18n.check`或本仓库等价入口；若本阶段确认无用户可见文案或 API 文档本地化影响，则在任务记录中明确说明。

## 12. 治理扫描与开发工具

- [x] 12.1 使用 Go 工具、`linactl`或等价跨平台入口实现治理扫描，禁止新增默认依赖平台专属 shell 语义的长期维护脚本。
- [x] 12.2 治理扫描覆盖插件`backend/hack/config.yaml`宿主核心表生成项、生产 Go 代码直接打开`sys_*`表、`shared.TableSys*`、旧领域接口和旧动态`host service`方法；Go 语言`internal`目录规则已经阻断的宿主`DAO/DO/Entity`导入和类型使用不重复扫描。
- [x] 12.3 治理扫描覆盖动态插件`plugin.yaml hostServices`和授权快照，阻断`data`服务声明宿主核心表或官方能力插件内部表。
- [x] 12.4 明确测试、Mock、安装 SQL 和迁移 SQL 的受控例外目录，例外不得进入插件生产运行路径。
- [x] 12.5 在根`Makefile`、`make.cmd`或 CI 中接入薄包装入口时，记录 Windows、Linux、macOS 的跨平台影响和验证方式。

## 13. 验证、审查与完成门禁

- [x] 13.1 运行变更范围 Go 编译门禁，至少覆盖`apps/lina-core/pkg/plugin/capability`、`pluginhost`、`pluginbridge`、受影响宿主领域 service、动态 host service 和所有迁移插件后端包。
- [x] 13.2 涉及 controller 构造、路由绑定、启动装配或 API 签名时，运行`cd apps/lina-core && go test ./internal/cmd -count=1`或更窄且能覆盖启动绑定的等价测试。
- [x] 13.3 运行领域能力单元测试、动态协议测试、`data`服务授权测试、缓存一致性测试、治理扫描测试和必要的回归测试。
- [x] 13.4 若本变更引入用户可观察页面、权限交互、接口联动或端到端工作流变化，使用`lina-e2e`规范补充 E2E 用例并记录覆盖；若无 E2E 影响则明确记录跳过原因。
- [x] 13.5 运行`openspec validate add-plugin-host-domain-capabilities --strict`，并修复所有校验问题。
- [x] 13.6 完成所有实现和验证后调用`lina-review`，审查 DI 来源、数据权限、缓存一致性、`i18n`、开发工具跨平台、测试策略、OpenSpec 任务状态和活跃变更边界。

## 执行记录

- 规则与变更上下文：本轮已重新读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`，并读取本变更`proposal.md`、`design.md`和全部增量规范。当前项目不考虑旧生产接口兼容，本迭代按一次性迁移执行。
- 插件本地规范：执行`find apps/lina-plugins -maxdepth 2 \( -type f -o -type l \) -name AGENTS.md -print`，未发现插件根目录本地`AGENTS.md`，因此所有插件目录变更按项目顶层规则和命中规则文件执行。
- 迁移范围与治理：官方插件生产代码已移除宿主`sys_*`核心表生成项和宿主`DAO/DO/Entity`直连，动态`data`服务收窄为当前插件自有表；新增`linactl plugins.check`跨平台 Go 扫描，覆盖`apps/lina-plugins`下所有包含`plugin.yaml`的插件目录、插件`backend/hack/config.yaml`、生产 Go 直接表访问、旧领域接口、旧动态`host service`方法、动态`data`核心表和外部插件表授权。测试、Mock、安装 SQL 和迁移 SQL 被记录为受控非生产例外。
- 领域能力与 DI 来源：新增`CapabilityContext`、领域命名`ID`、批量`MissingIDs`、`labelKey`/`label`和结构化错误契约；新增`usercap`、`authzcap`、`dictcap`、`filecap`、`sessioncap`、`configcap`、`notifycap`、`plugincap`、`jobcap`、`infracap`能力，并重整`orgcap`、`tenantcap`和`ai`。运行期依赖 owner 为宿主领域 owner，创建位置在宿主启动装配或`plugin`facade，传递路径为`pluginhost.Services`、`Services.Admin()`、动态`hostServices`适配器和 provider env；缓存、权限、插件状态、租户和`i18n`等关键服务复用启动期共享实例或共享后端，未在业务路径临时`New()`关键服务图。
- 源码插件管理面：源码插件通过`pluginhost.Services.Admin()`获得完整类型化`AdminService`目录；插件业务对象仍接收最窄领域接口，管理方法安全边界由领域 owner 执行租户、数据权限、目标状态、数量上限、系统 actor 和审计治理。
- 动态插件协议：动态插件`plugin.yaml hostServices`使用语言无关`service + method`声明；构建、安装、授权快照和运行时分发均使用 plugin-aware 规范化校验，未发布、未声明或未授权方法不会进入领域逻辑。
- 数据权限：读取、候选、批量、树形、导出和聚合路径均按领域能力边界在查询阶段或目标操作前接入租户与数据权限；批量读取以`MissingIDs`隐藏不存在与不可见差异，命令类`Ensure*`默认任一不可见整体失败。
- 缓存一致性：权限、角色关系、租户成员、插件状态、插件资源、动态路由、字典、组织树、运行时配置和`hostConfig`通过共享修订号与事务后失效治理；单机分支复用同一修订号抽象，集群分支通过共享后端、事件或协调机制避免本地状态退化。无新增仅当前节点可见的关键缓存。
- SQL/DAO：宿主未新增本迭代 SQL 文件；插件侧仅调整自身安装 SQL、DAO 生成范围和测试数据边界。插件安装 SQL 保持幂等，`linapro-tenant-core`新增成员唯一索引用于真实业务约束；未手动修改生成的宿主或插件`DAO/DO/Entity`，删除的生成文件来自插件`backend/hack/config.yaml`生成范围清理。
- API 与`i18n`：涉及动态插件 API 文档源文本和插件运行时文案的变更已按宿主与插件`i18n.enabled`边界维护；领域返回稳定值、`labelKey`和可选`label`，未启用`i18n`的插件不要求补自身语言包。`make i18n.check`通过，保留既有 48 条模块级`$t()`非阻塞 warning。
- 开发工具跨平台：新增治理扫描和 WASM 构建锁均使用 Go 工具链或`linactl`内部组件实现；根`Makefile`和`make.cmd`仅作为薄包装入口，未新增平台专属长期脚本。跨平台验证通过`go test ./hack/tools/linactl/... -count=1`、`make test.scripts`和`go run ./hack/tools/linactl plugins.check`。
- E2E 与测试策略：本轮按`lina-e2e`组织和修复宿主、插件、动态插件测试。搜索类用例改为自建自清理数据，插件生命周期和运行时测试改为稳定等待与隔离 fixture，避免依赖测试顺序、固定 seed 或并发构建副作用。关键用户可观察路径通过完整 E2E 覆盖。

## 验证证据

- `openspec validate add-plugin-host-domain-capabilities --strict`：通过。
- `git diff --check`和`git -C apps/lina-plugins diff --check`：通过。
- `go test ./hack/tools/linactl/... -count=1`、`go run ./hack/tools/linactl plugins.check`、`make test.scripts`：通过。
- `make i18n.check`：通过，保留 48 条非阻塞模块级`$t()`warning。
- `pnpm -C hack/tests test:validate`、`pnpm -C hack/tests exec tsc -p tsconfig.json --noEmit --pretty false`、`pnpm -C hack/tests test:service-deps`：通过。
- `make test.go plugins=1 race=false`：通过。
- `cd apps/lina-vben && pnpm test:unit`：通过，46 个文件、363 个测试通过。
- `cd apps/lina-vben && pnpm run check`：通过，保留既有 circular/unused dependency warning。
- 目标 E2E 模块：`iam:user` 28 通过，`iam:menu` 23 通过，`i18n` 30 通过，`settings` 93 通过；插件模块`extension:plugin`、`plugin:linapro-demo-dynamic`、`plugin:linapro-demo-source`、`plugin:linapro-monitor-operlog`、`plugin:linapro-ops-demo-guard`、`plugin:linapro-org-core`、`plugin:linapro-tenant-core`均已通过。
- 完整 E2E：`make test`日志`temp/20260606/085000-10-make-test-full.log`退出码为 0；并行段`36 passed`，串行段`546 passed`、`8 skipped`，合计`582 passed`、`8 skipped`、`0 failed`，串行段耗时约 51.6 分钟。

## Feedback

- [x] **FB-1**: 收敛`Services`公开领域能力目录，移除通知、会话和配置领域的重复公开入口
- [x] **FB-2**: 通用化插件规范检查入口并移除 Go 语法已阻断的宿主`DAO/DO/Entity`扫描
- [x] **FB-3**: 公开动态插件语言无关`hostServices`协议目录，便于开发者查看可声明服务、方法与资源边界
- [x] **FB-4**: 将动态插件 guest 能力 SDK 合并到`pluginbridge/guest`，移除`capability/guest`双入口混淆

### FB-1 执行记录

- 根因：`capability.Services`同时公开旧`Notify()`、`Session()`、`Config()`、`RuntimeConfig()`与新`Notifications()`、`Sessions()`、`configcap.Service`领域入口，通知、会话和配置领域在插件公开目录中存在重复语义；插件调用方仍在使用旧`contract.NotifyService`、`contract.SessionService`和含混的`Config()`入口。
- 实现：删除公开`contract.NotifyService`与`contract.SessionService`领域对象；`Services.PluginConfig()`唯一表示插件静态配置，`Services.Config()`唯一表示运行时配置领域；通知读面为`Services.Notifications()`，通知管理面为`Services.Admin().Notifications()`；会话读面为`Services.Sessions()`，会话管理面为`Services.Admin().Sessions()`。`session.Store`补充`BatchGetScoped`，`sessioncap.BatchGetSessions`改为数据库侧按`token_id IN (...)`批量读取并套用租户与数据权限过滤，不再通过第一页会话列表内存过滤导致可见会话误判缺失。`linapro-content-notice`、`linapro-monitor-online`和`linapro-monitor-server`均已迁移到新入口。
- DI 来源：未新增宿主运行期服务实例；通知、会话、运行时配置、业务上下文、租户过滤和插件配置能力均继续由`registrar.Services()`和`hostservices.New()`使用启动期共享实例传递。`linapro-monitor-online`构造函数新增的`BizCtx`、`TenantFilter`、`sessioncap.Service`和`sessioncap.AdminService`均来自同一 pluginhost 服务目录，未在业务路径临时`New()`关键服务图。
- 数据权限影响：插件访问通知、会话和用户数据仍经领域 owner 的`CapabilityContext`、租户过滤和数据范围校验；`linapro-monitor-online`强退改由`sessioncap.AdminService.RevokeSession`执行可见性校验；`sessioncap.BatchGetSessions`返回`Items`和不可解释的`MissingIDs`，不区分不存在与不可见，未新增绕过数据权限的路径。
- 缓存一致性影响：无新增缓存、快照或失效机制；会话状态仍由宿主共享 session store 和 auth revoke 机制负责，通知发送与删除仍由宿主 notify service 负责，运行时配置仍复用原`configcap`缓存边界。
- `i18n`影响：无运行时用户可见文案、菜单、路由、API 文档源文本、插件清单或语言包资源变更；本次只调整 Go 领域契约命名和后端调用路径。
- 开发工具跨平台影响：无脚本、`linactl`或 CI 入口变更；仅为`linapro-monitor-server`补齐与其他源码插件一致的`lina-core`模块依赖和本地`replace`，以支持跨平台 Go 工具链直接编译。
- 测试策略：本次为后端插件能力契约重构，无前端 UI 或端到端可观察工作流变化，未新增 E2E；使用反射治理测试、静态检索和相关 Go 包测试覆盖公开目录收敛、hostservices 适配、插件调用方和测试桩。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/capability ./internal/service/session ./internal/service/auth ./internal/service/plugin/internal/hostservices ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/wasm ./internal/service/plugin -count=1`通过；`GOWORK=off go test ./backend/... -count=1`分别在`linapro-content-notice`、`linapro-monitor-online`和`linapro-monitor-server`通过；`openspec validate add-plugin-host-domain-capabilities --strict`通过；静态检索确认`capability.Services`实现面不再包含旧`Notify()`、`Session()`或`RuntimeConfig()`入口，`contract`包不再定义旧通知和会话领域对象。

### FB-2 执行记录

- 根因：原`plugins.boundary.check`扫描器将 Go 语言`internal`目录天然阻断的`lina-core/internal/dao`、`do.Sys*`、`entity.Sys*`和`dao.Sys*`导入或类型使用也纳入静态规则，形成低价值重复检查；同时命令名绑定到`boundary`语义，缺少根`make plugins.check`入口，不适合作为后续扩展的通用插件规范辅助检查工具。
- 实现：删除`plugin-go-host-core-import`、`plugin-go-host-core-dao`、`plugin-go-host-core-do`和`plugin-go-host-core-entity`扫描规则，保留`backend/hack/config.yaml`宿主核心表生成项、生成文件、直接`sys_*`表访问、`shared.TableSys*`、旧`pluginbridge`、旧`host service`方法和动态`data`表归属检查；公开命令改为`linactl plugins.check`，新增根`make plugins.check`薄包装和`format=json`透传，补充`linactl`中英文说明；增量规范明确通用入口扫描`apps/lina-plugins`下所有包含`plugin.yaml`的插件目录。
- 规则加载与影响分析：本轮反馈已重新读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`。无插件目录文件变更，不触发插件本地`AGENTS.md`读取；无 HTTP API、SQL、数据读写、权限、缓存、前端 UI、运行时依赖或 E2E 资产影响。
- 数据权限影响：无数据查询、写入、导出、候选或存在性暴露路径变更；本次仅调整静态治理扫描规则和开发工具入口。
- 缓存一致性影响：无缓存、快照、修订号、失效或集群一致性路径变更。
- `i18n`影响：变更涉及 CLI 帮助、CLI 输出和工具 README 文案，但不涉及宿主或插件运行时 UI、菜单、路由、API 文档源文本、插件清单或语言包资源；无需维护`manifest/i18n`或`apidoc`翻译资源。
- 开发工具跨平台影响：核心逻辑仍在 Go `linactl`内部组件中；根`Makefile`仅新增薄包装目标，`make.cmd`保持参数转发，Windows 可通过`.\make plugins.check`进入同一`linactl plugins.check`实现；未新增平台专属脚本或 Shell 业务逻辑。
- 测试策略：本次是治理工具行为修复，无用户可观察页面或端到端业务流程变化，未新增 E2E；通过`plugingovernance`单元测试覆盖删除的 Go `internal`重复规则、保留的违规类别、受控非生产例外和文本输出，通过`linactl`注册测试覆盖新旧命令名。
- 验证证据：`go test ./hack/tools/linactl/... -count=1`通过；`go run ./hack/tools/linactl plugins.check`通过，扫描`438`个文件且`0`个发现项；`make plugins.check`通过；`make test.scripts`通过；`openspec validate add-plugin-host-domain-capabilities --strict`通过；`git diff --check`通过；`go run ./hack/tools/linactl help | rg "plugins\\.check|plugins\\.boundary\\.check"`只输出`plugins.check`。

### FB-3 执行记录

- 根因：动态插件已经通过`capability/guest`暴露 Go guest SDK 能力目录，但开发者需要查看`plugin.yaml hostServices`可声明的语言无关`service + method + resource`协议时，只能追到`pluginbridge/internal/hostservice`内部描述表；根`pluginbridge.go`只是包说明，不适合作为完整协议事实源，容易让公共协议入口和内部治理表的职责混淆。
- 实现：新增`pluginbridge/protocol/hostservice_catalog.go`，从内部`hostservice`描述表投影公开`HostServiceDescriptor`、`HostServiceMethodDescriptor`、`HostServiceResourceKind`和`HostServiceDescriptors()`/`HostServiceMethodDescriptors()`；公开字段仅包含语言无关的`service`、`method`、`capability`、资源形态、默认方法、payload 名称和`Published`状态，不暴露 Go 常量名、guest client 或 dispatcher 覆盖细节。根`pluginbridge.go`只补充说明`protocol`子包拥有 host-service catalog。
- DI 来源：无新增运行期依赖、构造函数、启动装配、插件宿主服务适配器或`WASM host service`实例；公开 catalog 仅从既有静态描述表复制投影，未引入共享实例或缓存状态。
- 数据权限影响：无数据查询、写入、导出、候选、授权关系变更或存在性暴露路径变更；本次只公开协议元数据，不改变运行时授权、领域数据权限或`data`服务表归属校验。
- 缓存一致性影响：无缓存、快照、修订号、失效、刷新或集群一致性路径变更。
- `i18n`影响：无运行时用户可见文案、菜单、路由、按钮、API 文档源文本、错误消息、插件清单或语言包资源变更。
- 开发工具跨平台影响：无脚本、`linactl`、CI、构建、测试或跨平台入口变更。
- 测试策略：本次是 Go 公共协议契约和可读性改进，无前端 UI、用户可观察页面或端到端工作流变化，未触发 E2E；新增`protocol`单元测试校验公开 catalog 与内部描述表同步、返回副本不可被调用方篡改，并确认公开 method descriptor 不暴露`MethodConst`、`GuestClient`和`Dispatcher`等实现字段。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/protocol ./pkg/plugin/pluginbridge/internal/hostservice -count=1`通过；`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... -count=1`通过；`openspec validate add-plugin-host-domain-capabilities --strict`通过；`git diff --check -- apps/lina-core/pkg/plugin/pluginbridge/pluginbridge.go apps/lina-core/pkg/plugin/pluginbridge/protocol/hostservice_catalog.go apps/lina-core/pkg/plugin/pluginbridge/protocol/hostservice_catalog_test.go openspec/changes/add-plugin-host-domain-capabilities/tasks.md`通过。

### FB-4 执行记录

- 根因：动态插件业务代码同时需要从`capability/guest`获取宿主服务能力客户端，又需要从`pluginbridge/guest`获取路由执行、请求绑定、响应 helper 和底层 host-call transport；两个包名都叫`guest`且都只服务动态插件，开发者需要理解一条人为分层才能选择导入路径。该分层没有继续降低复杂度，反而让动态插件`guest`侧公开入口发散。
- 修复：删除`apps/lina-core/pkg/plugin/capability/guest`目录，将 runtime、storage、network、record store、cache、lock、plugin config、notify、cron、host config、manifest、org、tenant 和`AI`等 guest 能力客户端合并到`apps/lina-core/pkg/plugin/pluginbridge/guest`。`pluginbridge/guest`主包说明同步扩展为动态插件运行时、路由 helper、host-call transport 和 guest 能力客户端统一入口；动态插件样例改为只从`lina-core/pkg/plugin/pluginbridge/guest`获取业务 host-service client 和桥接 helper。
- `recordstore`导入边界：为避免`pluginbridge/guest -> recordstore -> pluginbridge/guest`循环依赖，`recordstore`新增`HostServiceInvoker`窄 transport 回调和`OpenWithHostServiceInvoker`入口；`pluginbridge/guest`在`RecordStore()`中注入自身`invokeGuestHostService`。治理扫描同步阻断旧`capability/guest`导入，并阻断`recordstore`反向导入`pluginbridge/guest`。
- DI 来源：无新增宿主运行期依赖 owner、构造函数、启动装配、插件宿主服务适配器或`WASM host service`实例。新增的`recordstore.HostServiceInvoker`是 guest 进程内 transport 函数注入，不持有宿主服务实例、缓存状态或共享后端，也不在业务路径临时`New()`关键服务图。
- 数据权限影响：无数据查询、写入、导出、候选、授权关系变更或存在性暴露语义变化。动态`data`服务仍经原`HostServiceData`协议、授权快照、表归属校验、字段投影、租户和数据权限过滤执行；本次只调整 guest SDK 包归属和 transport 注入方向。
- 缓存一致性影响：无缓存、快照、修订号、失效、刷新或集群一致性路径变更；runtime、cache、config、plugin state 等 host service 方法的协议和宿主处理器均未改变。
- `i18n`影响：无运行时用户可见文案、菜单、路由、按钮、API 文档源文本、错误消息、插件清单或语言包资源变更。README 中英文镜像仅同步开发者导入路径说明，不涉及运行时翻译资源。
- 文档治理影响：更新`linapro-demo-dynamic`的`README.md`和`README.zh-CN.md`，中英文事实保持一致；宿主插件包 README 当前已经指向`pluginbridge/guest`统一入口，无需额外修改。
- 开发工具跨平台影响：无脚本、`linactl`、CI、构建入口或跨平台命令变更。
- 测试策略：本次是动态插件 guest SDK 包边界收敛和内部 transport 注入重构，无前端 UI、用户可观察页面或端到端工作流变化，未触发 E2E；通过 Go 包测试、治理扫描测试、动态插件后端测试和`wasip1`依赖解析验证。
- 规则读取：已重新读取`AGENTS.md`以及`openspec`、`documentation`、`architecture`、`plugin`、`backend-go`、`testing`、`data-permission`、`cache-consistency`和`i18n`规则；确认无 HTTP API、SQL、DAO/DO/Entity、前端 UI、缓存一致性、数据权限语义或开发工具跨平台入口变更。
- 插件本地规范：修改`apps/lina-plugins/linapro-demo-dynamic`前已执行`find apps/lina-plugins -maxdepth 2 \( -type f -o -type l \) -name AGENTS.md -print`，未发现插件根目录`AGENTS.md`普通文件或符号链接。
- 验证证据：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./pkg/plugin/capability/... -count=1`通过；`cd apps/lina-core && GOOS=wasip1 GOARCH=wasm go list -deps -tags wasip1 ./pkg/plugin/pluginbridge/guest ./pkg/plugin/capability/recordstore`通过；`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过；静态检索确认生产代码无旧`lina-core/pkg/plugin/capability/guest`导入，`recordstore`不导入`pluginbridge/guest`；`openspec validate add-plugin-host-domain-capabilities --strict`通过。
