## 1. 变更文档与数据模型

- [x] 1.1 创建 `add-media-plugin` OpenSpec 提案、设计、规范和任务清单，并记录中文-only i18n 决策
- [x] 1.2 将 `media_v2.md` 的 MySQL 表结构转换为 media 插件 PostgreSQL 安装 SQL 与卸载 SQL
- [x] 1.3 配置 media 插件本地 `gf gen dao` 并生成 DAO/DO/Entity

## 2. 源码插件骨架

- [x] 2.1 创建 `apps/lina-plugins/media` 目录、`plugin.yaml`、`plugin_embed.go`、README 和 manifest 说明文档
- [x] 2.2 将 media 插件接入 `apps/lina-plugins` 聚合模块与 `go.work`
- [x] 2.3 注册 media 插件后端路由和源码插件资产

## 3. 后端接口与服务

- [x] 3.1 实现媒体策略 RESTful DTO、控制器和服务
- [x] 3.2 实现设备/租户/租户设备策略绑定 RESTful DTO、控制器和服务
- [x] 3.3 实现策略解析接口，按租户设备、设备、租户、全局优先级返回结果
- [x] 3.4 实现流别名 RESTful DTO、控制器和服务
- [x] 3.5 新增后端业务错误元数据单元测试

## 4. 前端页面

- [x] 4.1 实现中文-only 的媒体策略、绑定和流别名管理页面
- [x] 4.2 实现前端 API client、弹窗表单和操作按钮权限控制
- [x] 4.3 确认本模块不创建运行时 i18n JSON、manifest i18n 或 apidoc i18n JSON

## 5. 测试与验证

- [x] 5.1 新增插件自有 E2E 冒烟用例 `TC0234-media-plugin-smoke.ts`
- [x] 5.2 运行 Go 单元测试和源码插件聚合编译
- [x] 5.3 运行前端类型检查、E2E 静态校验和 OpenSpec 严格校验

## 6. 验证记录

- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `go test ./...` 于 `apps/lina-plugins` 通过。
- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本模块未新增运行时 i18n、manifest i18n 或 apidoc i18n 资源。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts --list` 能发现 1 条插件自有 E2E 用例。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] 插件安装 SQL 使用 `psql` 重复执行通过，并完成策略、设备绑定、租户绑定、租户设备绑定、`auto_remove=0/1` 流别名的最小写入与清理冒烟验证。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过；默认 PATH 下该脚本会命中 `/usr/local/bin/pnpm` 8.6.0，因此验证时显式优先使用当前用户 pnpm。
- [ ] 宿主 `internal/service/plugin` 相关局部测试未通过：当前本地数据库缺少既有表 `plugin_multi_tenant_user_membership`；`internal/service/menu`、`internal/controller/menu`、`internal/controller/plugin` 相关测试已通过，失败项与 media 插件实现无关。

## Feedback

- [x] **FB-1**: 媒体管理页面触发 `SES_UNCAUGHT_EXCEPTION` 且页面高度持续增长
- [x] **FB-2**: 为媒体管理插件增加演示案例数据
- [x] **FB-3**: media 配置应全平台共享，不做宿主租户隔离
- [x] **FB-4**: 设置全局媒体策略失败并返回 `error.media.strategy.update.failed`
- [x] **FB-5**: 策略绑定中的设备绑定、租户绑定、租户设备绑定应拆成互不混用的独立页面操作和独立接口
- [x] **FB-6**: 通过宿主静态入口 `http://127.0.0.1:8080/#/media` 访问媒体管理时显示“插件页面未找到”
- [x] **FB-7**: 媒体管理各编辑弹窗字段回显不正确
- [x] **FB-8**: 核实媒体管理每个界面及界面触发接口的执行情况
- [x] **FB-9**: 将 media 对外租户字段命名从 `bizTenantId` 改回 `tenantId`
- [x] **FB-10**: media 数据库时间字段应保持 `media_v2.md` 给定的字段名和字段结构
- [x] **FB-11**: 媒体策略新增失败，新增弹窗疑似复用旧编辑状态并调用更新接口
- [x] **FB-12**: 仔细核实 media 每一个后端接口、前端页面和页面触发接口，确保模块完整且正确

## Feedback 验证记录

- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，覆盖媒体页面加载、三页签切换、前端未捕获异常、页面高度稳定性、策略绑定优先级、被引用策略删除保护和流别名 CRUD。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本次反馈修复未新增 i18n 资源。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] `psql` 连续执行 `manifest/sql/mock-data/001-media-mock-data.sql` 两次通过，案例数据保持 3 条策略、1 条设备绑定、1 条租户绑定、1 条租户设备绑定和 3 条流别名，不重复写入。
- [x] FB-3 调整后，插件清单改为 `scope_nature: platform_only`、`supports_multi_tenant: false`、`default_install_mode: global`，并移除 media 表结构、生成 DAO/DO/Entity 和服务层查询中的 `host_tenant_id` 宿主租户隔离维度。
- [x] 使用显式 PostgreSQL TCP 连接执行 media 插件卸载 SQL 和安装 SQL 通过；默认 socket 方式 `psql` 在本机不可用，因此验证命令使用 `postgresql://postgres:postgres@127.0.0.1:5432/linapro?sslmode=disable`。
- [x] `gf gen dao` 于 `apps/lina-plugins/media/backend` 通过，生成文件已与平台共享表结构对齐。
- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `go test ./...` 于 `apps/lina-plugins` 通过。
- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，2 条用例全部通过。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本次平台共享调整未新增 i18n 资源。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] SQL 静态扫描通过：media 插件 SQL 不包含 `host_tenant_id`、MySQL 方言标记或显式自增 `id` 写入。
- [x] FB-4 修复 `clearGlobalStrategies` 在平台共享模式下无 WHERE 更新被 GoFrame Safe 模式拒绝的问题；清理旧全局策略时限定 `global=1`。
- [x] 补充 `TC0234-media-plugin-smoke.ts` 全接口 API 场景，覆盖策略列表、详情、新增、修改、启用状态、设置全局、删除，设备绑定、租户绑定、租户设备绑定的独立列表/保存/删除，策略解析，流别名列表、详情、新增、修改、删除。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，3 条用例全部通过。
- [x] 后端日志扫描确认本轮 E2E 后未再出现 `MEDIA_STRATEGY_UPDATE_FAILED`、`更新媒体策略失败` 或 `there should be WHERE condition statement for UPDATE operation`。
- [x] FB-5 将原 `GET/PUT/DELETE /media/bindings` 拆分为 `GET/PUT/DELETE /media/device-bindings`、`/media/tenant-bindings`、`/media/tenant-device-bindings` 三组独立资源接口，接口不再接收或返回绑定 `scope`。
- [x] FB-5 前端将“策略绑定”单页签拆分为“设备绑定”“租户绑定”“租户设备绑定”三个独立页签，并将“策略解析”独立成页签；绑定弹窗不再展示绑定类型切换。
- [x] `rg -n "media/bindings|BindingScope|CodeMediaBindingScopeInvalid|ListBindings\\(|SaveBinding\\(|DeleteBinding\\(|activeBindingScope|saveMediaBinding|listMediaBindings|deleteMediaBinding" apps/lina-plugins/media apps/lina-vben hack/tests openspec/changes/add-media-plugin/specs -g '!backend/internal/dao/**' -g '!backend/internal/model/**'` 无命中，确认旧混合绑定接口和前端调用已移除。
- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `go test ./...` 于 `apps/lina-plugins` 通过。
- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，3 条用例全部通过，覆盖独立绑定接口、独立页签、策略解析和流别名 CRUD。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本模块仍按用户要求中文-only，未新增 i18n 资源。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] FB-6 根因确认为宿主 8080 使用内嵌静态前端资源，旧产物未包含 media 源码插件前端页；`linactl dev/build` 已调整为先清理 `apps/web-antd/dist`、执行 `pnpm -F @lina/web-antd build`、再同步到 `apps/lina-core/internal/packed/public`，确保宿主静态入口包含源码插件页面。
- [x] `make stop && make dev` 通过，启动日志包含 `Building frontend...` 与 `Host frontend embedded assets generated`，确认开发启动会刷新宿主内嵌前端资源。
- [x] `grep -ho "media/device-bindings" apps/lina-core/internal/packed/public/js/bootstrap-*.js`、`grep -ho "/media/strategies" ...`、`grep -ho "媒体策略" ...` 均可命中新产物，确认宿主内嵌前端包含 media 页面与独立接口调用。
- [x] `go test ./hack/tools/linactl` 通过，并覆盖 `runDev` 会生成宿主前端内嵌 `index.html` 与保留 `.gitkeep`。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts -g "TC-234d"` 通过，覆盖 `http://127.0.0.1:8080/#/media` 宿主静态入口可加载媒体管理页面且不显示“插件页面未找到”。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，4 条用例全部通过。
- [x] FB-7 修复策略、绑定、流别名弹窗回显状态维护：弹窗表单改为保持同一个 `reactive` 对象引用，打开时先重置校验和默认值，再写入后端详情或表格行数据，避免 `resetFields()` 和整体替换 `formData` 导致编辑数据被旧初始值覆盖。
- [x] FB-8 为策略、设备绑定、租户绑定、租户设备绑定、流别名编辑按钮补充稳定 `data-testid`，用于逐界面核实真实页面操作和接口执行。
- [x] 补充 `TC0234-media-plugin-smoke.ts` UI 场景 `TC-234e`，覆盖策略编辑详情 `GET`/更新 `PUT`、三类绑定页签列表 `GET`/保存 `PUT`、策略解析 `GET`、流别名详情 `GET`/更新 `PUT`，并断言每个编辑弹窗字段回显。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts -g "TC-234e"` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，5 条用例全部通过。
- [x] FB-9 将 media 对外接口、前端状态、页面列、表单、策略解析参数和 E2E 测试中的 `bizTenantId` 统一改回 `tenantId`；数据库列仍保持原始 `tenant_id`，且不恢复 `host_tenant_id` 宿主租户隔离字段。
- [x] FB-9 更新 media 插件 SQL 注释并通过 `gf gen dao` 重新生成插件本地 DAO/DO/Entity，使生成文件中的租户字段说明同步为“租户ID”。
- [x] `rg -n "bizTenantId|BizTenantId|bizTenantID|BizTenantID|业务租户" apps/lina-plugins/media apps/lina-core/internal/packed/public openspec/changes/add-media-plugin -g '!backend/internal/dao/**' -g '!backend/internal/model/**'` 仅命中 FB-9 任务描述，确认代码、前端产物和规范正文无旧字段残留。
- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `go test ./...` 于 `apps/lina-plugins` 通过。
- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本次字段命名回退未新增 i18n 资源。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] `make stop && make dev` 通过，后端与前端服务重新加载 `tenantId` 字段版本。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，5 条用例全部通过，覆盖 `tenantId` 版租户绑定、租户设备绑定、策略解析、页面回显和 8080 宿主静态入口。
- [x] FB-10 将 media 插件数据库时间字段恢复为 `media_v2.md` 原始结构：`media_strategy.create_time/update_time`、`media_stream_alias.create_time`，不再使用 `created_at/updated_at`，也不再为流别名虚构更新时间字段。
- [x] FB-10 在 PostgreSQL 中通过 `trg_media_strategy_update_time` 触发器等价实现策略表 `update_time` 自动更新时间，并重新执行插件卸载 SQL、安装 SQL 与 `gf gen dao`。
- [x] FB-10 将后端服务、API DTO、前端类型和页面列同步为 `createTime/updateTime`；流别名接口和页面仅保留 `createTime`。
- [x] `psql` 结构断言通过：`media_strategy` 字段为 `id,name,strategy,global,enable,creator_id,updater_id,create_time,update_time`，`media_stream_alias` 字段为 `id,alias,auto_remove,stream_path,create_time`。
- [x] `psql` 触发器断言通过：`media_strategy` 存在 `trg_media_strategy_update_time`。
- [x] `rg -n "created_at|updated_at|createdAt|updatedAt|CreatedAt|UpdatedAt|columns\\.CreatedAt|columns\\.UpdatedAt|\\.CreatedAt|\\.UpdatedAt" apps/lina-plugins/media openspec/changes/add-media-plugin -g '!backend/internal/dao/**' -g '!backend/internal/model/**'` 无命中。
- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `go test ./...` 于 `apps/lina-plugins` 通过。
- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本次时间字段回退未新增 i18n 资源。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] `make stop && make dev` 通过，后端与前端服务重新加载时间字段版本。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，5 条用例全部通过，覆盖策略和流别名接口、页面列表、编辑回显与 8080 宿主静态入口。
- [x] FB-11 修复策略、流别名和绑定新增入口复用旧弹窗状态的问题：新增时显式清空旧 `id`、`tenantId`、`deviceId` 和 `strategyId`，表单重置时移除旧主键，避免编辑后新增误走更新接口。
- [x] FB-12 逐项核实后端 API、Controller、Service、前端 `media-client.ts` 和页面按钮调用：策略、设备绑定、租户绑定、租户设备绑定、策略解析、流别名均保持独立 REST 资源和正确 HTTP 方法。
- [x] FB-12 为策略启停、设为全局、删除，以及三类绑定和流别名删除按钮补充稳定 `data-testid`，并在 `TC0234-media-plugin-smoke.ts` 的 UI 场景中真实点击验证页面触发接口。
- [x] FB-12 修复演示 mock 数据在已有全局策略时可能撞 `uk_media_strategy_single_global` 的问题；默认全局策略仅在当前没有全局策略时写入。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH pnpm -F @lina/web-antd typecheck` 于 `apps/lina-vben` 通过；从仓库根使用 `corepack pnpm -F` 会因当前 shell 解析到不匹配 pnpm 版本或缺少 workspace bin 而失败，验证改用前端工作区根路径。
- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `go test ./...` 于 `apps/lina-plugins` 通过。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts -g "TC-234e"` 通过，覆盖策略编辑后新增 `POST /media/strategies`、策略启停/设全局/删除、三类绑定编辑后新增与删除、流别名编辑后新增 `POST /media/stream-aliases` 与删除。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，5 条用例全部通过，覆盖媒体 API 语义、页面加载、8080 宿主静态入口和页面全操作链路。
- [x] `E2E_HOST_BASE_URL=http://127.0.0.1:8080 ./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts -g "TC-234d"` 通过，确认宿主 8080 静态入口已加载最新 media 页面；直接把 `E2E_BASE_URL` 改为 8080 会让测试全局登录访问非 hash 路由 `/auth/login` 并出现重定向循环，因此宿主入口验证使用专用 `E2E_HOST_BASE_URL`。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 于 `hack/tests` 通过。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH pnpm -F @lina/web-antd i18n:check` 于 `apps/lina-vben` 通过；本模块仍按用户要求中文-only，未新增 i18n 资源。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] `psql` 连续执行 `manifest/sql/mock-data/001-media-mock-data.sql` 两次通过，演示数据不重复写入且不再与全局策略唯一约束冲突。
- [x] `psql` 检查确认本地演示数据恢复为仅 `默认直播录制策略` 一条全局策略。
- [x] 静态扫描 `rg -n 'host_tenant_id|bizTenantId|BizTenantId|created_at|updated_at|createdAt|updatedAt|media/bindings|BindingScope|CodeMediaBindingScopeInvalid' apps/lina-plugins/media -g '!backend/internal/dao/**' -g '!backend/internal/model/**'` 无命中，确认旧字段、旧接口和旧时间命名无残留。
- [x] `make stop && make dev` 通过，宿主内嵌静态资源已重新生成；`apps/lina-core/internal/packed/public` 可命中 `media-strategy-toggle` 和 `media-device-binding-delete` 等最新页面标识。

