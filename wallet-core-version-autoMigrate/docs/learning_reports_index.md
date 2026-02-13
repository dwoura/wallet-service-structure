# 📚 学习报告汇总 (Learning Reports Index)

这里记录了我们在实战过程中针对关键技术点的**详细学习报告**和**架构决策文档**。请定期回顾以巩固知识点。

## ✅ 已完成模块 (Completed Modules)

### 模块 6: 基础设施 (Infrastructure)

- [**Docker 部署与配置**](./module_6_cicd_pipeline.md)
  - _核心知识点_: Dockerfile 多阶段构建, Docker Compose 服务编排, 环境变量注入.

### 模块 7: 企业级架构 (Enterprise Architecture)

- [**Kafka 架构迁移指南**](./module_7_kafka_transition.md)
  - _核心知识点_: Redis vs Kafka 选型对比, 双监听器 (Dual Listeners) 原理, Docker 内部网络, 常用运维命令.
- [**分布式事务 (数据一致性)**](./module_7_distributed_transactions.md)
  - _核心知识点_: 双写问题 (Dual Write Problem), CAP 定理, 本地消息表 (Transactional Outbox) 模式, 最终一致性实现.

### 模块 3: 观察者服务 (Observer)

- [**ETH 区块扫描器设计**](./module_3_observer.md)
  - _核心知识点_: Worker Pool 并发模型, 区块回滚处理, 幂等性入库.

---

## 🚧 进行中/计划中 (Planned)

### 模块 10: 工程化标准 (Standards)

- [**Go 工程化目录标准 & API 规范**](./module_10_project_standardization.md)
  - _核心变更_: `pkg/wallet` 拆分为 `internal/service` 和 `pkg/address`.
  - _新增规范_: Standard Project Layout, Unified JSON Response, Global Error Codes.
  - _新功能_: 集成了 Gin HTTP Server (`/health`).

### 模块 11: 生产级工程化 (Production Readiness)

- [**生产级工程化指南**](./module_11_production_readiness.md)
  - _结构化日志_: Zap JSON Logging.

### 模块 8: 安全加固 (Security)

- _(待创建)_ **离线签名 (Offline Signing)**: 这里的"离线"到底有多彻底？
