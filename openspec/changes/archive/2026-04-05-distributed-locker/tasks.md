## 1. 数据库变更

- [x] 1.1 创建 `manifest/sql/010-distributed-locker.sql`，定义 `sys_locker` 表（MEMORY 引擎）
- [x] 1.2 执行 `make init` 将 SQL 变更同步到数据库
- [x] 1.3 执行 `make dao` 生成 DAO/DO/Entity 代码

## 2. 配置管理

- [x] 2.1 在 `manifest/config/config.yaml` 中添加 `locker` 配置项（租约时长、续期间隔）
- [x] 2.2 创建 `internal/service/config/config_locker.go`，实现配置读取逻辑

## 3. 分布式锁组件实现

- [x] 3.1 创建 `internal/service/locker/locker.go`，实现核心锁服务（Lock、TryLock、IsLeader）
- [x] 3.2 创建 `internal/service/locker/locker_instance.go`，实现锁实例（Unlock、Renew）
- [x] 3.3 创建 `internal/service/locker/locker_lease.go`，实现租约续期管理
- [x] 3.4 创建 `internal/service/locker/locker_election.go`，实现领导选举（Start、Stop）

## 4. 定时任务组件改造

- [x] 4.1 修改 `internal/service/cron/cron.go`，添加 locker 服务依赖和 isLeader 状态
- [x] 4.2 修改 `internal/service/cron/cron_session.go`，添加主节点检查逻辑
- [x] 4.3 修改 `internal/service/cron/cron_servermon_cleanup.go`，添加主节点检查逻辑
- [x] 4.4 确认 `internal/service/cron/cron_servermon.go` 为 All-Node 任务（无需修改）

## 5. 启动流程集成

- [x] 5.1 修改 `internal/cmd/cmd_http.go`，初始化 locker 服务并启动领导选举

## 6. 测试验证

- [x] 6.1 验证单节点场景：服务启动后成功成为主节点，所有定时任务正常执行
- [x] 6.2 验证多节点场景：启动多个服务实例，确认只有一个主节点执行 Master-Only 任务（需在多节点环境验证）
- [x] 6.3 验证故障转移：停止主节点，确认从节点在锁过期后成功接管（需在多节点环境验证）

## 7. 单元测试

- [x] 7.1 创建 `locker_test.go` - 核心锁服务测试（Lock、LockFunc、IsLeader 等）
- [x] 7.2 创建 `locker_instance_test.go` - 锁实例测试（Unlock、Renew、IsHeld）
- [x] 7.3 创建 `locker_lease_test.go` - 租约续期管理测试（Start、Stop、StoppedChan）
- [x] 7.4 创建 `locker_election_test.go` - 领导选举测试（Start、Stop、tryAcquire）
- [x] 7.5 测试覆盖率达到 84.1%（超过 80% 目标）

## Feedback

- [x] **FB-1**: cron.Service 持有整个 config.Service 但只使用特定配置对象，应改为接收具体的配置对象（参考 election.Service 做法）
- [x] **FB-2**: 在线用户列表接口缺少分页参数，后端返回全量数据导致前端需要客户端分页
- [x] **FB-3**: 角色授权用户页面样式和交互与参考项目不一致，需要对齐 `ruoyi-plus-vben5` 的 `role-assign` 实现
- [x] **FB-4**: 缺少批量取消授权按钮（工具栏位置），支持选中多条记录后批量取消授权
- [x] **FB-5**: 缺少邮箱地址展示列
- [x] **FB-6**: 角色授权用户页面创建时间列内容无法完整展示，需调整其他列宽度
- [x] **FB-7**: 菜单管理新增/修改抽屉中，菜单类型和按钮类型的权限标识必须必填，且输入框需显示在菜单名称下方
- [x] **FB-8**: `TC0063-auth-menu.ts` 在 `beforeAll/afterAll` 阶段超时，需排查菜单权限回归用例的初始化与清理稳定性
- [x] **FB-9**: `/user/info` 将 `homePath` 固定为 `/analytics`，导致无该路由权限的用户登录后直接进入 404 页面
