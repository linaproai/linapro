# 任务

## 1. OpenSpec 记录

- [x] 1.1 创建本变更的 `proposal.md`、`design.md`、`tasks.md` 与增量规范
- [x] 1.2 明确记录 i18n、数据权限和缓存一致性影响判断

## 2. 后端实现

- [x] 2.1 在宿主 `/api.json` handler 中按当前请求 origin 覆盖 OpenAPI `servers[0].url`
- [x] 2.2 保留 `serverDescription` 作为服务地址描述，并在 host 缺失时保持安全降级
- [x] 2.3 移除对固定 `openapi.serverUrl` 作为运行时请求地址权威来源的依赖

## 3. 测试

- [x] 3.1 增加后端单元测试，覆盖直连映射端口、前端代理到后端端口、`X-Forwarded-Proto=https` 三类 origin
- [x] 3.2 新增 E2E 用例 `TC0175-api-docs-request-origin.ts`，验证 `/api.json` 的 `servers[0].url` 随前端代理访问和后端直连访问动态变化
- [x] 3.3 运行新增/相关后端单元测试与 E2E 测试

## 4. 审查

- [x] 4.1 运行 `openspec validate fix-apidoc-request-origin --strict`
- [x] 4.2 调用 `lina-review` 完成代码和规范审查

## 审查记录

- `lina-review` 审查结论：未发现阻断问题。本变更仅调整宿主 `/api.json` OpenAPI `servers` 地址生成逻辑、OpenAPI 元数据默认值、对应后端单元测试和接口文档 E2E；未新增或修改数据操作接口，不涉及角色数据权限接入变化；未修改 API DTO 文档源文本、运行时语言包、插件 manifest i18n 或 apidoc i18n JSON；未新增缓存，`servers` 按每次请求动态生成，不跨请求或跨实例保存状态。

## Feedback

- [x] **FB-1**: 接口文档中每个接口展示和 Try It 使用的请求地址前缀不得固定为 `http://localhost:8080`，应根据前端代理访问或后端直连访问的实际地址自动生成
