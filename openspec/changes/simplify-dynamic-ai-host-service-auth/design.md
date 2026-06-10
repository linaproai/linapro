## Context

`apps/lina-core`作为动态插件宿主，通过`plugin.yaml hostServices`、`pluginbridge`和`WASM host service`为动态插件暴露宿主能力。现有`ai`服务被建模为资源型 host service：插件需要声明`resources[].ref: purpose:<name>`，并可在资源属性中声明`defaultTier`、`maxOutputTokens`、`maxPayloadBytes`等策略。

这个模型把两类职责放在同一个清单中：插件是否能调用某个`AI`领域方法，以及该方法请求参数如何受限。前者属于动态插件安装和启用阶段的宿主能力授权；后者属于`AI`能力服务和智能中心的运行期策略。源码插件通过`ServicesForPlugin(...).AI().Text()`等类型化能力入口提交 DTO，不需要在插件清单中重复描述调用参数策略。动态插件也应保持同样的领域消费模型，只在 transport 和授权快照上不同。

## Goals / Non-Goals

**Goals:**

- 将动态插件`ai` host service 简化为`service + methods`授权。
- 移除`ai.resources`、`purpose`资源匹配和资源属性策略在主框架中的校验。
- 保持动态插件和源码插件一致的`AI` DTO 调用语义，由调用方显式提交`tier`和其他方法参数。
- 保留主框架对`service + method`、来源插件身份、DTO 编解码、结构化错误和错误脱敏的治理。
- 更新公开协议文档和示例，避免开发者继续使用旧的`ai.resources`格式。

**Non-Goals:**

- 不新增`AI`模型、渠道、档位、配额或限流数据模型。
- 不改变`linapro-ai-core`管理页面、HTTP API 或数据库结构。
- 不把`pluginbridge`改造成弱类型`AI`网关；仍调用`AI().Text()`、`AI().Document()`等类型化子能力。
- 不改变`storage`、`network`、`data`、`hostconfig`、`manifest`等仍需要资源边界的 host service。

## Decisions

### 1. `ai` host service 改为方法授权型

`ai`服务在 host service 目录中的资源类型改为`none`。声明`service: ai`时只允许`methods`，不得再声明`resources`、`paths`、`tables`或`keys`。

原因：动态插件访问`AI`能力的最小授权单位是领域方法，例如`text.generate`、`document.cite`。`purpose`是调用场景和审计字段，不应成为动态插件清单的授权资源。这样可以让插件清单更短，并与源码插件通过类型化能力服务提交 DTO 的模型一致。

替代方案是保留`resources`但设为可选。该方案会留下两套授权模型：有资源时按资源限制，无资源时按方法限制。对安装授权、运行时授权快照和文档说明都更复杂，也不符合本次简化目标。

### 2. 主框架不再限制`AI`业务参数

`WASM` host handler 不再读取`ai.resources`属性，不再做`purpose`匹配、`defaultTier`兜底、`maxOutputTokens`上限、payload 字节数、资产数量、mime 类型或 operation 开关校验。

原因：这些限制属于`AI`能力服务的业务策略。主框架的职责是 bridge transport 和方法授权，不应复制智能中心策略。请求仍会被类型化子能力服务校验，例如`purpose`必填、`tier`合法、资产引用非空、`maxOutputTokens`非负等。

### 3. guest SDK 调用`AI`时不再携带`resourceRef`

guest SDK 构造`ai` host service envelope 时，`resourceRef`保持为空，`purpose`仅出现在请求 payload 中。

原因：如果`resourceRef`继续携带`purpose:<name>`，运行时会暗示仍存在资源授权边界。清空`resourceRef`可以让协议语义与`plugin.yaml`保持一致。

### 4. 保留插件来源身份注入

动态插件调用`AI`能力时，宿主仍通过`Capability.ServicesForPlugin(..., pluginID)`注入可信来源插件身份。调用方不能在请求体中伪造`SourcePluginID`。

