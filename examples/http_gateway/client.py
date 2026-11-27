#!/usr/bin/env python3
"""跨语言示例：Python 客户端连接 Go 起的 HTTP 锁服务。

用法（先起服务端）：go run ./examples/http_gateway
然后：python3 ./examples/http_gateway/client.py

演示 acquire / renew / release / inspect 完整流程，以及两个客户端互斥：
另一个客户端抢同一把锁会收到 409 LockBusy。
"""
import os
import sys
import time

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "..", "python-storage-lock"))
from lock_client import LockClient, LockBusyError


def main() -> None:
    base = os.environ.get("STORAGE_LOCK_HTTP", "http://127.0.0.1:18080")
    client = LockClient(base)

    # 1. 获取锁
    lock_id = "cross-lang-demo"
    owner = f"python-{os.getpid()}"
    result = client.acquire(lock_id, owner_id=owner, ttl="5s")
    print(f"[python] acquire ok owner={result['owner_id']} token={result['fencing_token']} lease={result['lease_expire']}")

    # 2. 查看锁
    info = client.inspect(lock_id)
    print(f"[python] inspect: holder={info.get('owner_id')} version={info.get('version')}")

    # 3. 续租
    renewed = client.renew(lock_id, owner)
    print(f"[python] renew ok new_lease={renewed['lease_expire']}")

    # 4. 演示互斥：另一个 owner 抢锁应收到 409
    try:
        client.acquire(lock_id, owner_id="python-rival", ttl="5s")
        print("[python] FAIL: rival should not acquire while we hold")
        sys.exit(1)
    except LockBusyError:
        print("[python] rival correctly got 409 LockBusy")

    # 5. 释放
    client.release(lock_id, owner)
    print("[python] release ok")

    # 6. 释放后他人可再获取
    result2 = client.acquire(lock_id, owner_id="python-after", ttl="5s")
    print(f"[python] re-acquire after release ok token={result2['fencing_token']}")
    client.release(lock_id, "python-after")


if __name__ == "__main__":
    main()
