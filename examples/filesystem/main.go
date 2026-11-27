// 示例：用文件系统存储实现分布式锁。
//
// 运行方式：
//
//	go run ./examples/filesystem
//
// 文件系统用 flock 保证互斥（同机多进程可靠），锁记录写在指定目录下。
// 适合"零依赖、单机多进程"的场景——不需要 Redis/MySQL，文件即存储。
// 注意：flock 只在本机生效，跨机器互斥请用 redis/mysql/etcd。
package main

import (
	"context"
	"fmt"
	"time"

	filesystem_storage "github.com/storage-lock/go-filesystem-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
)

func main() {
	ctx := context.Background()

	// 第一步：构造文件系统存储（指定工作目录），再创建锁。
	storage := filesystem_storage.NewFileSystemStorage("/tmp/storage-lock-fs-example")
	lock, err := storage_lock.NewStorageLock(storage, "example-fs-lock")
	if err != nil {
		panic(err)
	}

	// 第二步：加锁
	owner := "owner-A"
	if err := lock.Lock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("FileSystem Lock success, owner =", owner)

	// 临界区工作
	time.Sleep(200 * time.Millisecond)

	// 第三步：解锁
	if err := lock.UnLock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("FileSystem UnLock success")
}
