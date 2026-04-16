## Why

当前 `lina` 服务在多节点部署场景下，所有节点上的定时任务（如 Session 清理、监控数据清理等）会同时执行，导致重复操作和数据冗余。需要引入分布式锁机制实现领导选举，确保 Master-Only 类型的定时任务在同一时间只有一个节点执行。

## What Changes

- 新增 `locker` 分布式锁组件，基于数据库实现锁的获取、释放和租约续期功能
- 新增 `sys_locker` 数据表（MEMORY 引擎），用于存储分布式锁状态
- 改造 `cron` 组件，区分 Master-Only Jobs 和 All-Node Jobs 两类定时任务
- 实现领导选举机制，主节点通过分布式锁抢占并定期续租约
- Master-Only Jobs 在执行前检查当前节点是否为主节点，非主节点跳过执行

## Capabilities

### New Capabilities

- `distributed-locker`: 分布式锁组件，提供锁获取、释放、租约续期等核心能力
- `leader-election`: 领导选举机制，支持主节点抢占、租约续期、故障转移

### Modified Capabilities

- `cron-jobs`: 定时任务管理，新增任务分类（Master-Only / All-Node）和主节点检查逻辑

## Impact

**后端服务层**:
- 新增 `internal/service/locker/` 组件目录
- 改造 `internal/service/cron/` 组件，集成分布式锁
- 改造 `internal/cmd/cmd_http.go`，初始化领导选举

**数据库**:
- 新增 `sys_locker` 表（MEMORY 引擎）

**配置文件**:
- `manifest/config/config.yaml` 新增 `locker` 配置项
