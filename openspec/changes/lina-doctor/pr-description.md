# PR 描述草案

## 变更摘要

- 精简`hack/scripts/install/install.sh`：托管安装入口现在只下载`LinaPro`仓库源码，并打印后续步骤。
- 安装入口默认通过 Git 解析最高稳定发布标签并保留`origin`远程地址，后续可通过`git fetch --tags --force origin`拉取新标签升级。
- 新增中文`lina-doctor`技能：通过 AI 工具诊断并按计划安装 Go、Node、pnpm、OpenSpec、GoFrame CLI、Playwright browsers 与`goframe-v2`技能。
- 外部化`goframe-v2`技能：仓库不再附带该技能目录，改由`npx skills add github.com/gogf/skills -g`安装到用户全局技能目录。

## Breaking Changes

- 安装路径不再执行环境检查、依赖安装、`make init`或`make mock`。
- `LINAPRO_NON_INTERACTIVE`和`LINAPRO_SKIP_MOCK`不再作为安装器输入。
- `hack/scripts/install/`下的平台脚本、`prereq.sh`和安装公共库已删除。
- 仓库内`.claude/skills/goframe-v2/`（实际通过`.agents/skills/goframe-v2/`追踪）已删除；后端开发前需通过`lina-doctor`或等价命令安装全局技能。

## 运维动作

- 将更新后的`hack/scripts/install/install.sh`重新部署到`https://linapro.ai/install.sh`。
- 部署后刷新 CDN 缓存，并在干净环境验证`curl -fsSL https://linapro.ai/install.sh | bash`只完成 clone 与 next steps 输出。

## 验证结果

- `openspec validate lina-doctor --strict`
- `bash -n hack/scripts/install/install.sh`
- `bash -n .agents/skills/lina-doctor/scripts/doctor-*.sh`
- `bash hack/tests/scripts/install-bootstrap.sh all`
- `bash hack/tests/scripts/doctor-check.sh`
- `bash hack/tests/scripts/doctor-plan.sh`
- `bash hack/tests/scripts/doctor-escalate.sh`
- `cd apps/lina-core && go test ./...`
- `cd apps/lina-core && make build`
- `cd apps/lina-vben && pnpm run build`
- `make test`完整 E2E 回归：parallel 108 passed；serial 356 passed、6 skipped；0 failed。

## i18n 评估

本变更不修改前端运行时页面、菜单、按钮、表单、表格、接口 DTO、插件 manifest 或后端运行时 i18n 资源。新增用户可见文本集中在 CLI 输出、安装 README 与中文 skill 文档中，不进入应用运行时语言包；因此无需新增或修改运行时 i18n JSON、manifest/i18n 或 apidoc i18n 资源。

## 缓存一致性评估

本变更不新增、不修改、不失效任何应用运行时缓存；`TC0098`仅在测试中等待插件运行时状态收敛后刷新投影，不改变缓存实现。分布式缓存一致性、失效范围、集群拓扑与权威数据源均无实现影响。
