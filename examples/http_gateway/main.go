// 示例：跨语言分布式锁——Go 起 HTTP 锁服务，Python/Java 客户端连接使用。
//
// 这是"锁服务网关"模式：Go 进程持有存储介质（memory/redis/mysql/etcd...），
// 通过无状态 HTTP 接口把锁语义暴露出去。非 Go 语言不需要重新实现 CAS/租约/
// fencing token，直接调 REST 接口即可与其他语言互操作。
//
// 运行方式：
//
//	go run ./examples/http_gateway            # 默认 memory 存储，监听 127.0.0.1:18080
//	go run ./examples/http_gateway -addr :8080 -storage redis
//
// 然后另开终端：
//
//	python3 ./examples/http_gateway/client.py      # Python 客户端抢同一把锁
//	javac -d /tmp ./examples/http_gateway/Client.java && java -cp /tmp Client
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/storage-lock/go-memory-storage"
	storage_lock "github.com/storage-lock/go-storage-lock"
	admin "github.com/storage-lock/storage-lock-http-api/admin"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:18080", "listen address")
	storageKind := flag.String("storage", "memory", "storage backend: memory|redis|mysql|etcd")
	flag.Parse()

	// 构造存储。生产环境把这里换成对应的 go-*-storage 仓库即可，
	// 锁服务本身不感知介质差异。
	storage := memory_storage.NewMemoryStorage()
	_ = storageKind // 切换存储的代码示例见各 examples/<backend>/main.go

	// 构造锁服务。NewService 内部强制租约模式（DisableWatchDog，无状态 HTTP
	// 服务不跑本地看门狗），客户端在租约内自己 renew 续租。
	service, err := admin.NewService(storage, "cross-lang", storage_lock.StorageLockOptions{})
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	// 挂在 / 下，所有客户端共用同一把锁的命名空间。
	mux.Handle("/", service.Handler())

	fmt.Printf("storage-lock HTTP gateway listening on http://%s (storage=%s)\n", *addr, *storageKind)
	fmt.Println("Python 客户端：python3 ./examples/http_gateway/client.py")
	fmt.Println("Java 客户端：  javac -d /tmp ./examples/http_gateway/Client.java && java -cp /tmp Client")

	go func() {
		if err := http.ListenAndServe(*addr, mux); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "server error:", err)
			os.Exit(1)
		}
	}()

	// 优雅退出
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}
