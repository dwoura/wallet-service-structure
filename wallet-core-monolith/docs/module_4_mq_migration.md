# 模块 4: 消息队列架构演进 (Redis Streams -> Kafka)

本模块记录了我们将消息队列从轻量级的 Redis Streams 迁移到企业级的 Apache Kafka 的全过程。这不仅是代码的替换，更是架构思维的升级。

## 1. 为什么要升级？ (架构对比)

| 特性           | Redis Streams                | Apache Kafka                  | 适用场景                               |
| :------------- | :--------------------------- | :---------------------------- | :------------------------------------- |
| **持久化**     | 内存为主，定期落盘 (RDB/AOF) | 磁盘顺序写，持久化能力极强    | Redis 适合短时缓冲，Kafka 适合永久日志 |
| **吞吐量**     | 高 (取决于内存/CPU)          | 极高 (百万级 TPS)             | 大数据、日志聚合必选 Kafka             |
| **消费模式**   | Consumer Groups              | Consumer Groups (更成熟)      | 两者概念类似                           |
| **消息回溯**   | 支持 (依赖内存容量)          | 支持 (支持几天甚至几年的回溯) | 审计、重放数据时 Kafka 优势巨大        |
| **运维复杂度** | 低 (只要有 Redis)            | 高 (依赖 Zookeeper/Kraft)     | 中小团队首选 Redis                     |

**升级理由**:
对于一个交易所级别的钱包系统，资金变动的**可靠性 (Reliability)** 和 **可追溯性 (Traceability)** 至关重要。Kafka 提供的磁盘持久化和强大的重放能力，能保证即使系统全面崩溃，我们也能从通过重放 Kafka 消息来重建账户余额。

## 2. 迁移策略 (Migration Strategy)

由于我们在 **Module 3** 中明智地定义了 `mq.Producer` 接口，本次迁移对业务代码的侵入极小。

### 接口回顾

```go
type Producer interface {
    Publish(ctx context.Context, topic string, payload []byte) error
}
```

### 代码变更点

1.  **新增实现**: 创建 `pkg/wallet/service/mq/kafka_producer.go` 实现上述接口。
2.  **依赖注入**: 在 `cmd/wallet-server/main.go` 中，将 `NewRedisProducer` 替换为 `NewKafkaProducer`。
3.  **配置调整**: 引入 Kafka 连接配置 (Doker Compose)。

## 3. 核心难点与面试题 (Key Challenges)

在实现 Kafka Producer 时，我们必须处理以下问题：

### A. 消息丢失 (Message Loss)

- **Redis**: 如果 Redis 宕机且 AOF 没刷盘，消息就丢了。
- **Kafka**: 我们可以配置 `acks=all` (所有 ISR 副本确认) 来保证**零丢失**。
  - _Code Review 重点_: 检查 Producer 初始化时的 `RequiredAcks` 配置。

### B. 消息重复 (Duplication)

- **场景**: 扫描器 Crash 重启，重新处理了 Block #100，再次发送 "充值到账" 消息。
- **解决**:
  - **下游幂等性**: 消费者 (Account Service) 必须检查 `tx_hash` 是否已处理过。这是**最重要**的防御防线。
  - **Kafka 幂等生产**: 开启 `Idempotent: true` (仅限单个分区内)。

### C. 顺序性 (Ordering)

- **需求**: 同一个用户的充值和提现必须按顺序处理吗？
- **Kafka**: 只要保证同一个 `UserID` 发送到同一个 Partition (分区)，Kafka 就能保证顺序。
  - _Code Review 重点_: `Publish` 方法是否正确使用了 `Key` (Partition Key)。

---

## 4. 实战代码 (Implementation)

接下来的步骤中，我们将引入 `github.com/segmentio/kafka-go` 库，并编写 Kafka 版本的实现。
