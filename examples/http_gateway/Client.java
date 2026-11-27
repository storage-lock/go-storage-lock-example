import lock.storage.LockClient;
import lock.storage.LockClient.AcquireResult;
import lock.storage.LockClient.LockBusyException;

/**
 * 跨语言示例：Java 客户端连接 Go 起的 HTTP 锁服务。
 *
 * 用法（先起服务端）：go run ./examples/http_gateway
 * 然后：javac -d /tmp Client.java && java -cp /tmp Client
 *
 * 演示 acquire / renew / inspect / release 完整流程，以及互斥（409）。
 */
public class Client {
    public static void main(String[] args) throws Exception {
        String base = args.length > 0 ? args[0] : "http://127.0.0.1:18080";
        LockClient client = new LockClient(base);
        String lockId = "cross-lang-demo-java";

        // 1. 获取锁
        AcquireResult r = client.acquire(lockId, "java-a", "5s");
        System.out.println("[java] acquire ok owner=" + r.ownerId
                + " token=" + r.fencingToken + " lease=" + r.leaseExpire);

        // 2. 查看锁
        String info = client.inspect(lockId);
        System.out.println("[java] inspect: " + info);

        // 3. 续租
        AcquireResult renewed = client.renew(lockId, "java-a");
        System.out.println("[java] renew ok new_lease=" + renewed.leaseExpire);

        // 4. 演示互斥：另一个 owner 抢锁应收到 409
        try {
            client.acquire(lockId, "java-rival", "5s");
            System.out.println("[java] FAIL: rival should not acquire while we hold");
            System.exit(1);
        } catch (LockBusyException expected) {
            System.out.println("[java] rival correctly got 409 LockBusy");
        }

        // 5. 释放
        client.release(lockId, "java-a");
        System.out.println("[java] release ok");

        // 6. 释放后他人可再获取
        AcquireResult r2 = client.acquire(lockId, "java-after", "5s");
        System.out.println("[java] re-acquire after release ok token=" + r2.fencingToken);
        client.release(lockId, "java-after");

        System.out.println("JAVA-OK");
    }
}
