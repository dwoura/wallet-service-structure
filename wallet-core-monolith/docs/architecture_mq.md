# 架构决策记录 (ADR): 消息队列选型

## 背景 (Context)

用户提问：**"消息队列不是应该用 Kafka、RocketMQ 这些吗？为什么选 Redis Streams？"**

## 决策 (Decision)

在本项目（区块链钱包学习与实战）中，我们选择 **Redis Streams** 作为主要的消息队列中间件，但保留架构上切换到 Kafka 的可能性。

## 详细对比 (Detailed Comparison)

| 特性           | Kafka / RocketMQ                                                                                    | Redis Streams                                                            | 本项目适用性分析                                                                                              |
| :------------- | :-------------------------------------------------------------------------------------------------- | :----------------------------------------------------------------------- | :------------------------------------------------------------------------------------------------------------ |
| **设计目标**   | **高吞吐、海量积压、持久化**。用于日志收集、大数据分析、像淘宝双11那样级别的削峰填谷。              | **轻量级、实时性、简单**。内置于 Redis 5.0+，无需额外部署。              | **Redis 胜出**。我们是金融系统（钱包），TPS 通常在几千到几万，不需要 Kafka 亿级的吞吐，更看重架构的简洁性。   |
| **部署成本**   | **极高**。需要 Zookeeper (旧版) 或 Kraft 模式。内存占用大 (JVM)，运维复杂（Rebalance, Partition）。 | **极低**。我们已经用了 Redis 做缓存和锁，开箱即用。                      | **Redis 胜出**。对于学习和中小型项目，部署 Kafka 会分散 50% 的精力在运维上。                                  |
| **数据持久化** | **强**。数据落盘，支持回溯消费几天前的数据。                                                        | **中**。虽然支持 AOF/RDB，但在内存受限时可能会丢弃旧消息（取决于配置）。 | **平局**。钱包系统的充值/提现消息需要强持久化，但我们可以通过 Postgres 数据库做最终兜底，由 MQ 仅负责“通知”。 |
| **消费模式**   | Consumer Groups (消费者组)                                                                          | Consumer Groups (完全借鉴了 Kafka 的设计)                                | **平局**。Redis Streams 完美复刻了 Kafka 的消费者组概念，可以实现负载均衡和 ACK 机制。                        |

## 我们的策略 (Strategy)

1.  **Interface Design (接口先行)**:
    我们不会在业务代码里直接 import `github.com/redis/go-redis` 或 `github.com/segmentio/kafka-go`。
    我们会定义一个接口：
    ```go
    type MQ interface {
        Publish(topic string, msg []byte) error
        Subscribe(topic string, handler func([]byte))
    }
    ```
2.  **Implementation (实现)**:
    先写一个 `RedisMQ` 实现类。
    **如果未来你想升级到 Kafka**，只需要写一个 `KafkaMQ` 实现类，业务逻辑一行代码都不用改。

## 结论

对于**学习**和**中小型金融系统**，Redis Streams 是性价比最高的选择。它能让你在 10 分钟内理解 "消费者组"、"ACK"、"Pending List" 等核心概念，而无需被 Kafka 的复杂配置劝退。

**但是**，如果你去字节跳动或蚂蚁金服，核心链路一定是 Kafka/RocketMQ。
