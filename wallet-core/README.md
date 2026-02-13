# Blockchain Wallet Core 🪙

这是一个生产级的区块链中心化钱包后端服务，采用 Go 语言开发，遵循业界最佳工程实践。

## 📚 核心文档

- [📅 学习任务清单](../学习任务清单.md): 本项目的开发路线图。
- [🏗️ 最佳实践指南 (Engineering Guide)](./docs/guide_best_practices.md): **必读**。包含了项目结构、命名规范、Server/Router 设计、测试标准等核心决策。
  - [目录结构](./docs/guide_best_practices.md#1-目录结构最佳实践-directory-structure)
  - [App vs Admin 结构](./docs/guide_best_practices.md#16-app-c端-与-admin-后台-的工程结构)
  - [Cmd 与脚本规范](./docs/guide_best_practices.md#17-cmd-目录与脚本最佳实践)
- [🗄️ 数据库设计](./docs/schema_design.md)
- [🛳️ 生产环境部署](./docs/module_11_production_readiness.md)
- [📈 高级扩展性 (Scalability)](./docs/guide_backend_asynq.md): Asynq 任务队列与模块化路由。
- [📊 业务监控 (Observability)](./docs/guide_monitor_business.md): Prometheus 业务指标埋点指南。

## 🚀 快速开始

```bash
# 1. 启动依赖 (Postgres, Redis, Kafka)
docker-compose up -d

# 2. 运行迁移
go run cmd/migrate/main.go

# 3. 启动服务
go run cmd/wallet-server/main.go
```

## 🛠️ 工具集

- **CLI 工具**: `go run cmd/wallet-cli/main.go`
- **测试**: `go test ./internal/...`