## Review 验证记录

- [x] 按 `media_v2.md` 原始表结构恢复媒体策略 `create_time`/`update_time` 与流别名 `create_time` 字段；策略 `update_time` 通过 PostgreSQL 触发器自动维护，流别名不再维护原表不存在的更新时间字段。
- [x] 将策略绑定保存接口按资源拆分为 `PUT /media/device-bindings/{deviceId}`、`PUT /media/tenant-bindings/{tenantId}`、`PUT /media/tenant-device-bindings/{tenantId}/{deviceId}`，匹配按自然键创建或替换绑定的 RESTful 更新语义。
- [x] 补充 `TC0234-media-plugin-smoke.ts` API 场景，覆盖三个独立绑定资源、策略绑定优先级解析、策略引用保护和流别名新增/更新/详情/删除。
- [x] `go test ./...` 于 `apps/lina-plugins/media` 通过。
- [x] `corepack pnpm -F @lina/web-antd typecheck` 通过。
- [x] `corepack pnpm -F @lina/web-antd i18n:check` 通过；本模块仍未新增运行时 i18n、manifest i18n 或 apidoc i18n 资源。
- [x] `./node_modules/.bin/tsc --noEmit --pretty false` 于 `hack/tests` 通过。
- [x] `./node_modules/.bin/playwright test apps/lina-plugins/media/hack/tests/e2e/TC0234-media-plugin-smoke.ts` 通过，2 条用例全部通过。
- [x] `PATH=/Users/wanna/Library/pnpm:$PATH node ./scripts/validate-e2e.mjs` 通过。
- [x] `openspec validate add-media-plugin --strict` 通过。
- [x] `git diff --check` 通过。
- [x] 审查确认媒体配置仍为平台共享，不引入宿主租户隔离和数据权限过滤；权威边界由 `media:management:*` 菜单权限控制。
- [x] 审查确认本轮未新增缓存，策略解析和页面列表均直接读取 PostgreSQL；不存在跨实例缓存一致性影响。
- [x] 审查确认本模块按用户要求中文-only，未新增运行时 i18n、manifest i18n 或 apidoc i18n 资源。
