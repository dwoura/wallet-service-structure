# 模块 7: Kafka 架构升级备忘录

## 1. "之前的 Kafka" vs "现在的配置" (Architecture Evolution)

你提到的 "之前弄的 Kafka" (指 `docker-compose.kafka.yml`) 和 "现在的表" (指主 `docker-compose.yml`) 的区别在于：**孤岛 vs 生态**。

| 特性       | 之前的 (docker-compose.kafka.yml)              | 现在的 (docker-compose.yml)                                        |
| :--------- | :--------------------------------------------- | :----------------------------------------------------------------- |
| **定位**   | **独立组件 (Standalone)**                      | **全栈集成 (Integrated)**                                          |
| **用途**   | 仅用于测试 Kafka 是否能启动，能否收发消息。    | 真实的业务环境。App, DB, Kafka 在同一个网络里。                    |
| **谁在用** | 没人用。它孤零零地跑着，没有生产者给它发消息。 | **Wallet Server** 是它的生产者/消费者。                            |
| **网络**   | 只有宿主机能访问 (localhost:9094)。            | **双向联通**: 内部 App 用 `kafka:29092`，外部用 `localhost:9094`。 |

**现在的配置做了什么？**

1.  **合并 (Merge)**: 把 Zookeeper 和 Kafka 的定义搬到了主文件中。
2.  **打通 (Link)**: 在 `wallet-server` 里加了 `depends_on: kafka`，确保 Kafka 先启动。
3.  **切换 (Switch)**: 注入了环境量 `MQ_TYPE=kafka`，告诉 App："别连 Redis 了，去连 Kafka"。

---

## 2. 为什么需要"双监听器" (Dual Listeners)?

你在 `docker-compose.yml` 里看到了一大坨 `KAFKA_ADVERTISED_LISTENERS` 配置，这是为了解决 **"内外有别"** 的问题：

- **内部 (Inside Docker)**: `wallet-server` 访问 `kafka:29092`。
- **外部 (Outside Docker)**: 你在 Mac 终端运行 Go 测试脚本时，访问 `localhost:9094`。

如果不这么配，通常会遇到 "连上了但发不出去" 的经典 Kafka 坑。

---

## 3. 常用命令速查 (Cheat Sheet) 🛠️

**启动/重启所有服务 (包含 Kafka)**:

```bash
# --build 确保重新编译 Go 代码
docker-compose up -d --build
```

**查看钱包服务日志**:

```bash
# -f 实时跟随日志
docker-compose logs -f wallet-server
```

**查看 Kafka 内部日志 (调试用)**:

```bash
docker-compose logs -f kafka
```

**停止并移除所有容器**:

```bash
docker-compose down
```