原因：来源身份用于调用日志、审计、用量归因和后续插件级治理。它是宿主上下文治理，不属于可由动态插件自行提交的请求参数。

### 5. 文档和示例同步检查

主框架插件 README 需要移除`ai.resources`示例。动态插件示例清单经静态检查未声明`ai`服务，因此不新增未使用授权；若示例插件未来需要调用`AI`，只保留`service: ai`和`methods`。

### 6. 动态插件普通领域能力补齐

动态插件`hostServices`目录补齐源码插件`capability.Services`普通消费面中的领域能力，包括`apidoc`、`auth`、`authz`、`user`、`bizctx`、`dict`、`file`、`i18n`、`infra`、`job`、`notification`、`plugin`、`route`和`session`。这些领域服务均采用方法授权型声明，不接受`resources`、`paths`、`tables`或`keys`，运行时统一从宿主启动期注入的同一个`capability.Services`目录进入对应`*cap.Service`或普通子服务。

已有`AI`、`Users`、`Org`、`Tenant`动态领域服务的解析路径优先使用共享领域目录，避免动态插件和源码插件分别维护平行能力来源。`notify.send`仍保留为资源型发送服务；普通通知读取通过新增`notification.messages.batch_get`进入`notifycap.Service`。插件配置读取继续作为插件作用域配置能力通过`Plugins().Config()`的`guest`入口使用，协议层保留`config.get`的只读 host service。

## Risks / Trade-offs

- [Risk] 插件安装时无法再从`plugin.yaml`直接看到`AI`调用场景和 token 上限。  
  Mitigation：通过方法授权、来源插件审计和`AI`调用日志保留治理证据；后续如需细粒度策略，应在`linapro-ai-core`中按插件、租户或策略配置实现。

- [Risk] 旧示例或测试仍携带`ai.resources`会被新校验拒绝。  
  Mitigation：同步更新示例、README 和单元测试，明确`ai`服务不接受资源声明。

- [Risk] 删除 host 层 payload 策略后，错误可能从`AI`能力服务返回而非 bridge 层返回。  
  Mitigation：保持结构化错误 envelope 和错误脱敏；测试覆盖未授权方法、禁止`ai.resources`、请求 DTO 参数直达能力服务。

## 影响分析

- `i18n`：新增动态`i18n`领域 host service 只透传宿主既有`i18ncap.Service`，不新增运行时 UI 文案、菜单、API 文档源文本或语言包；README 和 OpenSpec 属技术文档变更。确认无运行时`i18n`资源新增影响。
- 缓存一致性：不新增缓存、快照或失效机制；动态领域调用复用宿主启动期共享`capability.Services`和既有领域缓存/修订号策略。确认无新的缓存一致性机制影响。
- 数据权限：新增动态领域读取入口会进入现有`*cap.Service`，由领域实现继续执行租户、数据权限、可见性和批量上限；不开放`DAO/DO/Entity`、`gdb.Model`或核心表`data`访问。确认数据权限边界通过领域能力复用。
- 开发工具跨平台：不修改脚本、Makefile、CI 或`linactl`。确认无开发工具跨平台影响。
- 测试策略：使用`pluginbridge`和`WASM host service`单元测试覆盖清单校验、授权拒绝、共享领域目录解析、`CapabilityContext`传递和请求参数透传；不涉及用户可观察页面，确认无需新增 E2E。

## Migration Plan

1. 更新 OpenSpec 增量规范，明确`ai`不再是资源型 host service。
2. 更新`pluginbridge` host service 目录和清单校验。
3. 更新 guest SDK 和`WASM` host handler，移除`resourceRef`和资源属性策略。
4. 更新 README 和单元测试，并静态确认动态插件示例清单不含旧`ai.resources`声明。
5. 运行相关 Go 包测试和`openspec validate simplify-dynamic-ai-host-service-auth --strict`。
