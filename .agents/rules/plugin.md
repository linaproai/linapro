# 业务插件开发规则

## 业务插件设计要求

在新开发/设计业务插件时，应当和用户明确以下关键需求：

- **是否需要多语言支持**：以便确定`plugin.yaml`中是否需要声明`i18n`配置，编码时是否需要考虑`manifest/i18n/<locale>/`目录下的多语言资源维护。
- **是否需要支持多租户**：以便确定设计数据表的时候是需要增加`tenant_id`字段，编码时是否需要使用到多租户的领域能力。
- **使用源码插件还是动态插件**：源码插件以源码形式产出，随宿主编译和发布；动态插件以`WASM`形式产出，运行时加载，用户不可见源码。
- **是否随框架共同编译和发布**：以便确定分发模式`distribution`是`managed`还是`builtin`。

## 插件通用资源要求

插件ID命名规范：`<author>-<domain>-<capability>` 
- `<author>`: 插件的作者或组织名称，建议使用小写字母和数字的组合，长度不超过 `20` 字符。
- `<domain>`: 插件所属的功能领域或业务域，建议使用小写字母和数字的组合，长度不超过 `20` 字符。
- `<capability>`: 插件提供的具体能力或功能，建议使用小写字母和数字的组合，长度不超过 `20` 字符。
- 插件 ID 必须唯一，且在整个 LinaPro 插件生态中保持稳定。插件 ID 会同时进入 URL path、动态资源路径、文件名、数据库键、菜单 key、权限字符串、`i18n` namespace、`apidoc` namespace 和宿主能力发现，因此必须严格遵守上述命名规范，避免使用特殊字符、空格或过长的名称，以确保插件能够正确识别和加载。

插件目录结构规范：
```text
apps/lina-plugins/<plugin-id>/
├── plugin.yaml                      # 插件元数据与能力声明
├── plugin_embed.go                  # 插件源码嵌入宿主编译入口，源码插件必须维护
├── Makefile                         # 插件make指令入口
├── backend/                         # 插件后端源码
│   ├── api/                         # API DTO与路由契约
│   ├── cap/                         # 插件拥有领域能力公开契约，可选，仅用于 plugin-owned 能力
│   ├── internal/                    # 插件内部业务逻辑封装
│   │   ├── controller/              # HTTP控制器
│   │   ├── service/                 # 业务服务层
│   │   ├── dao/                     # make dao生成
│   │   └── model/                   # do/entity模型
│   ├── pkg/                         # 非标准目录，不承载跨插件领域能力入口
│   └── plugin.go                    # 插件注册入口
├── frontend/                        # 插件前端资源
│   ├── pages/                       # 插件页面
│   ├── slots/                       # 插槽页面，可选
│   └── icons/                       # 可选：侧栏菜单自定义 SVG
├── hack/                            # 插件自身脚本和工具
│   ├── config.yaml                  # 插件开发期工具配置入口，包含代码生成、自定义构建等配置
│   └── tests/                       # 插件测试内容
│       └── e2e/                     # 插件 e2e 测试内容
├── manifest/                        # 插件清单与资源
│   ├── config/                      # 插件运行期配置
│   │   ├── config.yaml              # 开发期默认配置
│   │   └── config.example.yaml      # 配置模板，不作为运行时默认值
│   ├── sql/                         # 安装与升级SQL
│   │   ├── mock-data/               # 演示数据，可选
│   │   └── uninstall/               # 卸载SQL
│   └── i18n/                        # 插件语言包
├── README.md                        # 插件说明文档
└── README.zh-CN.md                  # 插件中文说明文档
```

开发业务插件，插件的目录结构和代码结构需参考以下插件：
- 源码插件需要参考官方提供的 `linapro-demo-source` 插件
- 动态插件需要参考官方提供的 `linapro-demo-dynamic` 插件

源码插件和动态插件都必须遵守以下通用资源约定：
- 插件源码目录统一放在`apps/lina-plugins/<plugin-id>/`下，`<plugin-id>`必须与`plugin.yaml`中的`id`一致。
- 插件多语言资源放在`manifest/i18n/<locale>/`，API 文档翻译资源放在`manifest/i18n/<locale>/apidoc/`。
- 插件拥有非核心领域能力时，公开契约统一放在`backend/cap/<domain>cap`，动态`guest SDK`放在该能力契约下的`bridge`或等价子包，`provider SPI`放在`spi`或等价子包。
- 插件 SQL 必须遵守`.agents/rules/database.md`。
- 插件自有业务表名必须使用完整插件 ID 前缀：`plugin_<plugin_id_snake>_<entity>`（将`plugin.yaml`的`id`中`-`替换为`_`且不得截断 ID 段）。细则与正反例见`.agents/rules/database.md`「插件业务表命名要求」。
- 插件 i18n 资源必须遵守`.agents/rules/i18n.md`。
- 插件开发期工具配置统一维护在插件根`hack/config.yaml`，包括代码生成、自定义构建等插件本地工具配置。
- 插件自定义构建指令统一放在插件根`hack/config.yaml`的`build.commands`下，由仓库根`make build`或`linactl build`读取执行。

