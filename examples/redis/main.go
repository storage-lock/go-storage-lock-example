// 示例：用 Redis 存储实现分布式锁。
//
// 运行方式：
//
//	STORAGE_LOCK_REDIS_URI=redis://localhost:6379/0 go run ./examples/redis
//
// Redis 存储天然适合分布式锁：单机部署时靠原子脚本（Lua CAS + TIME）保证互斥，
// 集群/哨兵下由 Redis 本身提供一致性。这是生产环境最常用的分布式锁介质之一。
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	redis_locks "github.com/storage-lock/go-redis-locks"
)

func main() {
	uri := os.Getenv("STORAGE_LOCK_REDIS_URI")
	if uri == "" {
		uri = "redis://localhost:6379/0" // 默认本机，可被环境变量覆盖
	}

	ctx := context.Background()

	// 第一步：创建一把锁（传 URI + lockId）
	lock, err := redis_locks.NewRedisLockByURI(ctx, uri, "example-redis-lock")
	if err != nil {
		panic(err)
	}

	// 第二步：加锁（同一时刻只有一个 owner 能成功）
	owner := "owner-A"
	if err := lock.Lock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("Redis Lock success, owner =", owner)

	// 持有期间做"临界区"工作。看门狗会自动续租，不必手动管租约。
	time.Sleep(200 * time.Millisecond)

	// 第三步：解锁。只有持有者能解开自己的锁。
	if err := lock.UnLock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("Redis UnLock success")
}
