# 设计

## 问题

`apps/lina-core/manifest/config/metadata.yaml` 中的 `openapi.serverUrl` 当前固定为 `http://localhost:8080`。`internal/service/apidoc` 在构建文档时把该值写入 OpenAPI `servers`，Stoplight Elements 会把 `servers[0].url` 作为接口地址前缀和 Try It 请求目标。

固定地址在以下场景会失效：

- 容器将后端 `8080` 映射到宿主机 `8088`，浏览器通过 `http://host:8088` 访问后端直连页面。
- 后端部署在域名或 HTTPS 代理后，浏览器访问的 scheme/host 不等于 `localhost:8080`。
- 本地开发时，前端服务通过 Vite 代理访问后端，`/api.json` 实际由后端 `8080` 返回，此时接口文档应继续显示后端地址而不是前端服务端口。

## 方案

1. 保留 `apidoc.Build(ctx, server)` 的职责：构建路由、插件投影、文档标题、描述、版本和服务描述。
2. 在 `/api.json` HTTP handler 中，在文档构建完成且写出 JSON 前，按当前请求生成 origin 并覆盖 `document.Servers`。
3. origin 生成规则：
   - scheme 使用 GoFrame 请求的 `GetSchema()`，支持 `X-Forwarded-Proto` 与 TLS。
   - host 使用原始 `r.Host`，保留端口，避免 `GetHost()` 去掉端口。
   - 当请求 host 为空时不覆盖 `servers`，保留构建结果作为降级。
4. `metadata.yaml` 中保留 `serverDescription` 作为 `servers[0].description`，但不再依赖 `serverUrl` 作为运行时地址来源。
5. Stoplight 静态页继续使用相对路径 `/api.json?lang=...` 加载文档；前端代理访问和后端直连访问都由当前请求自然决定 `/api.json` 的 origin。

## 影响评估

- i18n：不新增、修改或删除运行时语言包、插件 manifest i18n 或 apidoc i18n JSON；本变更不修改 API DTO 文档源文本。
- 数据权限：未新增或修改数据操作接口；`/api.json` 已是接口文档读取端点，本变更只调整文档中的 server 地址。
- 缓存一致性：未新增缓存；`servers` 在每次 `/api.json` 请求中按请求动态覆盖，不跨请求、不跨实例保存状态，分布式部署中各实例均以自身收到的请求头为权威。
- 安全边界：生成地址只使用请求的 scheme 与 host，不引入用户可写的接口路径；部署在反向代理后如需对外地址，应由代理正确设置 Host 与 `X-Forwarded-Proto`。
