## Context

每个 LinaPro 插件按照仓库约定在 `manifest/sql/` 下分三类 SQL 资源：

```
plugin/manifest/sql/
├── 001-*.sql              # install: DDL + Seed DML，安装时执行
├── mock-data/             # 演示/测试数据，目前安装流程不会执行
│   └── 001-*-mock.sql
└── uninstall/             # uninstall: 反向 DDL + 数据清理
    └── 001-*.sql
```

宿主在 `apps/lina-core/internal/service/plugin/internal/lifecycle/` 下提供 `ExecuteManifestSQLFiles(manifest, direction)`，目前只对 `install` 与 `uninstall` 两个方向工作；mock-data 目录的扫描和执行**完全没有接入**。

数据库为 MySQL（`config.yaml` 中 `link: "mysql:..."`）；插件的 install SQL 普遍混合 DDL 与 Seed DML（如 `content-notice/manifest/sql/001-content-notice-schema.sql` 既建表又写字典种子数据），插件的 mock-data SQL 全部为纯 DML（`grep` 全量验证通过）。

本变更的目标是给用户在插件安装时多一个**"是否加载示例数据"**的选项，并在加载该选项时提供"全成全败"的事务化语义。

## Goals / Non-Goals

**Goals:**

- 用户在插件安装弹窗里看到"是否安装示例数据"复选框，勾选即在安装链路中执行 `manifest/sql/mock-data/*.sql`。
- 多个 mock-data SQL 文件以及对应的迁移账本写入在**同一数据库事务**中执行；任一失败则整体回滚，插件不会出现"半加载 mock"的脏状态。
- `plugin.autoEnable` 配置可以**逐项 opt-in**示例数据，默认关闭。
- 失败响应里携带可操作信息（失败文件、原因、"已自动回滚"提示），让用户决定是接受现状还是修复后卸载重装。
- 源码插件与动态插件 UX 一致。

**Non-Goals:**

- 不提供"安装后单独加载/清除示例数据"的独立 API 或 UI 入口（重装即可，参见后续 Decisions）。
- 不为 install SQL 自身提供事务化能力（受限于 MySQL DDL 隐式提交，物理上做不到）。
- 不为生产环境引入额外栅栏（既不加 publicconfig 开关、也不加额外权限）；勾选行为本身即视为"知情同意"，autoEnable 自动安装通过 opt-in 配置控制。
- 不修改宿主 `apps/lina-core/manifest/sql/mock-data/` 的加载链路（继续走 `make mock`，与本变更解耦）。
- 不向插件作者引入新的目录或新的清单字段——`mock-data/` 目录已经是仓库约定的一部分。

## Decisions

### Decision 1：Install SQL 不能事务化，Mock SQL 必须事务化

**选择**：维持 install SQL 现状（逐文件执行、依赖幂等），将 mock SQL 与对应的 `sys_plugin_migration` 写入包入 `dao.SysPluginMigration.Transaction(ctx, ...)` 闭包中执行。

**理由**：

- **MySQL/MariaDB 对 DDL 强制隐式提交**：`CREATE TABLE`、`ALTER TABLE`、`CREATE INDEX` 等 DDL 一旦执行立即结束当前事务并落库，`ROLLBACK` 无法回滚已经执行的 DDL。这是引擎层的约束，无法通过应用层规避。
- 插件 install SQL 普遍混合 DDL 与 Seed DML（已在样例插件验证），强行包事务等于自欺欺人。
- 插件 mock SQL 已经是纯 DML（已对 8 个插件全量验证），事务包裹是干净的。
- 现有 install SQL 已遵循"幂等可重入"规范（`CREATE TABLE IF NOT EXISTS` / `INSERT IGNORE`），失败重试是安全的。

**考虑过的替代方案**：

- **Saga 补偿（mock 失败时跑 uninstall SQL 清理 install 留下的表）**：能在用户视角呈现"全成全败"的语义，但引入新失败链路（补偿本身可能失败），且 install 已经写入的 Seed DML 会被一并清掉，对用户可见行为影响过大。否决，理由是"为低概率事件付出过高复杂度"。
- **要求 install SQL 全部为 DML、把 DDL 搬到别处**：违反插件作者既有目录约定，且 DDL 必须先于 DML 执行，本质是把问题转成"另一个文件夹的相同问题"。否决。

### Decision 2：Mock 阶段独立成 phase，不与 install 阶段耦合

**选择**：在 `catalog.MigrationDirection` 类型上新增第三个枚举值 `MigrationDirectionMock = "mock"`；在 `ExecuteManifestSQLFiles` 调用链中新增专用入口 `ExecuteManifestMockSQLFilesInTx(ctx, tx, manifest)`，仅当用户/配置 opt-in 时才被调用。

**理由**：

