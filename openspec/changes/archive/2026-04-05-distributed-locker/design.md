## Context

当前 `lina` 服务采用前后端分离架构，后端使用 GoFrame 框架，定时任务通过 `gcron` 组件管理。在单节点部署场景下运行正常，但在多节点部署（如 Kubernetes、多实例负载均衡）场景下，所有节点上的定时任务会同时执行，导致：

- **Session Cleanup**: 多节点同时清理，产生重复操作
- **Server Monitor Collector**: 每个节点采集自己的数据（这是预期行为）
- **Server Monitor Cleanup**: 多节点同时清理，可能产生竞态条件

需要引入分布式锁机制，区分 Master-Only Jobs 和 All-Node Jobs。

## Goals / Non-Goals

**Goals:**
- 实现基于数据库的分布式锁组件，支持锁获取、释放、租约续期
- 实现领导选举机制，确保多节点场景下只有一个 Master 节点
- Master-Only Jobs 只在主节点执行，故障时自动转移
- All-Node Jobs 在每个节点独立执行
- 服务重启时自动参与领导选举，无需手动干预

**Non-Goals:**
- 不使用 Redis 等外部依赖，仅基于数据库实现
- 不支持手动释放领导权
- 不实现分布式事务协调

## Decisions

### 1. 分布式锁存储方案

**选择: MySQL MEMORY 引擎**

| 方案 | 优点 | 缺点 |
|------|------|------|
| MySQL InnoDB | 数据持久化、支持事务 | 性能较低、需要定期清理过期锁 |
| MySQL MEMORY | 极快读写、自动清理（重启时） | 数据不持久、不支持事务 |
| Redis | 高性能、原生支持 TTL | 引入外部依赖 |

**理由:**
- 分布式锁是临时状态，服务重启后丢失是合理的
- MEMORY 引擎的低延迟特性适合频繁的锁操作
- 不引入 Redis 等外部依赖，降低部署复杂度

### 2. 领导选举策略

**选择: 乐观锁 + 租约续期**

```
选举流程:
┌─────────────────────────────────────────────────────┐
│  1. 尝试获取锁 (INSERT 或 UPDATE expire_time)       │
│  2. 成功 → 成为主节点，启动租约续期协程              │
│  3. 失败 → 成为从节点，等待下次选举机会              │
│  4. 主节点定期续租约 (UPDATE expire_time)           │
│  5. 续期失败 → 降级为从节点                         │
└─────────────────────────────────────────────────────┘
```

**理由:**
- 乐观锁实现简单，无需复杂的选举协议
- 租约机制确保故障节点自动释放锁
- 续期失败时立即降级，避免短暂的双主问题

### 3. 定时任务分类策略

| 任务类型 | 执行策略 | 原因 |
|----------|----------|------|
| Session Cleanup | Master-Only | 清理操作只需一个节点执行 |
| Server Monitor Collector | All-Node | 每个节点采集自己的系统资源数据 |
| Server Monitor Cleanup | Master-Only | 清理操作只需一个节点执行 |

**理由:**
- Server Monitor Collector 需要采集每个节点的 CPU、内存等数据，必须 All-Node 执行
- Cleanup 类操作只需要一个节点执行，避免重复

### 4. 组件架构设计

```
service/locker/
├── locker.go           # 核心锁服务
│   ├── Lock()          # 获取锁
│   ├── TryLock()       # 尝试获取锁（非阻塞）
│   └── IsLeader()      # 检查当前节点是否为主节点
├── locker_instance.go  # 锁实例
│   ├── Unlock()        # 释放锁
│   └── Renew()         # 续租约
├── locker_lease.go     # 租约续期管理
│   └── StartRenewal()  # 启动后台续期协程
└── locker_election.go  # 领导选举
    ├── Start()         # 启动选举
    └── Stop()          # 停止选举
```

## Risks / Trade-offs

### Risk 1: 数据库故障导致无法选举
**Mitigation**: 数据库故障时，所有节点降级为从节点，不执行 Master-Only Jobs。这是安全的降级行为。

### Risk 2: 网络分区导致短暂双主
**Mitigation**: 租约时长设计为 30 秒，续期间隔为 10 秒。即使发生网络分区，最多 30 秒后旧主节点的锁会过期，新主节点可以获取锁。

### Risk 3: MEMORY 表大小限制
**Mitigation**: sys_locker 表只存储一条锁记录（领导选举锁），不会超过 MEMORY 表的限制（默认 16MB）。

### Trade-off: 不支持手动释放领导权
**Reason**: 简化实现复杂度。服务优雅关闭时，租约会自然过期，其他节点自动接管。
