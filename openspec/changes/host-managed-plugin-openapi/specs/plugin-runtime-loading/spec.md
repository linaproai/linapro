## ADDED Requirements

### Requirement: 源码插件路由在注册时由宿主自动采集归属

系统 SHALL 在源码插件注册后端路由时由宿主自动采集路由归属、路径和文档元数据，而不是在生成系统接口文档时根据路径前缀或 `plugin.yaml` 清单推断。

#### Scenario: 源码插件注册任意路由地址

- **WHEN** 一个源码插件通过宿主发布的源码插件路由注册 facade 绑定后端路由
- **THEN** 该插件可以注册任意合法路由地址
- **AND** 宿主在注册时记录该路由所属的 `pluginID`
- **AND** 宿主不要求该路由使用固定前缀

#### Scenario: 标准 DTO 路由自动采集文档元数据

- **WHEN** 一个源码插件绑定的处理器使用标准 GoFrame DTO 签名 `func(ctx context.Context, req *Req) (res *Res, err error)`
- **THEN** 宿主从 `Req` 的 `g.Meta` 自动提取方法、路径、标签、摘要、描述和权限等文档元数据
- **AND** 宿主不要求开发者在 `plugin.yaml` 中重复声明这些路由信息

#### Scenario: 原始 handler 仍可注册但不自动文档化

- **WHEN** 一个源码插件绑定 `func(*ghttp.Request)` 这类原始 handler
- **THEN** 宿主仍记录该路由归属，避免把它误认为宿主静态接口
- **AND** 宿主默认不把该路由自动投影到 `OpenAPI` 文档

### Requirement: 源码插件中间件组合由插件自己维护

系统 SHALL 继续允许源码插件在注册代码中自行组合和维护中间件链，宿主只负责路由归属采集与真实绑定，不解释源码插件中间件的业务语义。

#### Scenario: 源码插件自定义中间件顺序

- **WHEN** 一个源码插件在路由注册时调用 `Middleware(...)` 组合宿主发布的原始 `ghttp.HandlerFunc`
- **THEN** 宿主按源码插件声明的顺序将这些中间件注册到底层路由组
- **AND** 宿主不把这些中间件转换为动态插件那种受控中间件描述符模型
- **AND** 源码插件仍然保留和宿主同进程开发时的中间件灵活性