- mock 阶段需要事务执行，与 install 阶段的 per-file 直执行模型签名不同，硬塞进同一个函数会让分支逻辑混乱。
- `MigrationDirection` 已是开放枚举类型，新增第三个值是最小改动。
- `sys_plugin_migration` 账本可继续复用，按 `direction='mock'` 区分。
- 将来若要扩展第四个阶段（例如版本升级时的"data fixup"），同一套结构可继续扩展。

**考虑过的替代方案**：

- **复用 install 方向，仅在 SQL 资产 Key 上加前缀区分**：账本无法区分 phase，运维侧失去结构化可追溯性，且对 SQL 资产排序逻辑造成干扰。否决。

### Decision 3：失败语义 — install/mock 解耦

**选择**：

| 失败点               | 处理                                                                                                      |
| -------------------- | --------------------------------------------------------------------------------------------------------- |
| install SQL 中途失败 | 与现状一致——返回错误，依赖用户修复并重试（幂等保证安全）                                                  |
| mock SQL 中途失败    | 整个 mock 事务回滚，插件保持"已安装、无 mock"状态；响应携带失败文件名 + 原因 + "已自动回滚 mock 数据"提示 |
| 未勾选 mock 选项     | mock 阶段完全跳过，等价于现状                                                                             |

**理由**：用户的明确诉求是"mock 失败时告知失败原因，由用户决定是否解决问题、还是取消 mock 重新安装"。这意味着**安装本身不应被 mock 失败连带回滚**——否则用户连"接受现状"的选择都没有了。"修复后卸载重装"是用户已接受的重试路径。

**考虑过的替代方案**：

- **mock 失败时也回滚 install（Saga 补偿）**：见 Decision 1 末尾否决理由。
- **mock 失败时用降级状态码（如 200 + warning）**：违反"明确告知失败原因"的诉求；前端容易误判为成功。否决。

### Decision 4：autoEnable 配置升级为联合类型

**选择**：`plugin.autoEnable` YAML schema 兼容两种条目写法：

```yaml
plugin:
  autoEnable:
    - "demo-control" # 字符串：默认不带 mock
    - id: "plugin-demo-source" # 对象：可显式 opt-in
      withMockData: true
    - id: "plugin-demo-dynamic" # 对象但不 opt-in，等价于字符串写法
```

**理由**：

- 默认行为安全：现有所有以字符串形式书写的 autoEnable 配置自动获得"启动期不安装 mock"的语义，零迁移成本。
- 表达力够用：开发环境可以为某几个特定插件 opt-in，不需要全量打开。
- 与"用户 UI 勾选即同意"的语义对齐：autoEnable 是机器代为决策，必须显式声明才装 mock。

**考虑过的替代方案**：

- **全局开关 `plugin.autoEnableInstallMockData: bool`**：粒度太粗。否决。
- **完全锁死，autoEnable 永不带 mock**：与"开发环境希望演示数据自动到位"的实际诉求冲突，把开发体验外包到"启动后人工触发"，反而更复杂。否决。

### Decision 5：autoEnable 启动场景下 mock 失败的处理

**选择**：与现有 `BootstrapAutoEnable` 的"任一失败 panic 阻塞启动"语义保持一致——若 opt-in 配置了 `withMockData: true` 且 mock 阶段失败，则启动失败 panic。

**理由**：

- 启动期是不可降级的关键链路，与运行期 API 行为不同。如果运维显式声明了"启动时带 mock"，那 mock 失败本身就是配置/数据问题，应阻塞启动让运维介入。
- 与 `plugin-startup-bootstrap` 已有的"Any failure for a listed auto-enable plugin must block host startup"要求一致，避免引入新的失败类别。

### Decision 6：迁移账本 `sys_plugin_migration.direction` 字段扩展

**选择**：

- `direction` 字段可取 `install` / `uninstall` / `mock` 三种值。如果当前列定义为 `VARCHAR`，无需 schema 变更，仅需扩展应用层枚举；如果定义为 `ENUM(...)` 或带 CHECK 约束，则在本迭代 `apps/lina-core/manifest/sql/<NNN>-plugin-install-with-mock-data.sql` 中通过 `ALTER TABLE` 扩展约束。
- mock 阶段的 ledger 行写入与 mock SQL 在**同一事务**内，回滚时一起回滚，确保账本最终状态与实际执行一致。
- 卸载链路无需特殊处理 `mock` 方向的账本——uninstall SQL 清表时这些 mock 数据行自然消失；账本里的 mock 行作为历史记录保留即可（也可在 uninstall 时一并删除，留待 tasks.md 决策）。

### Decision 7：失败响应载荷形态

**选择**：在插件模块的 `*_code.go` 中定义错误码 `bizerr.NewCode("plugin.install.mockDataFailed", ...)`，错误的 `messageParams` 中携带：

