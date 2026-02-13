# Go 工程架构最佳实践：路由与任务的扩展性设计

随着项目规模的扩大，把所有路由写在一个文件，或把所有定时任务写在一个 Service 里，会导致代码难以维护。本文档介绍在 Go 项目中通过**模块化**和**分层**来解决这些问题的最佳实践。

---

## 1. 路由层的扩展性 (Scalable Routing)

### 现状问题

如果将几百个 API 的路由注册全写在 `internal/server/router.go` 的 `NewHTTPRouter` 方法中，这个文件会变得巨大且难以阅读。

### 最佳实践：模块化路由 (Modular Routing)

将路由注册逻辑分散到各个业务模块中。

**推荐目录结构：**

```text
internal/
  server/
    router.go          // 主路由入口
    routes/            // [NEW] 路由模块目录
      user_routes.go   // 用户模块路由
      wallet_routes.go // 钱包模块路由
      trade_routes.go  // 交易模块路由
```

**代码示例：**

1.  **定义路由注册接口** (可选，或直接用函数)：

    ```go
    // internal/server/routes/user_routes.go
    package routes

    import "github.com/gin-gonic/gin"

    func RegisterUserRoutes(rg *gin.RouterGroup) {
        user := rg.Group("/users")
        {
            user.POST("/register", handler.Register)
            user.GET("/profile", handler.GetProfile)
        }
    }
    ```

2.  **主路由文件聚合**：

    ```go
    // internal/server/router.go
    func NewHTTPRouter() *gin.Engine {
        r := gin.Default()

        v1 := r.Group("/api/v1")
        {
            routes.RegisterUserRoutes(v1)   // 注册用户模块
            routes.RegisterWalletRoutes(v1) // 注册钱包模块
        }
        return r
    }
    ```

### 最佳实践：模块化 gRPC (Modular gRPC)

同理，gRPC 服务注册也不应该全部堆积在 `internal/server/grpc.go`。

**代码示例：**

1.  **定义 gRPC 注册函数**：

    ```go
    // internal/server/routes/grpc_address.go
    func RegisterAddressGRPC(s *grpc.Server, svc service.AddressService) {
        pb.RegisterAddressServiceServer(s, handler_grpc.NewAddressHandler(svc))
    }
    ```

2.  **主 Server 聚合**：

    ```go
    // internal/server/grpc.go
    func NewGRPCServer(addressService service.AddressService) *grpc.Server {
        s := grpc.NewServer()

        routes.RegisterAddressGRPC(s, addressService) // 注册 Address 服务
        // routes.RegisterWalletGRPC(s, walletService) // 注册更多服务...

        return s
    }
    ```

### 关于 Swagger 注释杂乱

Swagger 注释必须写在 Handler 函数上方，这确实会让 Handler 文件变长。
**解决方案：**

1.  **接受现状**：这是 Go Swagger 的标准做法，大多数项目都有一半篇幅是注释。
2.  **分离文档文件 (Advanced)**：虽然 swag 支持 `// @name` 等引用，但维护成本高，不推荐。
3.  **代码生成**：使用 protobuf/gRPC 定义 API，然后自动生成 Swagger (OpenAPI) 文件，这是微服务架构的终极解法（完全不写注释）。

---

## 2. 定时任务与后台任务 (Cron & Background Workers)

### 现状问题

`internal/service/cron.go` 适合轻量级任务（如每分钟同步汇率）。但如果遇到以下场景，它就不够用了：

1.  **任务耗时极长**：例如 "给 10 万用户发空投"，会阻塞 Cron 调度器。
2.  **任务量巨大**：每秒几千个任务。
3.  **需要重试/死信队列**：任务失败了怎么办？

### 最佳实践：任务队列 (Task Queue)

对于重型任务，不要用 `Cron` 直接跑，而是用 **"生产者-消费者"** 模型。

**架构演进：**

- **Level 1: Cron (当前)** -> 适合由于时间触发的轻量任务。
- **Level 2: Cron + Job Queue** -> Cron 只负责**生产**任务消息，Worker 负责**执行**。

**推荐方案：Asynq (基于 Redis 的分布式任务队列)**

引入 `github.com/hibiken/asynq`。

**目录结构：**

```text
internal/
  worker/              // [NEW] 专门的 Worker 模块
    processor.go       // 任务处理器 (Consumer)
    distributor.go     // 任务分发器 (Producer)
    tasks/             // 任务定义
      send_email.go
      sync_chain.go
```

**工作流程：**

1.  **Cron (Producer)**: 每分钟触发一次，但它不做具体业务，只是往 Redis 队列里推一条消息 `{"type": "sync_rates"}`。
2.  **Worker (Consumer)**: 取出消息，执行具体逻辑。支持并发控制（如限制 10 个并发）、自动重试（失败 3 次后进死信队列）。
3.  **单独部署**: Worker 可以独立于 API Server 部署，横向扩展。例如部署 10 个 Worker Pods 专门处理后台任务。

### 总结建议

1.  **路由**：拆分文件，按业务模块注册。
2.  **文档**：拥抱 Swagger 注释，或者转向 gRPC/Protobuf 自动生成。
3.  **任务**：
    - **简单定时** -> robfig/cron + 分布式锁。
    - **高并发/耗时** -> Asynq (Redis 队列)，Cron 只充当触发器。
