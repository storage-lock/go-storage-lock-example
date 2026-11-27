// 示例：用 etcd 存储实现分布式锁。
//
// 运行方式：
//
//	STORAGE_LOCK_ETCD_ENDPOINTS=127.0.0.1:2379 go run ./examples/etcd
//
// etcd 用事务（Txn CAS + 版本号）保证互斥，是 Kubernetes 生态的标准分布式锁介质。
// 比 Redis 更强的线性一致性，适合对一致性要求极高的场景。
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	etcd_locks "github.com/storage-lock/go-etcd-locks"
)

func main() {
	endpointsStr := os.Getenv("STORAGE_LOCK_ETCD_ENDPOINTS")
	if endpointsStr == "" {
		endpointsStr = "127.0.0.1:2379"
	}
	endpoints := strings.Split(endpointsStr, ",")

	ctx := context.Background()

	// 第一步：创建一把锁（传 endpoints 列表 + lockId）
	lock, err := etcd_locks.NewEtcdLockByEndpoints(ctx, endpoints, "example-etcd-lock")
	if err != nil {
		panic(err)
	}

	// 第二步：加锁
	owner := "owner-A"
	if err := lock.Lock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("Etcd Lock success, owner =", owner)

	// 临界区工作
	time.Sleep(200 * time.Millisecond)

	// 第三步：解锁
	if err := lock.UnLock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("Etcd UnLock success")
}
