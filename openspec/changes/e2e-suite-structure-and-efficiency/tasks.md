## 1. 基线盘点与治理骨架

- [x] 1.1 盘点当前 `hack/tests/e2e/` 文件树,输出旧目录到目标能力目录的迁移映射,同时记录重复 TC ID、非 `TC*.ts` 文件和高频 `waitForTimeout` 热点
- [x] 1.2 在 `hack/tests/` 下建立新的执行清单与支持目录骨架(如 `config/`、`support/`、`scripts/`、`debug/`),明确 `smoke`、模块范围和串行池声明位置
- [x] 1.3 新增 E2E 治理校验脚本,自动检查 TC 全局唯一性、非法文件混入、错误目录归属以及执行清单引用失效

## 2. 目录重组与编号治理

- [x] 2.1 创建目标能力目录树,将 `hack/tests/e2e/system/` 按 `iam / settings / org / content / scheduler` 等稳定能力边界拆分迁移并修复导入路径
- [x] 2.2 将 `monitor/`、`plugin/` 等目录按插件能力和子域继续细分到目标结构,保证监控与扩展中心相关用例可按能力定位
- [x] 2.3 将 `hack/tests/e2e/system/job/helpers.ts`、`hack/tests/e2e/debug/export-debug.ts` 等非用例文件迁出 `e2e/`,落到专用支持目录
- [x] 2.4 修复现有重复 TC ID 与命名冲突,同步更新受影响的导入、文档和执行清单引用

## 3. 运行入口与认证态复用

- [x] 3.1 为 Playwright 增加管理员登录态预生成步骤和 `storageState` 产物管理,并在 `playwright.config.ts` 中接入该准备流程
- [x] 3.2 重构 `hack/tests/fixtures/auth.ts`,让 `adminPage` 等高频夹具默认复用预生成登录态,同时保留认证主题用例的真实登录路径
- [x] 3.3 在 `hack/tests/package.json` 中新增 `test:smoke`、`test:module`、`test:full` 等入口,保留 `pnpm test` 的完整回归语义
- [x] 3.4 基于执行清单实现并行池与串行池调度,让完整回归可以按“可并行文件 + 高风险串行文件”两阶段运行

## 4. 等待策略与提速改造

- [x] 4.1 在 `hack/tests/support/` 或等价公共层新增表格、弹窗、提示反馈、路由稳定等状态型等待工具
- [x] 4.2 优先重构固定等待最密集的页面对象,至少覆盖菜单、角色、字典、用户、配置等高频页面,将主要 `waitForTimeout` 替换为状态型等待
- [x] 4.3 审查插件治理、权限治理、导入导出、运行参数等共享状态明显的测试文件,明确哪些进入串行池,哪些在修复隔离后可以进入并行池
- [ ] 4.4 对迁移后仍存在的高频固定等待进行第二轮清理,仅保留有明确业务理由的少量兜底等待并补充注释说明

## 5. 文档与验证

- [x] 5.1 为 `hack/tests/` 新增 `README.md` 与 `README.zh_CN.md`,说明目录边界、执行入口、清单机制和治理脚本使用方式
- [ ] 5.2 运行治理校验脚本、`pnpm test:smoke`、至少两个模块定向回归以及 `pnpm test`/`pnpm test:full`,记录迁移前后耗时和稳定性基线
  - 已完成 `pnpm run test:validate`、`pnpm test:smoke`、`pnpm run test:module -- iam:user`、`pnpm run test:module -- settings:config` 与多轮 `pnpm run test:full` 验证; 当前完整回归阻塞于 `TC0066-source-plugin-lifecycle` 与 `TC0067-runtime-wasm-lifecycle` 的既有插件用例失败,需在后续迭代单独跟进
- [x] 5.3 对照本次 OpenSpec 设计与规范检查目录结构、执行分层、登录态复用和并行边界是否全部落地,为后续 `/opsx:apply` 与实现评审做好准备
  - 已完成目录边界、执行清单、`storageState` 复用、认证主题真实登录路径与并行/串行分层的实现自查; 当前遗留项仅剩完整回归中的插件既有失败与部分固定等待二轮清理
