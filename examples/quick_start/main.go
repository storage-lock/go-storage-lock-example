package main

import (
	"context"
	"fmt"
	"sync"

	memory_locks "github.com/storage-lock/go-memory-locks"
	storage "github.com/storage-lock/go-storage"
)

// 快速开始：用内存存储演示分布式锁的基本用法（Lock/RLock/WLock/Subscribe）。
// 内存存储仅适用于单进程演示。其他后端也只有在其 CAS 和时间源满足核心库前提、
// 且真实部署通过语义验证后才适用于跨节点场景；不能仅因存储类型而视作生产就绪。
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

	// WithLock 将单调递增的 fencing token 交给临界区。
	// 受保护资源必须在同一个事务中校验并记录已接受的最大 token；
	// 此处用进程内状态演示规则，真实服务应在数据库事务/原子条件写中实现。
	fencedLock, err := memory_locks.NewLock(ctx, "example-fenced-lock")
	if err != nil {
		panic(err)
	}
	resource := newFencedResource()
	if err := fencedLock.WithLock(ctx, func(ctx context.Context, token storage.Version) error {
		return resource.Write(ctx, token, "protected value")
	}); err != nil {
		panic(err)
	}
	fmt.Println("Fenced write success")
}

type fencedResource struct {
	mu       sync.Mutex
	maxToken storage.Version
	value    string
}

func newFencedResource() *fencedResource {
	return &fencedResource{}
}

func (r *fencedResource) Write(ctx context.Context, token storage.Version, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if token <= r.maxToken {
		return fmt.Errorf("stale fencing token %d (accepted %d)", token, r.maxToken)
	}
	r.maxToken = token
	r.value = value
	return nil
}
