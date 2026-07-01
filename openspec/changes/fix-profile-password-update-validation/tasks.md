## 1. 接口校验修复

- [x] 1.1 移除`UpdateProfileReq.nickname`的必填校验，使个人中心仅提交`password`时能够通过请求校验。
- [x] 1.2 保持用户创建接口的`nickname`必填校验不变，避免放宽管理员创建用户契约。

## 2. 回归测试与验证

- [x] 2.1 增加 DTO 校验回归测试，覆盖`UpdateProfileReq`只提交`password`时校验通过，并覆盖`CreateReq`仍要求`nickname`。
- [x] 2.2 增加`hack/tests/e2e/iam/user/TC010-profile-password-update.ts`，覆盖个人中心修改密码只提交`password`字段且接口成功返回。
- [x] 2.3 运行`cd hack/tests && npx playwright test e2e/iam/user/TC010-profile-password-update.ts --project=chromium --trace=off --output=temp/tc010-results`。
- [x] 2.4 运行`cd apps/lina-core && go test ./api ./internal/cmd -count=1`。
- [x] 2.5 运行`openspec validate fix-profile-password-update-validation --strict`，并记录`i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响判断。

## 执行记录

- 根因：`UpdateProfileReq.nickname`复用了创建用户场景的必填校验，但个人中心资料更新服务层已经按字段执行局部更新，导致仅修改密码的请求在进入服务层前被校验拦截。
- 实现：移除`UpdateProfileReq.nickname`的`required`校验；管理员创建用户的`CreateReq.nickname`必填约束保持不变。
- API 生成：路由、HTTP 方法、控制器签名和 DTO 字段集合未变化，仅调整校验标签，不需要运行`make ctrl`。
- DI 来源检查：无新增运行期依赖、构造函数、启动装配或共享实例变更。
- 接口性能：不涉及列表、批量、树形、聚合或高数据量装配路径，不产生前端瀑布式调用或后端`N+1`查询风险。
- 数据权限：当前用户自服务更新仍基于登录上下文更新本人资料，不新增跨用户读取或写入路径；E2E 临时用户由管理员 API 创建并在用例内清理。
- `i18n`：未新增运行时文案、菜单、路由、按钮、语言包或 API 文档源文本；仅移除已有昵称必填校验触发点。
- 缓存一致性：不涉及缓存读写、派生状态、失效或分布式协调。
- 开发工具跨平台：未修改脚本、构建工具、CI 或跨平台入口。
- 测试策略：新增 API DTO 校验测试和`TC010`E2E；截图证据位于`temp/20260701/210352-profile-password-form.png`与`temp/20260701/210352-profile-password-success.png`。

## 验证记录

- `cd apps/lina-core && go test ./api ./internal/cmd -count=1`：通过。
- `cd hack/tests && npm run test:validate`：通过，校验`252`个 E2E 文件。
- `cd hack/tests && npx playwright test e2e/iam/user/TC010-profile-password-update.ts --project=chromium --trace=off --output=temp/tc010-results`：通过，`1 passed`。
- `openspec validate fix-profile-password-update-validation --strict`：通过。
- `git diff --check`：通过。
