## Purpose
定义系统 OpenAPI 文档请求地址前缀的运行时生成要求，确保接口文档适配当前访问入口。
## Requirements
### Requirement: 接口文档请求地址前缀必须按访问入口动态生成
系统 SHALL 在返回宿主 `/api.json` OpenAPI 文档时，根据当前请求的 scheme、host 和端口生成 `servers[0].url`，确保接口文档中每个接口展示的请求地址前缀和 Try It 请求目标与当前部署入口匹配。该运行时地址不得依赖固定写死的 `localhost` 地址。

#### Scenario: 通过前端开发服务代理访问接口文档
- **当** 管理员通过前端开发服务打开系统接口文档页面，且 `/api.json` 由前端服务代理到后端服务
- **则** `/api.json` 返回的 `servers[0].url` 使用后端服务收到请求时的 scheme、host 和端口
- **且** 接口文档中展示和 Try It 使用的请求地址前缀指向后端服务地址

#### Scenario: 通过后端服务直连访问接口文档
- **当** 管理员直接通过后端服务暴露地址打开系统接口文档页面，例如容器将 `9120` 映射为宿主机 `8088`
- **则** `/api.json` 返回的 `servers[0].url` 与浏览器访问后端服务时的 scheme、host 和端口保持一致
- **且** 接口文档不再展示固定的 `http://localhost:9120` 前缀

#### Scenario: 通过 HTTPS 代理访问接口文档
- **当** 管理员通过设置了 `X-Forwarded-Proto: https` 的代理请求 `/api.json`
- **则** `/api.json` 返回的 `servers[0].url` 使用 `https` scheme
- **且** host 和端口保留请求入口提供的值

### Requirement: 接口文档页面必须展示加载状态
系统 SHALL 在接口文档 iframe 加载 Stoplight Elements 脚本和 `/api.json` 文档内容期间展示可见 Loading 状态，避免接口数量较多时出现长时间空白页面。

#### Scenario: 接口文档内容仍在加载
- **当** 管理员打开系统接口文档页面
- **且** iframe 内的 Stoplight Elements 尚未完成接口索引渲染
- **则** iframe 内展示加载中的状态提示
- **且** 加载状态根据当前语言环境展示对应语言文案

#### Scenario: 接口文档完成渲染
- **当** Stoplight Elements 已经渲染出接口文档侧边栏内容
- **则** iframe 内的加载状态自动隐藏
- **且** 管理员可以继续浏览接口文档和调试面板

### Requirement: 中文接口文档标题必须使用本地化文案
系统 SHALL 在中文语言环境生成 `/api.json` 时，使用稳定的接口文档本地化键翻译接口模块标题和接口摘要，不得因接口数量较多、接口描述重复或 GET/DELETE 接口缺少请求体而回退为英文标题。

#### Scenario: 静态接口描述重复
- **当** 多个静态接口使用相同的 `g.Meta dc` 描述
- **且** 管理员在中文语言环境请求 `/api.json`
- **则** 每个接口仍按其请求 DTO 对应的稳定本地化键翻译接口摘要
- **且** 返回的 OpenAPI 文档不包含内部本地化辅助字段

#### Scenario: 启用 i18n 的源码插件接口
- **当** 已启用 `i18n` 的源码插件接口挂载在 `/x/<plugin-id>/...` 命名空间下
- **且** 管理员在中文语言环境请求 `/api.json`
- **则** 该源码插件接口优先按请求 DTO 对应的插件自有稳定本地化键翻译接口模块标题和接口摘要
- **且** 不得因路径位于 `/x/<plugin-id>/...` 而错误回退到动态插件路径派生键

