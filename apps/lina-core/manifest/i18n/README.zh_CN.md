# 运行时国际化资源

该目录用于存放`LinaPro`交付项目的国际化基线资源。

宿主会把`manifest/i18n/<locale>/`目录下的直属`JSON`文件作为运行时`UI`语言包加载，并与已启用插件资源一起聚合，最终通过运行时国际化接口输出生效结果。

接口文档翻译资源存放在`manifest/i18n/<locale>/apidoc/**/*.json`。它们与运行时`UI`语言包共享同一个语言目录，便于发现和治理，但通过`apidoc/`子目录保持隔离，因为`OpenAPI`文档体量较大，且只在渲染`/api.json`时需要。

## 目录约定

| 路径                                                        | 用途               |
| ----------------------------------------------------------- | ------------------ |
| `manifest/config/config.template.yaml` 的 `i18n` 配置段       | 默认语言、多语言开关、排序、原生名 |
| `manifest/i18n/<locale>/*.json`                             | 宿主按语义域拆分的运行时语言包 |
| `manifest/i18n/<locale>/apidoc/**/*.json`                   | 宿主接口文档语言包 |
| `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/*.json` | 插件自有运行时语言包 |
| `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/apidoc/**/*.json` | 插件自有接口文档语言包 |

规则如下：

- 语言目录名必须使用规范化语言编码，例如`zh-CN`、`en-US`。
- 内置运行时语言从包含直属运行时`JSON`文件的`manifest/i18n/<locale>/`目录自动发现。
- 默认配置文件的 `i18n` 配置段维护无法安全从语言目录推导且部署方可能调整的元数据。运行时文本方向按当前宿主约定固定为 `ltr`。
- 新增内置语言不得要求修改 Go 常量、SQL seed 文件或前端 TypeScript 语言清单。
- 宿主只把`manifest/i18n/<locale>/*.json`这类直属文件识别为运行时语言包。
- 宿主只把`manifest/i18n/<locale>/apidoc/**/*.json`识别为接口文档语言包。
- 运行时 UI 消息文件可使用层级 JSON 或扁平 dotted key 编写。
- 宿主会把运行时 UI 消息文件归一化为扁平 key，用于聚合、缺失检查、导出、来源诊断和插件打包。
- 只有在返回前端运行时国际化接口结果时，宿主才会把归一化后的扁平 key 转换为嵌套对象。
- 接口文档语言包可使用层级 JSON 或扁平 dotted key 编写，并统一归一化为结构化 `core.*` 和 `plugins.*` 键；`en-US/apidoc`保持空占位，且不翻译`eg/example`示例值或生成`entity`元数据。
- 接口文档语言包可使用宿主自有的 `core.common.*` fallback key 维护标准响应、分页、时间字段等重复元数据；具体结构 key 存在时仍优先生效。
- 运行时业务文件使用`framework.json`、`menu.json`、`dict.json`、`config.json`、`error.json`、`artifact.json`、`job.json`、`notify.json`、`role.json`、`public-frontend.json`这类语义文件名。禁止使用数字顺序前缀。
- 宿主接口文档文件使用`common.json`或`core-api-<module>.json`命名；插件接口文档文件使用`plugin-api-<module>.json`或`<plugin-id>-api-<module>.json`命名。

## 为什么使用`JSON`和 key 归一化

`JSON`作为交付格式，是因为它与现有前端语言包工作流一致，便于通过`HTTP API`导出和离线校对，也便于在动态插件`Wasm`产物中直接嵌入，不需要额外转换层。

层级 JSON 是运行时 UI 消息推荐使用的文件编写格式，因为它可以减少重复前缀，并提升代码审查时的可读性。扁平 dotted key 仍然被接受，适合小范围补丁和渐进迁移。

扁平 key 仍然是系统治理格式，因为它可以让资源比对、缺失翻译检查、导出、来源诊断和插件打包都保持简单、稳定、可预测。同一个语言`JSON`文件中若同时存在层级`JSON`与等价扁平 dotted key，则扁平 dotted key 覆盖层级值，便于显式处理迁移差异。

示例：

```json
{
  "framework": {
    "description": "AI驱动的全栈开发框架"
  },
  "menu": {
    "dashboard": {
      "title": "工作台"
    }
  },
  "plugin": {
    "org-center": {
      "name": "组织中心"
    }
  }
}
```

## 键命名规范

