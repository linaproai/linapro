# 业务插件开发规则

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
├── Makefile                         # 插件make指令入口
├── backend/
│   ├── api/                         # API DTO与路由契约
│   ├── internal/
│   │   ├── controller/              # HTTP控制器
│   │   ├── service/                 # 业务服务层
│   │   ├── dao/                     # make dao生成
│   │   └── model/                   # do/entity模型
│   └── plugin.go                    # 插件注册入口
├── frontend/
│   ├── pages/                       # 插件页面
│   └── slots/                       # 插槽页面，可选
├── hack/                            # 插件自身脚本和工具
│   ├── config.yaml                  # 插件开发期工具配置入口，包含代码生成、自定义构建等配置
│   └── tests/                       # 插件测试内容
│       └── e2e/                     # 插件 e2e 测试内容
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

开发业务插件，插件的目录结构和代码结构：
- 源码插件需要参考 `linapro-demo-source` 插件
- 动态插件需要参考 `linapro-demo-dynamic` 插件

源码插件和动态插件都必须遵守以下通用资源约定：

- 插件源码目录统一放在`apps/lina-plugins/<plugin-id>/`下，`<plugin-id>`必须与`plugin.yaml`中的`id`一致。
- 插件必须维护`plugin.yaml`和`manifest/`。
- 插件前端页面或公开静态资源统一放在`frontend/pages/`或由`plugin.yaml`的`public_assets`显式声明的目录下。
- 插件安装 SQL 放在`manifest/sql/`。
- 插件卸载 SQL 放在`manifest/sql/uninstall/`。
- 插件 Mock 数据 SQL 放在`manifest/sql/mock-data/`。
- 插件多语言资源放在`manifest/i18n/<locale>/`，API 文档翻译资源放在`manifest/i18n/<locale>/apidoc/`。
- 插件 SQL 必须遵守`.agents/rules/database.md`。
- 插件 i18n 资源必须遵守`.agents/rules/i18n.md`。
- 插件开发期工具配置统一维护在插件根`hack/config.yaml`，包括代码生成、自定义构建等插件本地工具配置。
- 插件自定义构建指令统一放在插件根`hack/config.yaml`的`build.commands`下，由仓库根`make build`或`linactl build`读取执行。

## 插件后端同构开发结构要求

源码插件和动态插件必须保持一致的后端业务开发结构，以降低开发者学习、迁移和维护成本。两类插件必须遵守以下结构：

- 每个插件必须同时维护`plugin.yaml`、`backend/`、`frontend/`与`manifest/`。
- 禁止在`backend`目录下创建后端公共组件或者目录，插件后端的业务模块应当严格封装在`backend/internal/service/`目录下。
- `backend/plugin.go`用于声明插件后端入口、路由注册、生命周期接入或动态路由桥接入口。
- 插件后端 Go 代码必须遵守`.agents/rules/backend-go.md`。

## 插件数据库访问要求

- 涉及数据库访问的插件应在插件根`hack/config.yaml`中维护`gfcli.gen.dao`等代码生成工具配置，GoFrame 生成工作目录仍为插件`backend/`。
- 禁止插件重新依赖宿主的`dao/do/entity`生成工件。
- 动态插件涉及宿主数据访问时，必须通过`plugin.yaml`的`hostServices`资源边界和宿主授权的 host service 协议。

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