```jsonc
{
  "errorCode": "plugin.install.mockDataFailed",
  "messageKey": "plugins.install.error.mockDataFailed",
  "messageParams": {
    "pluginId": "content-notice",
    "failedFile": "001-content-notice-mock-data.sql",
    "rolledBackFiles": ["001-content-notice-mock-data.sql"],
    "cause": "Duplicate entry 'admin' for key ...",
  },
}
```

i18n key `plugins.install.error.mockDataFailed` 在三个 locale 下分别表达"插件 {pluginId} 安装成功，但示例数据 {failedFile} 加载失败：{cause}。已自动回滚 mock 数据，您可以接受当前状态或修复后卸载重装"等同义文案。

**理由**：与项目 `bizerr` 治理规范一致，前端通过 errorCode 决定如何提示，messageParams 让 i18n 文案保留可填充字段。

### Decision 8：前端复选框位置与默认值

**选择**：复用现有 `plugin-host-service-auth-modal.vue`（**所有插件**安装时都会经过的弹窗，包括没有 host service 授权需求的源码插件），在插件基础信息表格之后增加：

```
☐ 是否安装示例数据   ?
   仅用于演示和功能验证；生产环境不建议勾选。
```

- 默认 `false`，需要用户主动勾选。
- 仅在插件具备 `manifest/sql/mock-data/` 目录时显示（前端通过插件元数据接口获知 `hasMockData` 字段）。
- tooltip 文案 + 复选框 label 走 i18n。

## Risks / Trade-offs

- **[Risk] 插件作者写出的 mock SQL 在某些环境下失败（如时区差异、外键依赖错乱）** → Mitigation：mock 失败已设计为不影响 install；提供清晰的失败响应让用户知道如何处置；E2E 用例覆盖典型失败路径。
- **[Risk] autoEnable opt-in mock 在 CI/E2E 环境意外打开后出现幂等冲突** → Mitigation：现有 mock SQL 全部使用 `INSERT IGNORE` 或 `WHERE NOT EXISTS` 模式，已是仓库治理规范要求；启动失败时账本能定位到具体 SQL 文件。
- **[Risk] 联合类型 schema 在 GoFrame 反序列化下踩坑（YAML 数组元素既可能是 string 也可能是 map）** → Mitigation：通过 `interface{}` 接收后由 `normalizePluginAutoEnableIDs` 改造为 `normalizePluginAutoEnableEntries` 方法做标准化；补 panic allowlist 与 unit test 防回归。
- **[Risk] 前端复选框与现有"安装并启用"操作混杂，UX 拥挤** → Mitigation：将"是否安装示例数据"放在插件基础信息表格下方的独立区域并加问号 tooltip；E2E 用例验证布局可见性。
- **[Risk] mock 阶段日志噪音过大（每个 SQL 文件一行 INFO）** → Mitigation：复用 install 阶段已有的日志策略；失败时使用 `logger.Warningf(ctx, ...)` 携带 pluginID + filename，避免吞噪。
- **[Trade-off] install 自身失败仍然不可回滚**：用户接受了这一物理限制；通过文档（错误码消息 + 提示）引导用户走"幂等重试"路径，已是当前最优解。

## Migration Plan

仓库约定项目无历史包袱、不考虑兼容性，因此：

1. 直接修改 `apps/lina-core/internal/service/plugin/` 下相关源文件与公共接口签名，不保留旧接口适配。
2. 数据库变更通过新增 `apps/lina-core/manifest/sql/<NNN>-plugin-install-with-mock-data.sql` 一次性落地（若 `sys_plugin_migration.direction` 当前为受约束类型则在该文件 ALTER；否则纯应用层改动即可）。
3. `plugin.autoEnable` 配置文件向后兼容（字符串写法不变），无需迁移现有部署。
4. `make init` + `make dao` 流程跑通即可，按现有规范执行。

回滚策略：因为本变更不破坏现有数据（既不删表也不改既有列语义），出问题时回退代码即可，无需特殊数据回滚步骤。

## Open Questions

1. **uninstall 时是否清理账本里 `direction='mock'` 的历史行？** 倾向"清理"（与 install/uninstall 行为一致——卸载即清账本），但保留为 tasks.md 阶段的微决策。
2. **前端 `hasMockData` 元数据来自哪里？** 倾向在 `pluginListItem` 响应里增加一个布尔字段，由后端扫描 manifest FS 决定。需要确认是否会影响列表性能（可缓存）。
3. **错误响应里 `rolledBackFiles` 是否需要包含**已经成功执行的 mock SQL 文件清单？倾向包含，让用户清楚"全部回滚"的事实，但需要在 lifecycle 层把已执行文件清单透出来。
