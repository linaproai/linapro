## 1. 规范与文档

- [x] 1.1 更新 `apps/lina-core/pkg/plugin/README.md` 与 `README.zh-CN.md`：同权同信 + 动态外部登录发布说明
- [x] 1.2 更新 `.agents/rules/plugin.md` 或 `architecture.md` 中与「动态/源码信任」冲突的表述（若有）
- [x] 1.3 修正 `domainhostcall` / 旧注释中「永久不对动态开放」措辞

## 2. Protocol 与 catalog

- [x] 2.1 新增 wire 常量：`external_login.login_by_verified_identity`、`users.create_from_external`
- [x] 2.2 登记到 hostservices catalog（auth / users 方法列表）
- [x] 2.3 同步 `protocol_hostservice_contract.go` 导出别名

## 3. WASM 分发与 guest

- [x] 3.1 `wasm_host_service_auth.go`：分发外部登录；校验 resource ownership；盖章 pluginID 调 auth
- [x] 3.2 `wasm_host_service_users.go`：分发 CreateFromExternal
- [x] 3.3 `wasm_host_service_registry.go`：注册新方法
- [x] 3.4 `domainhostcall_auth.go` / `domainhostcall_users.go`：真实 host call
- [x] 3.5 更新/替换 fail-closed 单测

## 4. 验证

- [x] 4.1 相关 go test 通过
- [x] 4.2 `openspec validate publish-external-login-to-dynamic-plugins --strict`
