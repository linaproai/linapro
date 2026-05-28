# LinaPro 后端文档目录

本目录承载跨包通用的技术指南。宿主内部架构说明放在对应代码所在包的
`README` 或文档注释里；本目录只放跨多个包、或描述对接契约的指南。

> English source: [README.md](./README.md).

## 文档清单

| 文档 | 主题 |
|---|---|
| [auth-provider-integration.md](./auth-provider-integration.md) / [auth-provider-integration.zh-CN.md](./auth-provider-integration.zh-CN.md) | 如何把一个新的第三方认证 Provider（Google、Discord、GitHub、自建 OIDC 等）作为源码插件接入 LinaPro。|

## 规范

- 每篇指南都提供英文源文件（`<name>.md`）和中文镜像（`<name>.zh-CN.md`），二者必须内容等价。
- 代码片段仅作示例。可复制的最终模板存放在对应的仓库代码里，通常位于 `apps/lina-core/pkg/...` 或 `apps/lina-plugins/<plugin-id>/...` 之下。
- 当指南引用公开 API 表面时，必须给出对应的文件路径，便于读者回到代码源核对行为。
