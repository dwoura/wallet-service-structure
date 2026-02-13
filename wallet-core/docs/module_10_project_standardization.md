# Module 10: Go 工程化标准与目录重构 (Project Standardization)

## 1. 为什么我们需要重构？ (Why)

在项目初期，为了快速验证想法，我们将大部分代码都堆在了 `pkg` 或 `cmd` 下。随着功能增加（引入了 MQ、分布式事务、多种加密算法），项目变得臃肿且职责不清。

**当前存在的问题：**

- **边界模糊**: `pkg/wallet/service` 里面既有业务逻辑（`Deposit`），又有通用库（`bip39`）。外部项目引用时会引入不必要的依赖。
- **私有性缺失**: 核心业务逻辑（如资金归集）应该是私有的 (`internal`)，不应被外部项目导入。
- **难以维护**: 缺乏统一的配置管理和错误的层级划分。

## 2. 标准 Go 项目布局 (The Standard Layout)

我们将遵循 Go 社区公认的 `golang-standards/project-layout` 进行重构。

### 重构目标 (Target Structure)

```text
wallet-core/
├── cmd/
│   └── wallet-server/    # [入口] main.go, 只负责初始化和依赖注入
├── internal/             # [私有代码] 不允许被外部 import
│   ├── model/            # [模型] 数据库 Struct (User, Deposit)
│   ├── service/          # [业务逻辑] AddressService, SweeperService, RelayService
│   ├── handler/          # [接口层] HTTP/gRPC Handler (未来)
│   └── config/           # [配置] 统一配置加载
├── pkg/                  # [公共库] 可被外部使用的通用代码
│   ├── database/         # DB/Redis 连接辅助
│   ├── bip39/            # 助记词算法
│   ├── bip32/            # HD Wallet 算法
│   └── utils/            # 通用工具
├── api/                  # [协议定义] Protobuf, Swagger
└── docs/                 # [文档] 学习报告
```

## 3. 重构日志 (Change Log)

### 3.1 移动通用库 (Moving Libraries)

- `pkg/wallet/bip39` -> `pkg/bip39`
- `pkg/wallet/bip32` -> `pkg/bip32`
- `pkg/database` -> 保持不变

### 3.2 下沉业务逻辑 (Internalizing Services)

- `pkg/wallet/service/*` -> `internal/service/*`
  - _原因_: 钱包的具体业务逻辑（如何归集、如何扫描）是应用私有的，放入 `internal` 可以强制编译器禁止外部引用，保护核心代码。

### 3.3 修复与清理 (Fixes & Cleanup)

- **恢复 Address 包**: 由于误删，重新在 `pkg/address` 实现了 BTC/ETH 地址生成逻辑。
- **Build 修复**: 清理了 Go Build Cache (`go clean -cache`) 以解决 `address` 包的幽灵依赖问题。
- **依赖更新**: 所有 import 路径已更新为新结构。

## 4. 当前结构 (Current Structure)

```text
wallet-core/
├── cmd/wallet-server/    (Entry point)
├── internal/
│   ├── service/          (Address, Sweeper, Relay, Observer)
│   └── model/            (DB Models)
├── pkg/
│   ├── address/          (BTC/ETH Addr Generators)
│   ├── bip32/            (HD Wallet)
│   ├── bip39/            (Mnemonic)
│   ├── database/         (DB Connection)
│   ├── errno/            [NEW] (Global Error Codes)
│   └── utils/
```

## 5. API 统一响应标准 (Unified API Response)

为了让前端/移动端对接更加顺滑，我们定义了标准的 JSON 响应格式和错误码规范。

### 5.1 JSON 结构 (`internal/handler/response`)

```json
{
  "code": 0,           // 业务状态码 (0=成功, 非0=错误)
  "msg": "Success",    // 提示信息
  "data": { ... }      // 业务数据
}
```

### 5.2 全局错误码 (`pkg/errno`)

| Code  | Message               | 说明                  |
| :---- | :-------------------- | :-------------------- |
| 0     | Success               | 成功                  |
| 10001 | Internal server error | 服务器内部错误        |
| 20101 | User not found        | 用户不存在 (业务错误) |

