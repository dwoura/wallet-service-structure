# 模块 7: 分布式事务与数据一致性 (Distributed Transactions)

## 1. 核心问题: 双写一致性 (The Dual Write Problem)

在我们的 `EthObserver` 中，有一段看似完美的代码：

```go
// 1. 写入数据库
db.Create(&deposit)

// 2. 发送 MQ 消息
producer.Publish(..., deposit)
```

**这就引发了经典的"双写问题" (Dual Write Problem)：**

1.  **情况 A (MQ 失败)**: 数据库写入成功了，用户看到余额加了，但 MQ 挂了（或网络抖动）。
    - _后果_: `OrderService` 收不到消息，无法触发后续业务（如自动购买理财），导致**状态不一致**。
2.  **情况 B (DB 失败)**: 先发 MQ 成功了，但数据库事务回滚了。
    - _后果_: 下游收到了"充值成功"的消息处理了业务，但上游其实根本没入账。**资金损失！**

## 2. 业界解决方案

### 方案一：二阶段提交 (2PC / XA)

- **原理**: 强一致性协议，数据库和 MQ 都支持 Prepare/Commit。
- **缺点**: 性能极差，MQ 中间件支持度不一，**不推荐**用于互联高并发场景。

### 方案二：这是我们采用的 -> 事务性发件箱 (Transactional Outbox) / 本地消息表

这是最成熟、在大厂应用最广泛的**最终一致性**方案。

**核心思想**: 将 "发 MQ" 这个动作，变成 "写 DB" 的动作。

#### 流程设计：

1.  **开启数据库事务 (Begin TX)**
2.  **业务操作**: `INSERT INTO deposits ...`
3.  **记录消息**: `INSERT INTO outbox_messages (topic, payload, status) VALUES (..., 'PENDING')`
    - _关键点_: 这两步在**同一个事务**里！要么都成功，要么都失败。
4.  **提交事务 (Commit TX)**
5.  **异步 Relay (中继)**:
    - 有一个独立的 `MessageRelayService` (或者后台 Goroutine)。
    - 轮询 `outbox_messages` 表中状态为 `PENDING` 的消息。
    - 发送给 Kafka。
    - 发送成功后，更新状态为 `SENT`。

## 3. 实战代码规划 (Implementation Plan)

我们将在接下来的步骤中实现这个模式：

1.  **定义模型**: 创建 `model.OutboxMessage` 结构体。
2.  **改造 Observer**:
    - 删除直接调用的 `producer.Publish`。
    - 改为 `db.Create(&outboxMsg)`。
3.  **创建 Relay Service**:
    - 启动一个 Goroutine，每 100ms 扫描一次数据库。
    - 负责将消息搬运到 Kafka。

---

> **面试金句**:
> "在分布式系统中，我们无法保证跨资源的强一致性（CAP 定理），但我们可以通过**本地消息表 (Local Message Table)** 配合 **至少一次投递 (At-least-once Delivery)**，来实现**最终一致性 (Eventual Consistency)**。"
