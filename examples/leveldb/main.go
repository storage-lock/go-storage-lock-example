// 示例：用 LevelDB 存储实现分布式锁。
//
// 运行方式：
//
//	go run ./examples/leveldb
//
// LevelDB 是嵌入式 KV：单机进程内共享一把锁、重启后锁记录仍在磁盘上。
// 适合"单机多进程/多实例"场景（比如同一台机器上的多个 worker），
// 不需要独立部署数据库。它是文件系统之外最常见的本地持久锁介质。
package main

import (
	"context"
	"fmt"
	"time"

	leveldb_locks "github.com/storage-lock/go-leveldb-locks"
)

func main() {
	ctx := context.Background()

	// 第一步：创建一把锁（传 db 路径 + lockId）。路径不存在会自动创建。
	lock, err := leveldb_locks.NewLevelDBLockByPath(ctx, "/tmp/storage-lock-leveldb-example", "example-leveldb-lock")
	if err != nil {
		panic(err)
	}

	// 第二步：加锁
	owner := "owner-A"
	if err := lock.Lock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("LevelDB Lock success, owner =", owner)

	// 临界区工作
	time.Sleep(200 * time.Millisecond)

	// 第三步：解锁
	if err := lock.UnLock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("LevelDB UnLock success")
}