### 5.3 实践 (Practice)

我们在 `cmd/wallet-server/main.go` 中集成了一个轻量级的 Gin HTTP Server，并实现了 `/health` 接口来演示这一标准。

````go
```go
// HealthCheck handler
func HealthCheck(c *gin.Context) {
    response.Success(c, gin.H{"status": "UP"})
}
````

## 6. RPC 通信与微服务 (gRPC & Protobuf)

为了在大规模微服务架构中实现高性能通信，我们也引入了 gRPC。

### 6.1 定义服务 (`api/proto/address.proto`)

```protobuf
service AddressService {
  rpc GetAddress (GetAddressRequest) returns (GetAddressResponse);
}
```

我们定义了 `AddressService`，它暴露了 `GetAddress` 方法，用于生成或获取充值地址。

### 6.2 代码生成与实现

使用 `protoc` 生成 Go 代码后，我们在 `internal/handler/grpc/address_handler.go` 中实现了服务接口：

```go
type AddressHandler struct {
    pb.UnimplementedAddressServiceServer
    service service.AddressService
}
```

这种设计模式将 **传输层 (gRPC)** 与 **业务逻辑层 (Service)** 完美解耦。

### 6.3 服务集成

最终，我们在 `main.go` 中启动了一个独立的 gRPC Server (默认端口 50051)，与 HTTP Server (8080) 并行运行，对外提供服务。

## 7. 进阶：还有哪些工程化标准？(What's Missing?)

为了达到**真正的企业级生产可用 (Production Ready)**，我们还需要补齐以下短板：

### 7.1 结构化日志 (Structured Logging)

- **现状**: 使用 `log.Printf`，输出是纯文本，难以被 ELK (Elasticsearch/Logstash/Kibana) 或 Datadog 解析。
- **标准**: 使用 **Uber Zap** 或 **Slog**。
- **格式**: JSON。
  ```json
  {
    "level": "info",
    "ts": 16800000,
    "caller": "service/payment.go:42",
    "msg": "payment received",
    "amount": 100,
    "currency": "BTC"
  }
  ```
- **价值**: 可以根据字段检索日志（例如：`currency="BTC" AND amount > 10`）。

### 7.2 优雅停机 (Graceful Shutdown)

- **现状**: 按 Ctrl+C 强制杀掉进程。如果正在处理充值请求，可能会导致数据库数据不一致。
- **标准**: 监听系统信号 (`SIGINT`, `SIGTERM`)。
- **行为**: 收到信号后，停止接收新请求，等待当前正在处理的请求（如 HTTP Handler, MQ Consumer）完成后，再断开 DB 连接并退出。
- **价值**: **零数据丢失**，对金融系统至关重要。

### 7.3 配置管理 (Configuration Management)

- **现状**: 使用 `os.LookupEnv`。配置项分散，不支持 YAML/JSON 配置文件，不支持热加载。
- **标准**: 使用 **Viper**。
- **能力**: 优先级管理 (Flag > Env > Config File > Default)，支持动态监听配置变更。

### 7.4 输入校验 (Input Validation)

- **现状**: 手动 `if req.Amount < 0`。容易遗漏。
- **标准**: 使用 **go-playground/validator**。
- **方式**: Struct Tag。
  ```go
  type Request struct {
      Email string `validate:"required,email"`
      Age   int    `validate:"gte=18"`
  }
  ```

### 7.5 接口文档 (Swagger/OpenAPI)

- **现状**: 口头沟通或 Markdown。前端不知道改了字段。
- **标准**: 使用 **swag** 注释生成 OpenAPI Spec。
- **价值**: 自动生成可交互的 API 文档界面。

## 8. 路由层重构 (Router Refactoring)

### 8.1 痛点

随着 API 增多，`main.go` 变得越来越臃肿，包含了几十行路由注册代码，违反了单一职责原则。

### 8.2 解决方案

我们将路由注册逻辑剥离到 `internal/router/router.go`。

- **main.go**: 只负责启动服务。
- **router.go**: 负责路由表维护、中间件注册。

```go
// main.go
r := router.NewHTTPRouter()
```

这将大大提高代码的可维护性。