| 范围           | `key`模式                                                     | 示例                                      |
| -------------- | ------------------------------------------------------------- | ----------------------------------------- |
| 框架元数据     | `framework.<field>`                                           | `framework.description`                   |
| 菜单标题       | `menu.<menu_key>.title`                                       | `menu.dashboard.title`                    |
| 字典类型名称   | `dict.<dict_type>.name`                                       | `dict.sys_normal_disable.name`            |
| 字典选项标签   | `dict.<dict_type>.<value>.label`                              | `dict.sys_normal_disable.1.label`         |
| 配置元数据     | `config.<config_key>.name`                                    | `config.sys.account.captchaEnabled.name`  |
| 公共前端文案   | `publicFrontend.<group>.<field>`                              | `publicFrontend.login.title`              |
| 插件名称       | `plugin.<plugin_id>.name`                                     | `plugin.org-center.name`                  |
| 插件描述       | `plugin.<plugin_id>.description`                              | `plugin.org-center.description`           |
| 语言显示名     | `locale.<locale>.name`                                        | `locale.en-US.name`                       |
| 语言原生名     | `locale.<locale>.nativeName`                                  | `locale.en-US.nativeName`                 |
| 校验或错误消息 | `validation.<module>.<field>.<rule>`或`error.<module>.<code>` | `validation.auth.login.username.required` |

建议：

- 优先复用稳定业务标识，例如`menu_key`、`dict_type`、`config_key`、`plugin_id`。
- 一个语义消息只维护一个`key`。
- 不要把界面层级直接映射成存储层级。`key`是稳定标识，不是`UI`树结构。
- 同一概念优先扩展已有前缀，不要重新造一套平行别名。

## 交付维护流程

1. 在`manifest/i18n/<locale>/`中新增或更新基线语言文件。
2. 当新语言需要启用、排序、原生名兜底或默认语言选择时，更新默认配置文件中的 `i18n.locales` 列表。
3. 当插件提供用户可见文案时，在`apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/`中补充对应语言文件。
4. 当`API DTO`源文案变化时，在宿主或插件自己的`manifest/i18n/<locale>/apidoc/`中补充接口文档语言文件。
5. 启动宿主后，请求`GET /api/v1/i18n/runtime/locales?lang=<locale>`确认语言列表和元数据。
6. 请求`GET /api/v1/i18n/runtime/messages?lang=<locale>`确认聚合后的运行时结果。
7. 请求 `/api.json?lang=<locale>` 确认接口文档本地化结果。
8. 使用`GET /api/v1/i18n/messages/missing?locale=<locale>`检查目标语言相对默认语言仍缺失的翻译键。
9. 使用`GET /api/v1/i18n/messages/diagnostics?locale=<locale>`确认当前生效值来自宿主文件还是插件文件。
10. 当需要离线校对或维护资源文件时，使用`GET /api/v1/i18n/messages/export?locale=<locale>`导出合并后的扁平 key 目录。

## 运行时接口缓存

`GET /api/v1/i18n/runtime/messages?lang=<locale>`会返回聚合后的运行时语言包，并通过`ETag`响应头标识当前语言和运行时语言包版本。

客户端应按语言持久化`{etag, messages, savedAt}`，下一次请求时带上`If-None-Match`。当服务端返回`304 Not Modified`时，响应体为空，客户端复用本地持久化语言包。

运行时语言包失效必须带明确`scope`。宿主文件、源码插件和动态插件变更，应只失效受影响的语言、扇区或插件。整包清空只适用于进程级重载和测试清理。

## 源码文案命名空间

部分运行时 key 由代码源码文案拥有，例如内置调度任务标签。拥有该文案的业务包必须通过`i18n.RegisterSourceTextNamespace(prefix, reason)`注册对应前缀。

`i18n`基础服务不得硬编码业务前缀。缺失翻译检查只会跳过已由所属模块注册的源码文案命名空间。

## 校验规则

交付前建议至少检查以下内容：

- 每个已启用语言文件都必须是合法的`JSON`。
- 默认配置文件`i18n.locales`中需要暴露的每个语言都必须存在对应的`manifest/i18n/<locale>/`运行时目录。
- 同一语言目录直属运行时`JSON`文件中归一化后的消息`key`必须保持唯一。
- 目标语言相对默认语言的缺失翻译检查必须通过。
- 插件自有文案默认使用`plugin.<plugin_id>.`前缀，除非该插件明确提供共享框架元数据。
- 新增的后端用户可见错误消息和校验消息必须使用翻译键，而不是直接硬编码文案。

## 编写示例

```json
{
  "framework": {
    "description": "AI驱动的全栈开发框架"
  },
  "menu": {
    "system": {
      "title": "系统管理",
      "users": {
        "title": "用户管理"
      }
    }
  },
  "config": {
    "sys": {
      "account": {
        "captchaEnabled": {
          "name": "登录验证码"
        }
      }
    }
  },
  "publicFrontend": {
    "login": {
      "title": "欢迎回来"
    }
  },
  "plugin": {
    "org-center": {
      "name": "组织中心",
      "description": "部门、岗位与层级治理"
    }
  }
}
```
