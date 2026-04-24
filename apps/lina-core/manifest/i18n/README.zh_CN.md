# 运行时国际化资源

该目录用于存放`LinaPro`交付项目的运行时多语言基线消息包。

宿主会把`manifest/i18n/<locale>.json`作为项目级基线资源加载，并与已启用插件资源、数据库覆写一起聚合，最终通过运行时国际化接口输出生效结果。

## 目录约定

| 路径                                                        | 用途               |
| ----------------------------------------------------------- | ------------------ |
| `manifest/i18n/zh-CN.json`                                  | 简体中文基线语言包 |
| `manifest/i18n/en-US.json`                                  | 英文基线语言包     |
| `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>.json` | 插件自有语言包     |

规则如下：

- 文件名必须使用规范化语言编码，例如`zh-CN.json`、`en-US.json`。
- 宿主只把顶层`manifest/i18n/<locale>.json`识别为运行时语言包。
- 运行时消息统一使用扁平`key`维护。
- 只有在返回前端运行时国际化接口结果时，宿主才会把扁平`key`转换为嵌套对象。

## 为什么使用`JSON`和扁平`key`

`JSON`作为交付格式，是因为它与现有前端语言包工作流一致，便于通过`HTTP API`做导入导出，也便于在动态插件`Wasm`产物中直接嵌入，不需要额外转换层。

扁平`key`作为唯一管理格式，是因为它可以让后端存储、数据库覆写、缺失翻译检查和插件打包都保持简单、稳定、可预测。

示例：

```json
{
  "framework.description": "AI驱动的全栈开发框架",
  "menu.dashboard.title": "工作台",
  "plugin.org-center.name": "组织中心"
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

1. 在`manifest/i18n/`中新增或更新基线语言文件。
2. 当插件提供用户可见文案时，在`apps/lina-plugins/<plugin-id>/manifest/i18n/`中补充对应语言文件。
3. 启动宿主后，请求`GET /api/v1/i18n/runtime/messages?lang=<locale>`确认聚合后的运行时结果。
4. 使用`GET /api/v1/i18n/messages/missing?locale=<locale>`检查目标语言相对默认语言仍缺失的翻译键。
5. 使用`GET /api/v1/i18n/messages/diagnostics?locale=<locale>`确认当前生效值来自宿主文件、插件文件还是数据库覆写。
6. 如果线上需要热修正文案，可通过`POST /api/v1/i18n/messages/import`导入数据库覆写，再通过`GET /api/v1/i18n/messages/export`导出生效结果回写代码库。

## 校验规则

交付前建议至少检查以下内容：

- 每个已启用语言文件都必须是合法的`JSON`。
- 同一语言文件内的消息`key`必须保持唯一。
- 目标语言相对默认语言的缺失翻译检查必须通过。
- 插件自有文案默认使用`plugin.<plugin_id>.`前缀，除非该插件明确提供共享框架元数据。
- 新增的后端用户可见错误消息和校验消息必须使用翻译键，而不是直接硬编码文案。

## 业务内容接入约束

`sys_i18n_content`用于承载“绑定具体业务记录”的多语言标题、摘要、描述或正文内容。

业务模块接入时请遵循以下锚点约束：

| 字段            | 约束                                         | 示例                       |
| --------------- | -------------------------------------------- | -------------------------- |
| `business_type` | 使用稳定的模块级标识，不使用会变化的展示名称 | `notice`、`cms_article`    |
| `business_id`   | 使用稳定主键或不可变业务编码                 | `42`、`article-homepage`   |
| `field`         | 使用业务聚合中的稳定字段名                   | `title`、`summary`、`body` |
| `locale`        | 使用规范化运行时语言编码                     | `zh-CN`、`en-US`           |
| `content_type`  | 仅使用 `plain`、`markdown`、`html`、`json`   | `markdown`                 |

推荐读取策略：

1. 业务主表保留源语言字段，作为最终兜底值。
2. 业务服务按 `business_type + business_id + field + locale` 查询 `sys_i18n_content`。
3. 若目标语言缺失，则回退到运行时默认语言。
4. 若默认语言也缺失，则回退到业务主表中的原始字段值。

缓存规范：

- 缓存粒度使用完整锚点 `business_type + business_id + field`。
- 缓存内容按“同一锚点下的全部语言变体”存储，而不是每个 locale 单独缓存，便于一次失效后整体刷新。
- 业务模块在新增、修改、删除、导入、发布多语言内容后，必须立即失效对应锚点缓存。
- 禁止在没有显式失效策略的前提下永久缓存“未命中”结果，否则后续写入无法及时生效。

使用边界：

- 可复用的界面文案和元数据标签使用`sys_i18n_message`。
- 只有“绑定具体业务记录”的内容才接入`sys_i18n_content`。
- 附件或富媒体引用继续放在业务主表或关联表中，`sys_i18n_content`只存放多语言文本载荷。

## 编写示例

```json
{
  "framework.description": "AI驱动的全栈开发框架",
  "menu.system.title": "系统管理",
  "menu.system.users.title": "用户管理",
  "config.sys.account.captchaEnabled.name": "登录验证码",
  "publicFrontend.login.title": "欢迎回来",
  "plugin.org-center.name": "组织中心",
  "plugin.org-center.description": "部门、岗位与层级治理"
}
```
