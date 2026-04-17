## ADDED Requirements

### Requirement: 目录级 README 采用双语镜像
系统 SHALL 要求仓库内所有目录级主说明文档同时提供英文主文档 `README.md` 与中文镜像文档 `README.zh_CN.md`。

#### Scenario: 目录已存在主 README
- **WHEN** 某个目录存在主说明文档
- **THEN** 该目录同时存在英文 `README.md` 与中文 `README.zh_CN.md`
- **AND** `README.md` 使用英文编写
- **AND** `README.zh_CN.md` 使用中文编写

#### Scenario: 新增目录说明文档
- **WHEN** 开发者为某个目录新增主说明文档
- **THEN** 在同一次变更中同步创建 `README.md` 与 `README.zh_CN.md`
- **AND** 不允许只新增单语版本的目录级 README

### Requirement: 双语 README 内容保持同步
系统 SHALL 要求英文主文档与中文镜像文档在结构和信息上保持一致，仅语言不同。

#### Scenario: 更新英文 README
- **WHEN** 开发者更新某目录的 `README.md`
- **THEN** 同一次变更同步更新对应的 `README.zh_CN.md`
- **AND** 两份文档保持相同的章节结构和技术事实

#### Scenario: 校验双语 README 一致性
- **WHEN** reviewer 检查某目录下的 README 对
- **THEN** 能够确认两份文档表达的是同一组技术信息
- **AND** 不出现英文主文档已更新但中文镜像仍停留在旧内容的情况

### Requirement: 仓库根目录必须提供双语入口 README
系统 SHALL 在仓库根目录提供英文主入口 README 与中文镜像 README，作为项目的统一入口说明。

#### Scenario: 查看仓库根目录说明
- **WHEN** 用户访问仓库根目录
- **THEN** 可以看到英文 `README.md` 与中文 `README.zh_CN.md`
- **AND** 两份文档都能完整介绍项目定位、目录结构、开发方式和关键规范入口
