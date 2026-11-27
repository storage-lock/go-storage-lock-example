// 示例：用 MySQL 存储实现分布式锁。
//
// 运行方式：
//
//	STORAGE_LOCK_MYSQL_DSN="root:pass@tcp(127.0.0.1:3306)/test?parseTime=true&charset=utf8mb4&loc=Local" go run ./examples/mysql
//
// 不设环境变量时默认连 127.0.0.1:13307（本地 storage-lock 测试容器，root:rootpass）。
//
// MySQL 走的是通用 SQL 基类：条件 UPDATE（CAS）保证互斥。所有 SQL 家族存储
// （postgres/mariadb/tidb/sqlserver/oracle 等 13 个仓库）用法完全一致，只是
// 构造函数不同。这是企业内网最常见的分布式锁介质。
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	mysql_locks "github.com/storage-lock/go-mysql-locks"
)

func main() {
	dsn := os.Getenv("STORAGE_LOCK_MYSQL_DSN")
	if dsn == "" {
		// 默认指向本地 storage-lock 测试用的 MySQL 容器（root:rootpass，端口 13307）。
		// 换成你的 MySQL 即可：root:密码@tcp(host:3306)/库名?parseTime=true&charset=utf8mb4&loc=Local
		dsn = "root:rootpass@tcp(127.0.0.1:13307)/storage_lock_test?parseTime=true&charset=utf8mb4&loc=Local"
	}

	ctx := context.Background()

	// 第一步：创建一把锁（传 DSN + lockId）。首次使用会自动建表（go-mysql-storage.Init）。
	lock, err := mysql_locks.NewMysqlLockByDsn(ctx, dsn, "example-mysql-lock")
	if err != nil {
		panic(err)
	}

	// 第二步：加锁
	owner := "owner-A"
	if err := lock.Lock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("MySQL Lock success, owner =", owner)

	// 临界区工作
	time.Sleep(200 * time.Millisecond)

	// 第三步：解锁
	if err := lock.UnLock(ctx, owner); err != nil {
		panic(err)
	}
	fmt.Println("MySQL UnLock success")
}
