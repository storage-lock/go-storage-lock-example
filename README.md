# Storage Lock Examples

本仓库是 [storage-lock](https://github.com/storage-lock) 生态的示例与选型指南。
示例展示 API 用法，不构成对应后端或线上拓扑的生产安全认证。每个存储的用法都是同一套模式：

> 根模块直接展示可通过 Go Proxy 下载的示例；etcd、Redis、LevelDB 位于各自独立子模块，依赖组织内私有 backend，需要在有仓库权限的完整工作区运行。MySQL 示例使用公开 backend 模块。

```
构造 Storage（DSN/URI/路径）→ NewLock 创建锁 → Lock/UnLock（或 RLock/WLock）→ 看门狗自动续租
```

## 快速开始

| 示例 | 存储 | 运行 |
|------|------|------|
| [quick_start](examples/quick_start/main.go) | 内存（演示用） | `go run ./examples/quick_start` |
| [redis](examples/redis/main.go) | Redis（需要组织内 backend 模块权限） | 工作区内 `STORAGE_LOCK_REDIS_URI=redis://localhost:6379/0 go run ./examples/redis` |
| [mysql](examples/mysql/main.go) | MySQL（需要组织内 backend 模块权限） | 工作区内 `STORAGE_LOCK_MYSQL_DSN=root:pass@tcp(host:3306)/db go run ./examples/mysql` |
| [etcd](examples/etcd/main.go) | etcd（需要组织内 backend 模块权限） | 工作区内 `STORAGE_LOCK_ETCD_ENDPOINTS=127.0.0.1:2379 go run ./examples/etcd` |
| [leveldb](examples/leveldb/main.go) | LevelDB（需要组织内 backend 模块权限） | 工作区内 `go run ./examples/leveldb` |
| [filesystem](examples/filesystem/main.go) | 文件系统（flock） | `go run ./examples/filesystem` |

## 跨语言互操作（HTTP 锁服务网关）

[examples/http_gateway](examples/http_gateway)（独立子模块）演示"锁服务网关"模式：
**Go 进程持有存储介质，通过无状态 HTTP 接口把锁语义暴露给任意语言**。
非 Go 语言不需要重新实现 CAS / 租约 / fencing token。

```bash
# 1. 起 Go 服务（默认 memory 存储）
cd examples/http_gateway && go run .

# 2. 另开终端：Python 客户端
python3 client.py

# 3. 或 Java 客户端
javac -d /tmp -cp <java-storage-lock-src> Client.java
java -cp /tmp Client
```

三个进程共享同一把锁的语义：acquire 互斥（第二个抢锁收到 409）、renew 续租、
inspect 查看、release 释放、fencing token 递增。已由
[storage-lock-http-api 的跨语言集成测试](../storage-lock-http-api/admin/cross_language_test.go) 验证。

## 介质 × 能力选择矩阵

选介质先看两个问题：**部署形态**（要不要独立服务）和**一致性需求**（能不能容忍
单点/时钟漂移）。下表按"从轻到重"排列：

| 存储介质 | 部署形态 | 可考虑的使用范围 | 必须单独验证 |
|----------|----------|------------------|--------------|
| 内存 | 单进程 | 演示、单进程内多协程 | 不具备跨进程协调能力 |
| 文件系统 (flock) | 单机共享文件系统 | 同主机多进程（依赖文件系统锁语义） | 容器挂载、网络文件系统、故障恢复与共享路径 |
| LevelDB | 嵌入式、本机文件 | 本机多进程（依赖实现的文件锁与数据持久性） | 不能假定多主机共享目录安全 |
| Redis | 独立服务 | 仅当该部署为锁记录提供线性一致 CAS 和可靠时间时 | 主从切换、异步复制、故障恢复；不以 RedLock/quorum 替代前提 |
| SQL/NoSQL | 独立服务 | 仅限具体适配器与事务/CAS 语义经验证的部署 | 隔离级别、条件更新、主从切换、服务端时钟；不能按数据库家族泛化承诺 |
| etcd / ZooKeeper | 独立服务 | 仅限具体适配器正确使用其一致性原语的部署 | 集群多数派、租约/会话、网络分区及版本兼容 |
| 对象存储 (S3/OSS/COS/BOS) | 独立服务 | 仅当条件写语义满足线性一致 CAS，且另有可靠协调时间源 | 供应商/区域/桶配置；TimeProvider 不会修复非原子的条件写 |
| 消息队列 | 独立服务 | 不作为本库的通用锁存储推荐 | 分区/消费语义并不自动提供锁所需的 CAS 和协调时间 |

### 能力矩阵（存储能力 × 锁特性）

| 存储能力 | 需要的锁特性 | 不支持的后果 |
|----------|--------------|--------------|
| CAS（条件更新） | 互斥、可重入、公平排队 | 无法保证互斥（无原子比较） |
| 可靠时间源（GetTime） | 租约过期、看门狗续租 | 租约不可靠 → 用固定 TTL + 手动续租 |
| 原子条件删除 | 释放锁（无墓碑） | 降级为墓碑（LockCount=0），语义不变 |
| Watch/订阅 | 跨实例通知（Subscribe） | 降级为轮询（PollingLockWatcher） |

> 判定某个具体存储配置是否可承载跨节点分布式锁，必须确认同一 lockId 的条件写是**线性一致 CAS**，且时间源可靠、单调并在节点间协调。能力标志只是实现的自我声明；仅有 CRUD 或最终一致副本不满足前提。对象存储即使注入外部时间源，也仍须单独证明其条件写满足线性一致 CAS。
> 本仓库示例不是生产认证。请对实际客户端、服务端版本、复制/故障切换拓扑运行契约及锁语义测试；未执行并通过的组合标为“未验证”，不要据此承诺生产互斥。
> 本项目不建议通过多个独立后端的 quorum/RedLock 来推导互斥性。详见 [go-storage 的 Capability](../go-storage/storage.go) 和 [核心能力边界](../go-storage-lock/README.md#能力边界)。

### Fencing token 使用边界

`WithLock` 会把 fencing token 传给回调。锁租约失效后旧进程可能恢复运行，因此受保护资源必须在**同一事务/原子操作**中校验 token 并更新“已接受最大 token”；小于等于已接受值的请求必须拒绝。示例 `examples/quick_start` 展示本地互斥实现和拒绝旧 token 的测试。生产资源应在数据库条件更新或等价原子事务中实现，不能用进程内 mutex 替代跨实例校验；相同 token 的重试还需业务幂等规则。

## 下一步

- 完整语义测试套件：[go-storage-lock-test](../go-storage-lock-test)（本地 6 介质 + 外部服务按环境变量跳过）
- 混沌测试（分区/抖动/时钟漂移/并发交接）：[chaos_test.go](../go-storage-lock-test/chaos_test.go)
- HTTP 锁服务（生产化网关）：[storage-lock-http-api](../storage-lock-http-api)
- 可观测性：Prometheus 指标 + Web 看板 [go-storage-lock-prometheus](../go-storage-lock-prometheus)、审计日志 [go-storage-lock-auditlog](../go-storage-lock-auditlog)
