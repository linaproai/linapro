## 1. 基线审计与冲突分类

- [x] 1.1 复盘最近一次完整 E2E 回归日志，整理共享状态冲突案例、受影响测试文件、触发原因和修复方向
- [x] 1.2 扫描 `hack/tests/e2e` 中插件生命周期、系统参数、字典、菜单权限、运行时 i18n 缓存、公共配置缓存等高风险操作命中项
- [x] 1.3 为每个命中项判定隔离类别、是否必须串行、是否可通过 fixture 或语义断言解除冲突
- [x] 1.4 在变更文档中记录 i18n 影响评估：本变更不新增产品运行时文案，仅治理测试和文档；涉及 i18n 缓存测试时保持语义验证

## 2. 执行清单与运行器治理

- [x] 2.1 扩展 `hack/tests/config/execution-manifest.json`，为串行条目或高风险文件增加机器可读 isolation category
- [x] 2.2 更新 `hack/tests/scripts/run-suite.mjs`，在 full、smoke、module 模式输出并行文件数、串行文件数、worker 数和串行隔离类别摘要
- [x] 2.3 保持 module 模式继续按同一串行/并行边界拆分，避免模块回归绕过隔离规则
- [x] 2.4 为运行器新增脚本级单元测试或可执行验证用例，覆盖 full/module 拆分和 category 摘要输出

## 3. E2E 校验门禁

- [x] 3.1 扩展 `hack/tests/scripts/validate-e2e.mjs`，校验 isolation category 格式、引用路径存在性和串行条目分类完整性
- [x] 3.2 增加高风险启发式检测，覆盖插件 install/enable/disable/uninstall/sync/upload/upgrade、系统参数写入、字典导入/修改、菜单权限修改和 runtime i18n ETag 缓存断言
- [x] 3.3 对命中高风险模式但不在串行边界或未声明分类的文件输出可操作错误信息
- [x] 3.4 如存在合理并行安全例外，新增带原因的 allowlist 结构并在校验中强制填写说明

## 4. Fixture 与测试用例整改

- [x] 4.1 收敛源码插件状态准备逻辑，确保依赖插件页面/API/mock 数据的测试统一走幂等 fixture
- [x] 4.2 梳理动态插件和源码插件测试中的本地文件、附件、插件表和 mock 数据清理逻辑，确保单文件可独立运行
- [x] 4.3 调整缓存/ETag 类测试，断言请求携带条件头，并接受 `304` 或合法的 `200 + 新 ETag + body`
- [x] 4.4 调整从本地化 UI 文本反推业务状态的测试，业务断言改用稳定 ID、code、labelKey、permission key 或 API counter
- [x] 4.5 对已知冲突相关文件进行最小范围整改，至少覆盖 `TC0124-runtime-i18n-etag.ts`、插件生命周期相关用例、组织部门树计数用例和内容通知依赖用例

## 5. 文档与验收材料

- [x] 5.1 更新 `hack/tests/README.md` 和 `hack/tests/README.zh_CN.md`，说明隔离类别、串行边界、fixture 前置条件和缓存语义断言规则
- [x] 5.2 在本变更目录新增 E2E 冲突治理记录，列出冲突类型、代表用例、修复方式和后续新增用例检查清单
- [x] 5.3 确认新增或修改的测试治理文档语言与当前变更保持中文；仓库级 README 新增内容保持中英文镜像一致

## 6. 验证与审查

- [x] 6.1 运行 `cd hack/tests && pnpm test:validate`，确认 E2E 命名、目录、隔离类别和高风险检测通过
- [x] 6.2 运行脚本级验证或相关 Node 测试，确认 `run-suite.mjs` 和 `validate-e2e.mjs` 新逻辑稳定
- [x] 6.3 运行受影响 E2E 子集：`TC0124-runtime-i18n-etag.ts`、插件生命周期相关用例、组织部门树计数用例、内容通知依赖用例
- [x] 6.4 运行 `cd hack/tests && pnpm test` 完整回归，记录并行阶段、串行阶段和跳过项结果
- [x] 6.5 运行 `openspec validate e2e-global-state-isolation`
- [x] 6.6 执行 `lina-review`，重点审查 E2E 隔离边界、i18n 影响结论、fixture 独立性、文档中英文一致性和测试覆盖

### 验证记录

- `pnpm test:validate`：通过，校验 138 个 E2E 文件、28 个 scope、7 个 smoke 文件、103 个串行文件。
- `node scripts/run-suite.mjs module i18n --list`：通过，输出 module 拆分与隔离类别摘要。
- 受影响子集：`TC0066-source-plugin-lifecycle.ts`、`TC0124-runtime-i18n-etag.ts`、`TC0021-user-dept-tree-count.ts`、`TC0037-notice-crud.ts` 覆盖通过；`TC0124` 修复后单文件 5/5 通过。
- 单文件复验：`TC0044c`、`TC0097`、`TC0058f~h`、`TC0012` 通过。
- 完整回归：第二轮 `pnpm test` 并行阶段 102/102 通过；串行阶段执行 326 个断言，260 通过、6 跳过、4 未运行。真实断言失败为 `TC0044c` 和 `TC0097`，均已修复并单文件复验通过；其余 54 个失败来自前后端 dev server 中途退出后的 `ERR_CONNECTION_REFUSED` 级联，重启服务后抽样复验设置/字典尾段通过。
