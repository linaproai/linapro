# 插件目录结构规则

## 适用范围

本规则约束 LinaPro 插件的通用资源目录、源码插件与动态插件共享的后端开发目录结构、源码插件编译嵌入对接、动态插件 WASM 运行时对接、动态插件产物资源视图、插件 manifest 资源、安装卸载 SQL、前端页面、宿主能力接缝和生命周期资源归属。

源码插件和动态插件必须共享插件清单、生命周期资源、SQL、i18n、前端静态资源和后端业务开发结构约定。两类插件的差异仅体现在与宿主的对接方式、运行时加载方式和交付形态上：源码插件随宿主编译嵌入，动态插件通过 WASM artifact、`pluginbridge`和`hostServices`协议接入。禁止让动态插件绕过统一的`api`、`controller`、`service`分层开发结构，也禁止让动态插件绕过通用插件资源约定。

## 文档同步要求

主框架插件能力发生变更时，必须审查`apps/lina-core/pkg/plugin`目录下的README文档是否需要同步更新。

## 插件通用资源要求

插件ID命名规范：`<author>-<domain>-<capability>` 
- `<author>`: 插件的作者或组织名称，建议使用小写字母和数字的组合，长度不超过 20 字符。
- `<domain>`: 插件所属的功能领域或业务域，建议使用小写字母和数字的组合，长度不超过 20 字符。
- `<capability>`: 插件提供的具体能力或功能，建议使用小写字母和数字的组合，长度不超过 20 字符。
- 插件 ID 必须唯一，且在整个 LinaPro 插件生态中保持稳定。插件 ID 会同时进入 URL path、动态资源路径、文件名、数据库键、菜单 key、权限字符串、`i18n` namespace、`apidoc` namespace 和宿主能力发现，因此必须严格遵守上述命名规范，避免使用特殊字符、空格或过长的名称，以确保插件能够正确识别和加载。

插件目录结构规范：
```text
apps/lina-plugins/<plugin-id>/
├── plugin.yaml
├── plugin_embed.go
├── backend/
│   ├── api/                         # API DTO与路由契约
│   ├── hack/
│   │   └── config.yaml              # make dao等插件开发配置
│   ├── internal/
│   │   ├── controller/              # HTTP控制器
│   │   ├── service/                 # 业务服务层
│   │   ├── dao/                     # make dao生成
│   │   └── model/                   # do/entity模型
│   └── plugin.go                    # 插件注册入口
├── frontend/
│   ├── pages/                       # 插件页面
│   └── slots/                       # 插槽页面，可选
├── manifest/
│   ├── config/
│   │   ├── config.yaml              # 开发期默认配置
│   │   └── config.example.yaml      # 配置模板，不作为运行时默认值
│   ├── sql/                         # 安装与升级SQL
│   │   ├── mock-data/               # 演示数据，可选
│   │   └── uninstall/               # 卸载SQL
│   └── i18n/                        # 插件语言包
├── README.md
└── README.zh-CN.md
```

源码插件和动态插件都必须遵守以下通用资源约定：

- 插件源码目录统一放在`apps/lina-plugins/<plugin-id>/`下，`<plugin-id>`必须与`plugin.yaml`中的`id`一致。
- 插件必须维护`plugin.yaml`和`manifest/`。
- 插件前端页面或公开静态资源统一放在`frontend/pages/`或由`plugin.yaml`的`public_assets`显式声明的目录下。
- 插件安装 SQL 放在`manifest/sql/`。
- 插件卸载 SQL 放在`manifest/sql/uninstall/`。
- 插件 Mock 数据 SQL 放在`manifest/sql/mock-data/`。
- 插件多语言资源放在`manifest/i18n/<locale>/`，API 文档翻译资源放在`manifest/i18n/<locale>/apidoc/`。
- 不得把插件生命周期资源回流到宿主目录中。
- 插件 SQL 必须遵守`.agents/rules/database.md`。
- 插件 i18n 资源必须遵守`.agents/rules/i18n.md`。

## 插件后端同构开发结构要求

源码插件和动态插件必须保持一致的后端业务开发结构，以降低开发者学习、迁移和维护成本。两类插件必须遵守以下结构：

- 每个插件必须同时维护`plugin.yaml`、`backend/`、`frontend/`与`manifest/`。
- 插件后端统一采用`backend/api/`、`backend/plugin.go`、`backend/internal/controller/`、`backend/internal/service/`结构。
- `backend/api/`用于声明构建期可解析的 API DTO、请求响应契约和路由元数据。
- `backend/plugin.go`用于声明插件后端入口、路由注册、生命周期接入或动态路由桥接入口。
- `backend/internal/controller/`用于承载插件侧请求处理、参数转换、调用服务和响应投影逻辑。动态插件中的 controller 是 WASM guest 内部的请求处理分层，不等同于宿主原生 controller，但目录和职责必须保持一致。
- `backend/internal/service/`用于承载插件业务编排、领域逻辑、中间件实现和对宿主能力接缝的调用。动态插件中的 service 通过`pluginbridge`、WASM host call 或版本化 host service 协议访问宿主能力。
- 禁止再将业务`service`目录直接放在`backend/service/`下。
- 插件业务编排、领域逻辑和中间件实现必须收敛到`backend/internal/service/`。
- 只有实现宿主稳定能力接缝的 provider/adapter 才允许放在`backend/provider/`等非`internal`目录中。
- 插件后端 Go 代码必须遵守`.agents/rules/backend-go.md`。

## 插件数据库访问要求

- 插件若需要自有数据库访问，必须在插件自己的`backend/`下维护`hack/config.yaml`。
- 插件的`make dao`生成结果必须放在`backend/internal/dao/`与`backend/internal/model/{do,entity}/`。
- 禁止插件重新依赖宿主的`dao/do/entity`生成工件。
- 源码插件和动态插件都不得把宿主`DAO`、`DO`、`Entity`、私有缓存快照、运行时状态或内部配置结构作为插件接口契约、服务参数或响应结构。
- 动态插件涉及宿主数据访问时，必须通过`plugin.yaml`的`hostServices`资源边界和宿主授权的 host service 协议，不得直接依赖宿主私有 DAO、DO 或 Entity 工件。

## 源码插件对接要求

源码插件是随宿主源码编译和嵌入交付的插件。源码插件必须遵守以下对接要求：

- 源码插件必须维护`plugin_embed.go`作为宿主编译嵌入和静态资源装配入口。
- 源码插件应通过 registrar 或等价上下文把`backend/plugin.go`中声明的 controller、service、路由、中间件和生命周期能力接入宿主。
- 源码插件 provider/adapter 只能承载宿主稳定能力接缝实现，业务编排和领域逻辑仍必须放在`backend/internal/service/`。

## 动态插件对接要求

动态插件是以运行时 WASM artifact 交付和加载的插件。动态插件必须保持与源码插件一致的后端业务开发结构，并额外遵守以下运行时对接结构：

- 动态插件源码目录应维护`main.go`作为 WASM guest 构建入口。
- 动态插件的 controller和service 是 guest 内部开发分层，宿主不得把它们当作源码插件原生 controller和service 直接加载；宿主只能通过`pluginbridge`、WASM host call 或版本化 host service 协议与动态插件交互。
- 动态插件涉及 Go guest 代码、WASM host service、host call 协议或插件桥接时，必须遵守`.agents/rules/backend-go.md`中关于动态插件 host service、WASM host service、错误处理和共享实例的要求。
