## ADDED Requirements

### Requirement: 插件清单与生命周期必须支持新增语言时的零代码扩展
插件清单与生命周期 SHALL 在宿主新增内置语言时自动覆盖新语言的运行时 UI 翻译资源与 apidoc 翻译资源,无需修改宿主代码或单个插件源码。源码插件 SHALL 在自身 `manifest/i18n/<locale>.json` 与 `manifest/i18n/apidoc/<locale>/**/*.json` 内追加该语言资源,动态插件 SHALL 在打包阶段把该语言资源写入 release 自定义节;宿主在加载、启停、升级、卸载链路中按现有规则自动发现、装载、清理这些新增语言的资源。

#### Scenario: 启用阿拉伯语后插件资源自动接入
- **WHEN** 宿主启用 `ar-SA` 内置语言
- **AND** 源码插件 `apps/lina-plugins/<plugin-id>/manifest/i18n/ar-SA.json` 存在
- **THEN** 该插件被启用时,其 `ar-SA` 翻译资源自动加入运行时翻译聚合结果
- **AND** 该插件被停用或卸载时,`ar-SA` 翻译资源同步从聚合结果中移除
- **AND** 整条链路不要求宿主代码修改、不要求其他插件代码修改

#### Scenario: 动态插件按 release 携带阿拉伯语资源
- **WHEN** 一个动态插件在新版本中追加 `manifest/i18n/ar-SA.json` 后重新打包发布
- **AND** 宿主升级该插件到新 release
- **THEN** 升级后阿拉伯语翻译资源生效,旧版本资源不再使用
- **AND** 升级期间宿主仅清理与该插件相关的扇区缓存,不影响其他语言或其他插件
