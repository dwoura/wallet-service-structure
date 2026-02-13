# ADR 002: 单体到微服务的演进策略 (Monolith-First Strategy)

> **状态**: Accepted  
> **日期**: 2026-02-13  
> **决策者**: System Architect  
> **相关文档**: [Module 14: 微服务拆分架构指南](./module_14_guide_microservices_architecture.md)

---

## 背景 (Context)

在学习区块链钱包开发的过程中,我们面临一个关键的架构选择:

**是从一开始就采用微服务架构,还是先构建单体应用,再逐步演进到微服务?**

这个决策直接影响:

- 开发效率 (学习曲线)
- 系统复杂度 (调试难度)
- 未来扩展性 (生产就绪)

---

## 决策 (Decision)

**我们选择 "Monolith-First" (单体优先) 策略。**

即:

1. **阶段 1 (当前)**: 构建模块化的单体应用 (`wallet-core`)
2. **阶段 2 (学习)**: 通过文档和架构图理解微服务拆分
3. **阶段 3 (未来)**: 在生产需求驱动下,逐步拆分为微服务

---

## 理由 (Justification)

### 1. 学习效率优先

**单体架构的学习优势**:

| 维度         | 单体                       | 微服务                             |
| :----------- | :------------------------- | :--------------------------------- |
| **调试**     | 在 IDE 里打断点,一步步跟踪 | 需要分布式追踪工具 (Jaeger)        |
| **测试**     | 直接运行 `go test`         | 需要 Docker Compose 启动 5+ 个容器 |
| **部署**     | `docker run` 一条命令      | 需要 K8s 集群 + Helm Charts        |
| **错误排查** | 看一个日志文件             | 需要聚合 5+ 个服务的日志 (ELK)     |

**实际案例**:

```go
// 单体: 调试充值流程
func TestDepositFlow(t *testing.T) {
    // 1. 启动服务
    server := NewWalletServer()

    // 2. 模拟充值
    observer.ScanBlock(12345)

    // 3. 验证余额更新
    balance := userService.GetBalance(userID)
    assert.Equal(t, 1.5, balance)
}

// 微服务: 同样的测试
func TestDepositFlow(t *testing.T) {
    // 1. 启动 5 个服务 (Observer, User, Wallet, DB, Kafka)
    docker.Compose("up -d")

    // 2. 等待服务就绪
    time.Sleep(30 * time.Second)

    // 3. 模拟充值 (需要跨服务调用)
    observerClient.ScanBlock(12345)

    // 4. 等待消息传递
    time.Sleep(5 * time.Second)

    // 5. 验证余额 (需要调用另一个服务)
    balance := userClient.GetBalance(userID)
    assert.Equal(t, 1.5, balance)
}
```

**结论**: 对于学习阶段,单体架构可以让我们 **专注于业务逻辑**,而不是被分布式系统的复杂性淹没。

---

### 2. 避免过度设计 (YAGNI 原则)

**YAGNI**: You Aren't Gonna Need It (你不会需要它)

**微服务的隐藏成本**:

```
微服务架构的"税"
├── 服务发现 (Consul/Etcd) ────► 学习成本 2 周
├── API Gateway (Kong/APISIX) ──► 学习成本 1 周
├── 分布式追踪 (Jaeger) ────────► 学习成本 1 周
├── 配置中心 (Consul KV) ───────► 学习成本 1 周
├── 服务网格 (Istio) ───────────► 学习成本 4 周
└── 总计: 9 周 (2 个月+)
```

**反问**:

- 我们现在的 QPS 是多少? **< 10** (本地测试)
- 我们需要独立扩容吗? **不需要** (单机足够)
- 我们有多个团队并行开发吗? **没有** (只有一个学习者)

**结论**: 在没有真实需求的情况下引入微服务,是 **过度设计**。

---

### 3. 单体不等于混乱

**关键**: 我们构建的是 **模块化单体 (Modular Monolith)**,而不是 **大泥球 (Big Ball of Mud)**。

**我们的模块化实践**:

```
wallet-core/
├── internal/
│   ├── service/          # 业务逻辑层 (可独立拆分)
│   │   ├── user.go       # 用户服务 (未来 → User Service)
│   │   ├── wallet.go     # 钱包服务 (未来 → Wallet Service)
│   │   ├── observer.go   # 观察者 (未来 → Observer Service)
│   │   ├── sweeper.go    # 归集器 (未来 → Sweeper Service)
│   │   └── broadcaster.go # 广播器 (未来 → Broadcaster Service)
│   ├── handler/          # API 层 (可独立拆分)
│   └── model/            # 数据层 (共享)
└── api/proto/            # gRPC 接口定义 (已为微服务准备)
```

**关键设计**:

- ✅ 每个 `service/*.go` 都是 **独立的包**,没有循环依赖
- ✅ 已经定义了 **gRPC 接口** (`api/proto/*.proto`)
- ✅ 使用 **消息队列** (Kafka) 解耦服务

**结论**: 我们的单体架构 **已经为未来的拆分做好了准备**,拆分时只需要:

1. 把 `internal/service/user.go` 移到新项目 `user-service/`
2. 把函数调用改为 gRPC 调用
3. 部署到独立容器

---

### 4. 业界最佳实践

**Martin Fowler (微服务架构之父) 的建议**:

> "Almost all the successful microservice stories have started with a monolith that got too big and was broken up."  
> (几乎所有成功的微服务案例,都是从一个变得太大的单体开始拆分的。)

