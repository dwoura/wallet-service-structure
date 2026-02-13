# Kafka 架构演进: 从 Zookeeper 到 KRaft

你提到的 "KRaft" 是 Kafka 历史上最大的架构升级。这里为你详细拆解它们的区别。

## 1. 传统架构 (Zookeeper Mode)

在 Kafka 2.8 版本之前，Kafka **必须** 依赖 Zookeeper (ZK) 才能运行。

- **Zookeeper 的作用**:
  - **元数据管理**: 记录有哪些 Topic、Partition、Replica 都在哪里。
  - **Controller 选举**: 决定哪个 Broker (节点) 是"老大" (Controller)，负责管理集群。
  - **健康检测**: 谁挂了，Zookeeper 最先知道。

- **缺点**:
  - **运维复杂**: 你要维护两套系统 (Kafka 集群 + ZK 集群)。
  - **性能瓶颈**: 所有的元数据变更都要经过 ZK，当 Topic 达到百万级时，ZK 会成为瓶颈。

## 2. 新一代架构 (KRaft Mode)

**KRaft (Kafka Raft)** 模式移除了对 Zookeeper 的依赖。

- **原理**:
  - Kafka 内部引入了 **Raft 共识算法** (类似于 Etcd/Consul)。
  - 一部分 Kafka 节点被指定为 **Controller Nodes** (控制器节点)，它们自己内部通过 Raft 投票选老大，自己管理元数据。
  - **"Kafka on Kafka"**: 元数据被存储在一个特殊的内部 Topic 中 (`@metadata`)，像处理普通消息一样处理元数据变更。

- **优点**:
  - **架构简单**: 只需要部署 Kafka，不需要 Zookeeper。
  - **启动更快**: 以前重启集群需要从 ZK 加载大量数据，现在几乎秒级。
  - **扩展性更强**: 支持通过数百万个 Partition。

## 3. 面试/生产建议

- **现状**: Kafka 3.3+ 已经标记 KRaft 为 **Production Ready (生产就绪)**。
- **选择**:
  - 如果是**新项目** (Greenfield)，**强烈建议直接用 KRaft**。
  - 如果是**老项目**，通常还在用 Zookeeper，迁移成本较大。

---

## 4. 端口修改说明

为了避免与你本地现有的服务 (`event-services`) 冲突，我将做以下调整：

1.  **Zookeeper 模式 (Legacy)**:
    - Zookeeper: `2181` -> `2182`
    - Kafka: `9092` -> `9094`
2.  **KRaft 模式 (Modern)**:
    - 我们单独创建一个 `docker-compose.kraft.yml`，不需要 Zookeeper，端口设定为 `9095`。

这样你可以自由选择体验哪种架构，且互不干扰。
