# Module 5 进阶: 分布式锁与并发安全 (Distributed Lock)

## 1. 为什么需要分布式锁？

在生产环境中，为了高可用 (High Availability)，我们的 `wallet-server` 通常会部署多个实例 (Replicas)。
例如，我们部署了 3 个 `wallet-server` 节点，它们都启动了 `SweeperService`，并且都监听同一个 Redis/Kafka Topic。

### 场景模拟 (Race Condition)

1.  用户充值 1 ETH。
2.  **MQ 发送即使**: "Deposit ID 100 到了"。
3.  **节点 A** 收到消息，开始归集。
    - 查询余额 -> 构造交易 -> **签名** -> **广播 (Nonce=5)**。
4.  **节点 B** 也收到了同样的消息 (可能是 MQ 重发，或者消费者组配置失误)，同时也开始归集。
    - 查询余额 (此时节点 A 的交易还没打包，余额看似没变) -> 构造交易 -> **签名** -> **广播 (Nonce=5)**。

### 后果

- **最好的情况**: 链上节点发现 Nonce 相同，拒收第二笔。但在极短时间内，两笔交易可能进入不同的 Mempool，导致网络拥堵或 Gas 浪费。
- **最坏的情况**: 如果我们在归集逻辑中没有严格检查 Nonce，或者链支持替换交易，可能导致**重复归集** (虽然这里是归集全部余额，不太可能扣两次钱，但会浪费昂贵的手续费)。
- **如果是提现业务 (Withdrawal)**: 后果就是很严重的**双花 (Double Spending)**，给用户转了两次钱！

## 2. 解决方案: Redis 分布式锁

我们使用 Redis 的 `SETNX` (SET if Not eXists) 命令来实现互斥锁。

- **Key**: `lock:sweeper:deposit:<tx_hash>`
- **Value**: 当前节点 ID + 时间戳
- **TTL**: 锁的自动过期时间 (防止死锁)

### 流程

1.  `Sweeper` 收到消息 `TxHash=0x123...`
2.  尝试 `SETNX lock:sweeper:deposit:0x123 ... EX 60`
3.  **成功**: 获得锁，开始执行归集逻辑。执行完释放锁。
4.  **失败**: 说明其他节点正在处理，**直接丢弃**该消息 (或稍后重试)。

## 3. 面试考点 (Interview Q&A)

### Q: 为什么不用 Go 标准库的 `sync.Mutex`?

**A**: `sync.Mutex` 只能锁住**当前进程**内的协程。我们的服务是多实例部署的 (多进程/多服务器)，内存不共享，所以必须借助于外部的共享存储 (Redis/Etcd/Zookeeper) 来实现锁。

### Q: 如果拿到锁的节点挂了 (Crash) 怎么办？

**A**: 这就是为什么必须给锁设置 **TTL (Time To Live/过期时间)**。即使节点挂了没来得及解锁，Redis 会在 TTL 结束后自动删除该 Key，避免死锁。

### Q: Redlock 算法了解吗？

**A**: (进阶) 上面的 `SETNX` 是单节点 Redis 锁。如果 Redis 自身是集群且发生了主从切换，锁可能丢失。Redlock 算法通过向多个独立的 Redis 节点加锁来提高安全性。但对于一般的钱包归集业务，简单的 `SETNX` 配合数据库的唯一索引 (Unique Index) 已经足够安全了。
