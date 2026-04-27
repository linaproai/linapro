# locales

该目录用于存放仅供 `web-antd` 应用使用的国际化扩展，例如 dayjs 多语言设置、Ant Design Vue 多语言接线以及应用级翻译资源。

应用级语言包会从 `langs/<locale>/*.json` 自动发现。语言切换器由 `GET /api/v1/i18n/runtime/locales` 返回的运行时元数据驱动，因此新增内置语言时应只新增语言 JSON 资源，并按需在宿主默认配置 `i18n.locales` 中补充元数据，不需要修改前端 TypeScript 语言清单。

## 运行时元数据

`GET /api/v1/i18n/runtime/locales`是语言切换器、默认语言、原生名和语言切换开关的唯一来源。运行时文本方向固定为`ltr`；语言切换时`<html dir>`和 Ant Design Vue 的`ConfigProvider.direction`都保持`ltr`。

第三方库语言接线应优先通过语言编码约定和已生成的语言 loader key 推导。新增内置语言时，不应再维护前端语言注册表或逐语言兜底 map。

## 运行时语言包缓存

运行时 UI 文案从`GET /api/v1/i18n/runtime/messages?lang=<locale>`加载，并按`linapro:i18n:runtime:<locale>`键在`localStorage`中持久化 7 天。

缓存内容为`{etag, messages, savedAt}`。命中新鲜持久化缓存时先立即渲染，再在后台带`If-None-Match`刷新；若服务端返回`304 Not Modified`，则保持当前语言包不变。