## 领域能力使用要求

- 依赖领域能力的宽接口而非窄接口。例如，插件需要依赖`tenantcap.Service`，而不是依赖`tenantcap.FilterService`或者`tenantcap.PluginService`窄接口。
- 非核心`plugin-owned`领域能力的跨插件公开契约必须位于`owner`插件`backend/cap/<domain>cap`下。源码消费插件只能`import owner`插件的`backend/cap/...`公开契约，不得`import`目标插件`backend/internal/...`、`backend/internal/dao`、`backend/internal/model`、`controller`、`service`、`provider adapter`、私有缓存或`backend/pkg`作为领域能力入口。
- 源码插件生产代码`import`其他插件`backend/cap/...`时，必须在自身`plugin.yaml dependencies.plugins`声明对应`owner`插件和版本约束，并在自身`go.mod`中声明匹配的`Go module`依赖。
- 动态插件申请`owner`插件能力时，必须在`plugin.yaml hostServices`中显式声明`owner`、`service`、`version`和`methods`，并在`dependencies.plugins`中声明对应`owner`插件。不得使用`dependencies.capabilities`、软依赖或自动安装策略表达`owner`能力依赖。
- `owner`插件应通过类型安全`helper`生成通用`capability descriptor`。`lina-core`只接收`descriptor`、执行依赖治理、授权快照、动态路由和审计，不为每个非核心领域新增领域专属`*cap`包、`provider facade`、`wire codec`或`dispatcher`分支。

## 插件后端开发结构要求

源码插件和动态插件必须保持一致的后端业务开发结构，以降低开发者学习、迁移和维护成本。两类插件必须遵守以下结构：

- 每个插件必须同时维护`plugin.yaml`、`backend/`、`frontend/`与`manifest/`。
- 除`backend/cap`用于发布`plugin-owned`领域公开契约外，禁止在`backend`目录下创建后端公共组件或者目录。插件后端的业务模块应当严格封装在`backend/internal/service/`目录下。
- `backend/cap`只允许承载公开契约、`DTO`、值对象、错误语义、动态`bridge SDK`、`provider SPI`和`descriptor helper`，不得包含`DAO`、`DO`、`Entity`、`controller`、业务`service`实现、`provider`密钥、私有缓存结构或宿主`dispatcher`。
- `backend/plugin.go`用于声明插件后端入口、路由注册、生命周期接入或动态路由桥接入口。
- 插件后端 Go 代码必须遵守`.agents/rules/backend-go.md`。

## 插件数据库访问要求

- 涉及数据库访问的插件应在插件根`hack/config.yaml`中维护`gfcli.gen.dao`等代码生成工具配置，GoFrame 生成工作目录仍为插件`backend/`。
- 禁止插件重新依赖宿主的`dao/do/entity`生成工件。
- 禁止插件重新依赖其他插件的`dao/do/entity`生成工件。跨插件业务调用必须通过`owner`插件`backend/cap`公开契约、受治理`capability descriptor`或动态`hostServices`协议完成。
- 动态插件涉及宿主数据访问时，必须通过`plugin.yaml`的`hostServices`资源边界和宿主授权的`host service`协议。
- 动态插件涉及`owner`插件领域能力时，必须通过`owner-aware hostServices`声明和`owner`插件发布的`bridge SDK`访问，不得把动态`guest SDK`当作绕过宿主授权、依赖检查、租户和数据权限边界的直连入口。

## 源码插件对接要求

源码插件是随宿主源码编译和嵌入交付的插件。源码插件必须遵守以下对接要求：

- 源码插件必须维护`plugin_embed.go`作为宿主编译嵌入和静态资源装配入口。
- 源码插件应通过 registrar 或等价上下文把`backend/plugin.go`中声明的 controller、service、路由、中间件和生命周期能力接入宿主。
- 源码插件 provider/adapter 只能承载宿主稳定能力接缝实现，业务编排和领域逻辑仍必须放在`backend/internal/service/`。

## 插件信任与能力对等

- 源码插件与动态插件在经宿主安装或升级治理、并处于启用状态后，适用**同权、同信任级**：不得仅因插件 `type=dynamic` 而永久拒绝发布某一领域能力。
- 能力可用性由安装/启用状态、依赖满足、`hostServices`（或等价）声明与授权、以及方法级校验共同决定，与插件类型无关。

## 动态插件对接要求

动态插件是以运行时 WASM artifact 交付和加载的插件。动态插件必须保持与源码插件一致的后端业务开发结构，并额外遵守以下运行时对接结构：

- 动态插件源码目录应维护`main.go`作为 WASM guest 构建入口。
- 动态插件的 controller和service 是 guest 内部开发分层，宿主不得把它们当作源码插件原生 controller和service 直接加载；宿主只能通过`pluginbridge`、WASM host call 或版本化 host service 协议与动态插件交互。
- 动态插件涉及 Go guest 代码、WASM host service、host call 协议或插件桥接时，必须遵守`.agents/rules/backend-go.md`中关于动态插件 host service、WASM host service、错误处理和共享实例的要求。
