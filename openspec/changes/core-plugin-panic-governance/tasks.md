## 1. Panic 治理扫描改造

- [x] 1.1 在 `apps/lina-core/internal/cmd` 测试中实现插件 `init` 注册 fail-fast 的 AST 识别 helper（限 `apps/lina-plugins/*/backend/plugin.go` + `init` + `panic(err)` 类模式）
- [x] 1.2 调整 `TestProductionPanicsMatchAllowlist` 匹配逻辑：宿主精确白名单 + 插件模式自动放行；stale 检查仅针对宿主条目
- [x] 1.3 从 `productionPanicPolicy.Allowances` 删除全部 `apps/lina-plugins/...` 枚举条目，保留 `apps/lina-core` 精确条目
- [x] 1.4 补充 helper 单测或等价反向断言：字面量 panic / 非常规模式不得被自动放行

## 2. 错误文案能力化（P0-C）

- [x] 2.1 修改 `CodePluginTenantProvisioningPolicyInvalid` 默认英文源文案为 multi-tenant / framework tenant governance 语义（错误码不变）
- [x] 2.2 同步 `apps/lina-core/manifest/i18n/**/error.json` 中对应翻译键
- [x] 2.3 更新依赖该文案子串的单元测试断言

## 3. 验证与门禁

- [x] 3.1 运行 `apps/lina-core/internal/cmd` 包相关测试（官方插件工作区就绪时包含 panic 治理）
- [x] 3.2 运行受影响的 plugin 服务测试（含 tenant provisioning policy 错误断言）
- [x] 3.3 静态检索：`cmd_test.go` 不再出现 `apps/lina-plugins/linapro-` allowlist 路径；生产错误定义/i18n 中 provisioning 文案不再含 `linapro-tenant-core`
- [x] 3.4 记录影响分析：i18n 有影响（已同步）；缓存 / 数据权限 / 跨平台工具 / DI 运行期依赖无影响
- [x] 3.5 `openspec validate core-plugin-panic-governance --strict` 通过后执行 `lina-review`

### 影响分析记录（3.4）

| 领域 | 判断 |
| --- | --- |
| i18n | 有影响：`PLUGIN_TENANT_PROVISIONING_POLICY_INVALID` 英文源文案与 `en-US/error.json` 已改为 multi-tenant governance；`zh-CN` 原文案已是多租户治理语义，无需改键 |
| 缓存 | 无影响 |
| 数据权限 | 无影响 |
| 跨平台工具 | 无影响 |
| DI / 运行期依赖 | 无新增依赖；仅测试 AST 扫描与错误文案 |
| E2E | 未触发：无用户可观察页面行为变更，仅错误自然语言与测试治理 |
