package main

import (
	"context"
	"fmt"

	memory_locks "github.com/storage-lock/go-memory-locks"
)

// 快速开始：用内存存储演示分布式锁的基本用法（Lock/RLock/WLock/Subscribe）。
// 内存存储仅适用于单进程演示，生产环境请换成 mysql/redis/etcd 等持久化存储——
// 各存储的用法都是"构造 Storage → NewStorageLock → Lock/UnLock"同一套模式，
// 具体 DSN 写法参考各 go-*-locks 仓库的 examples 目录。
func main() {
	ctx := context.Background()

	// 第一步：创建一把锁（内存版无需 DSN，直接给 lockId）
	lock, err := memory_locks.NewLock(ctx, "example-lock")
	if err != nil {
		panic(err)
	}

	// 第二步：加写锁
	if err := lock.Lock(ctx, "owner-A"); err != nil {
		panic(err)
	}
	fmt.Println("Lock success")
	defer func() {
		if err := lock.UnLock(ctx, "owner-A"); err != nil {
			panic(err)
		}
		fmt.Println("UnLock success")
	}()

	// 第三步：读锁（与写锁互斥，此处写锁未释放会阻塞，故只演示订阅）
	ch, err := lock.Subscribe(ctx, "example-lock")
	if err != nil {
		panic(err)
	}
	defer lock.Unsubscribe("example-lock", ch)
	select {
	case info := <-ch:
		fmt.Println("Subscribe got:", info.ToJsonString())
	default:
		fmt.Println("Subscribe channel ready")
	}

	// 第四步：读锁单独演示（先释放写锁的场景请参考 go-storage-lock-test 的语义测试）
	rwLock, err := memory_locks.NewLock(ctx, "example-rw-lock")
	if err != nil {
		panic(err)
	}
	if err := rwLock.RLock(ctx, "reader-1"); err != nil {
		panic(err)
	}
	fmt.Println("RLock success")
	if err := rwLock.RUnLock(ctx, "reader-1"); err != nil {
		panic(err)
	}
	fmt.Println("RUnLock success")
}
