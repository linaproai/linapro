## ADDED Requirements

### Requirement: 接口文档请求地址前缀必须按访问入口动态生成
系统 SHALL 在返回宿主 `/api.json` OpenAPI 文档时，根据当前请求的 scheme、host 和端口生成 `servers[0].url`，确保接口文档中每个接口展示的请求地址前缀和 Try It 请求目标与当前部署入口匹配。该运行时地址不得依赖固定写死的 `localhost` 地址。

#### Scenario: 通过前端开发服务代理访问接口文档
- **当** 管理员通过前端开发服务打开系统接口文档页面，且 `/api.json` 由前端服务代理到后端服务
- **则** `/api.json` 返回的 `servers[0].url` 使用后端服务收到请求时的 scheme、host 和端口
- **且** 接口文档中展示和 Try It 使用的请求地址前缀指向后端服务地址

#### Scenario: 通过后端服务直连访问接口文档
- **当** 管理员直接通过后端服务暴露地址打开系统接口文档页面，例如容器将 `8080` 映射为宿主机 `8088`
- **则** `/api.json` 返回的 `servers[0].url` 与浏览器访问后端服务时的 scheme、host 和端口保持一致
- **且** 接口文档不再展示固定的 `http://localhost:8080` 前缀

#### Scenario: 通过 HTTPS 代理访问接口文档
- **当** 管理员通过设置了 `X-Forwarded-Proto: https` 的代理请求 `/api.json`
- **则** `/api.json` 返回的 `servers[0].url` 使用 `https` scheme
- **且** host 和端口保留请求入口提供的值