**真实案例**:

| 公司        | 初期架构               | 拆分时机                  |
| :---------- | :--------------------- | :------------------------ |
| **Amazon**  | 单体 (Obidos)          | 2001 年,员工 > 1000 人时  |
| **Netflix** | 单体 (DVD Rental)      | 2008 年,用户 > 1000 万时  |
| **Uber**    | 单体 (Python Monolith) | 2014 年,城市 > 100 个时   |
| **Shopify** | 单体 (Ruby on Rails)   | **至今仍是单体** (模块化) |

**结论**:

- 微服务是 **规模化的解决方案**,不是起点
- 过早拆分会导致 **分布式单体** (Distributed Monolith) - 既有微服务的复杂度,又没有单体的简洁性

---

## 演进路径 (Migration Path)

### 阶段 0: 单体 (当前)

```
┌─────────────────────────────────┐
│     wallet-server (单进程)       │
│                                 │
│  ┌─────────────────────────┐   │
│  │  HTTP + gRPC + Workers  │   │
│  └─────────────────────────┘   │
└─────────────────────────────────┘
```

**特点**:

- ✅ 开发快
- ✅ 调试简单
- ❌ 无法独立扩容
- ❌ 私钥和 API 在同一进程 (安全风险)

---

### 阶段 1: 模块化单体 (重构)

```
┌─────────────────────────────────┐
│     wallet-server (单进程)       │
│                                 │
│  ┌─────┐ ┌──────┐ ┌─────────┐  │
│  │User │ │Wallet│ │Observer │  │
│  │Svc  │ │Svc   │ │Svc      │  │
│  └─────┘ └──────┘ └─────────┘  │
│     ▲        ▲         ▲        │
│     └────────┴─────────┘        │
│       (内部函数调用)              │
└─────────────────────────────────┘
```

**特点**:

- ✅ 代码按服务拆分
- ✅ 接口定义清晰 (gRPC Proto)
- ✅ 为拆分做好准备
- ❌ 仍然是单进程

---

### 阶段 2: 微服务 (拆分)

```
┌─────────┐   ┌──────────┐   ┌───────────┐
│  User   │   │  Wallet  │   │  Observer │
│ Service │   │ Service  │   │  Service  │
└────┬────┘   └────┬─────┘   └─────┬─────┘
     │             │                │
     └─────────────┴────────────────┘
              gRPC / Kafka
```

**特点**:

- ✅ 独立部署
- ✅ 独立扩容
- ✅ 故障隔离
- ❌ 复杂度高

---

## 拆分触发条件 (When to Split)

**不要盲目拆分,等到出现以下信号时再拆**:

### 信号 1: 性能瓶颈

```
场景: Observer 扫描区块时占用 100% CPU,导致 HTTP API 响应变慢。
解决: 拆分 Observer 为独立服务,部署到独立机器。
```

### 信号 2: 团队扩张

```
场景: 团队从 1 人增长到 5 人,大家都在改同一个代码库,频繁冲突。
解决: 按团队拆分服务 (Team A 负责 User Service, Team B 负责 Wallet Service)。
```

### 信号 3: 安全需求

```
场景: 要上生产环境了,需要把 Broadcaster (持有私钥) 物理隔离。
解决: 拆分 Broadcaster 为独立服务,部署到隔离机房。
```

### 信号 4: 独立扩容需求

```
场景: 双十一流量暴增,HTTP API 需要扩容到 10 个实例,但 Observer 只需要 1 个。
解决: 拆分后独立扩容。
```

---

## 风险与缓解 (Risks & Mitigation)

### 风险 1: 过早拆分

**症状**:

- 花了 2 个月搭建微服务基础设施
- 业务逻辑还没写完
- 调试一个 Bug 需要启动 5 个服务

**缓解**:

- ✅ 先完成业务逻辑 (Module 1-13)
- ✅ 再学习微服务理论 (Module 14 文档)
- ✅ 最后在真实需求驱动下拆分

---

### 风险 2: 单体变成大泥球

**症状**:

- 所有代码都在 `main.go`
- 循环依赖 (A 依赖 B, B 依赖 A)
- 无法测试 (所有逻辑耦合在一起)

**缓解**:

- ✅ 遵循 **分层架构** (Handler → Service → Model)
- ✅ 使用 **依赖注入** (不要用全局变量)
- ✅ 编写 **单元测试** (强制解耦)

---

## 总结

**我们选择 "Monolith-First" 的核心原因**:

1. **学习效率**: 专注业务逻辑,不被分布式系统复杂性干扰
2. **避免过度设计**: 在没有真实需求时,不引入不必要的复杂度
3. **模块化准备**: 我们的单体已经为未来拆分做好了准备
4. **业界共识**: 几乎所有成功的微服务都是从单体演进而来

**下一步**:

- ✅ 继续完成 Module 1-13 的学习
- ✅ 阅读 Module 14 的架构文档 (理论)
- ⏭️ 等到真正需要时,再动手拆分 (实践)

---

**参考资料**:

- [Martin Fowler: MonolithFirst](https://martinfowler.com/bliki/MonolithFirst.html)
- [Sam Newman: Building Microservices](https://www.oreilly.com/library/view/building-microservices/9781491950340/)
- [Shopify Engineering: Deconstructing the Monolith](https://shopify.engineering/deconstructing-monolith-designing-software-maximizes-developer-productivity)
