# Storage Lock Examples

本仓库是 [storage-lock](https://github.com/storage-lock) 生态的示例与选型指南。
每个存储的用法都是同一套模式：

> 公开模块默认只展示可通过 Go Proxy 下载的示例；etcd、Redis、LevelDB、MySQL 示例依赖组织内私有 backend 模块，需要在有仓库权限的完整工作区中运行。

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

[examples/http_gateway](examples/http_gateway) 演示"锁服务网关"模式：
**Go 进程持有存储介质，通过无状态 HTTP 接口把锁语义暴露给任意语言**。
非 Go 语言不需要重新实现 CAS / 租约 / fencing token。

```bash
# 1. 起 Go 服务（默认 memory 存储）
go run ./examples/http_gateway

# 2. 另开终端：Python 客户端
python3 ./examples/http_gateway/client.py

# 3. 或 Java 客户端
javac -d /tmp -cp <java-storage-lock-src> ./examples/http_gateway/Client.java
java -cp /tmp Client
```

三个进程共享同一把锁的语义：acquire 互斥（第二个抢锁收到 409）、renew 续租、
inspect 查看、release 释放、fencing token 递增。已由
[storage-lock-http-api 的跨语言集成测试](../storage-lock-http-api/admin/cross_language_test.go) 验证。

## 介质 × 能力选择矩阵

选介质先看两个问题：**部署形态**（要不要独立服务）和**一致性需求**（能不能容忍
单点/时钟漂移）。下表按"从轻到重"排列：

| 存储介质 | 部署形态 | 一致性 | 时钟依赖 | 适合场景 |
|----------|----------|--------|----------|----------|
| 内存 | 无（单进程） | 单机最强 | 无 | 演示、单进程内多协程 |
| 文件系统 (flock) | 无（单机） | 单机互斥强 | 无 | 同一台机器的多个 worker/进程 |
| LevelDB | 无（嵌入式） | 单机互斥强 | 无 | 单机多进程 + 持久化锁记录 |
| Redis | 独立服务 | 主从可能丢锁（需 RedLock/哨兵） | 无 | 跨机器、吞吐优先 |
| MySQL/PG/SQL 家族 | 独立服务 | 强（事务 + 行锁） | 无 | 企业已有数据库、事务一致性 |
| etcd | 独立服务 | 线性一致（最强） | 无 | K8s 生态、强一致 + 服务发现 |
| 对象存储 (S3/OSS/COS/BOS) | 独立服务 | 条件 PUT 弱一致 | 有 | 已有对象存储、低频互斥 |
| ZooKeeper | 独立服务 | 顺序一致 | 无 | 已有 ZK、强一致 + 服务发现 |
| MongoDB / Cassandra | 独立服务 | 各自一致性模型 | 无 | 已有该 NoSQL、跨机器互斥 |
| 消息队列 (Kafka/RocketMQ) | 独立服务 | 弱 | 有 | 已有 MQ、低竞争场景 |

### 能力矩阵（存储能力 × 锁特性）

| 存储能力 | 需要的锁特性 | 不支持的后果 |
|----------|--------------|--------------|
| CAS（条件更新） | 互斥、可重入、公平排队 | 无法保证互斥（无原子比较） |
| 可靠时间源（GetTime） | 租约过期、看门狗续租 | 租约不可靠 → 用固定 TTL + 手动续租 |
| 原子条件删除 | 释放锁（无墓碑） | 降级为墓碑（LockCount=0），语义不变 |
| Watch/订阅 | 跨实例通知（Subscribe） | 降级为轮询（PollingLockWatcher） |

> 判定你的存储是否可做分布式锁：必须**至少满足 CAS + 可靠时间**。
> 只有 CRUD 能力（如纯对象存储、无条件的 KV）不适合直接承载锁语义。
> 详见 [go-storage 的 Capability](../go-storage/storage.go)。

## 下一步

- 完整语义测试套件：[go-storage-lock-test](../go-storage-lock-test)（本地 6 介质 + 外部服务按环境变量跳过）
- 混沌测试（分区/抖动/时钟漂移/并发交接）：[chaos_test.go](../go-storage-lock-test/chaos_test.go)
- HTTP 锁服务（生产化网关）：[storage-lock-http-api](../storage-lock-http-api)
- 可观测性：Prometheus 指标 + Web 看板 [go-storage-lock-prometheus](../go-storage-lock-prometheus)、审计日志 [go-storage-lock-auditlog](../go-storage-lock-auditlog)
