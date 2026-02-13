# Module 11: 生产级工程化 (Production Readiness)

在完成了核心功能开发后，我们进入了"从能用到好用"的跨越阶段。本模块关注那些在 Demo 中容易被忽视，但在生产环境中决定生死的非功能性需求。

## 1. 结构化日志 (Structured Logging)

### 1.1 痛点 (The Pain)

传统的 `log.Printf` 输出的是非结构化的文本：
`2026/02/12 10:00:00 User 1 deposited 100 BTC`

当我们需要在 Kibana 或 Datadog 中回答 "过去一小时有多少笔大于 10 BTC 的充值？" 时，文本日志几乎无法检索。

### 1.2 解决方案 (The Solution)

我们引入了 **Uber Zap**，将日志输出为机器可读的 JSON 格式。

**代码实现:** `pkg/logger`
我们封装了一个全局 `logger` 包，并替换了 `main.go` 中的标准 log。

```go
// 之前
log.Printf("User %s (ID: %d) 的 BTC 充值地址: %s", user.Username, user.ID, btcAddr)

// 现在
logger.Info("BTC 充值地址",
    zap.String("username", user.Username),
    zap.Uint64("uid", uint64(user.ID)),
    zap.String("address", btcAddr),
)
```

**输出效果:**

```json
{
  "level": "info",
  "ts": "2026-02-12T19:29:04.761+0800",
  "caller": "wallet-server/main.go:232",
  "msg": "BTC 充值地址",
  "username": "admin",
  "uid": 1,
  "address": "1QAUQ4opPaGfnow7qBMVcmMhYg9Ubv33x9",
  "index": 1
}
```

### 1.3 优势

1.  **可索引**: 字段如 `uid`, `address` 自动成为索引字段。
2.  **上下文**: 自动携带 `caller` (文件名:行号) 和 `ts` (时间戳)。
3.  **高性能**: Zap 是 Go 生态中性能最高的 Logger 之一 (Zero Allocation)。

## 2. 优雅停机 (Graceful Shutdown)

### 2.1 痛点 (The Pain)

在 K8s 或 Docker 环境中，服务重启是常态。如果直接 `kill -9` 或强制退出：

1.  **数据不一致**: 正在处理的充值请求可能只执行了一半（例如：扣除了未确认数，但未增加余额）。
2.  **客户端报错**: 正在调用的 API 请求会直接断开连接 (Connection Reset)。

### 2.2 解决方案 (The Solution)

我们通过 `os/signal` 监听系统信号 (`SIGINT`, `SIGTERM`)，收到信号后执行以下步骤：

1.  **停止接收新请求**: HTTP/gRPC Server 调用 Shutdown/GracefulStop。
2.  **等待当前请求完成**: 设置一个超时时间（如 5秒），等待所有 Handler 执行完毕。
3.  **释放资源**: 关闭数据库连接、MQ 消费者等。

**代码实现:** `cmd/wallet-server/main.go`

```go
// 监听信号
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit // 阻塞等待

// 优雅关闭 HTTP
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    logger.Fatal("HTTP Server Force Shutdown", zap.Error(err))
}

// 优雅关闭 gRPC
grpcServer.GracefulStop()
```

这种机制确保了在滚动更新 (Rolling Update) 期间，用户请求零失败，数据零丢失。

## 3. 配置管理 (Configuration Management)

### 3.1 痛点 (The Pain)

之前我们使用 `os.LookupEnv` 硬编码在 `main.go` 中读取环境变量。

1.  **难以管理**: 几十个环境变量散落在代码各处，不知道到底有哪些配置。
2.  **缺乏默认值**: 每次启动都要设一大堆 env，否则报错。
3.  **不支持配置文件**: 开发环境想用 `config.yaml`，生产环境想用 `ENV`，难以统一。

### 3.2 解决方案 (The Solution)

我们引入了 **Viper**，实现了分层配置管理。
优先级: `环境变量 (ENV)` > `配置文件 (config.yaml)` > `默认值 (Defaults)`。

**代码结构:**

- `pkg/config/config.go`: 定义配置结构体 `Config`，并提供全局单例 `Global`。
- `config.yaml`: 默认配置文件 (Git 追踪，作为模板)。

**使用方式:**

```go
// 之前
dbHost := getEnv("DB_HOST", "localhost")

// 现在
dbHost := config.Global.DB.Host
```

**Docker 部署友好:**
Viper 自动将环境变量映射到结构体。例如 `config.Global.Redis.Addr` 可以通过环境变量 `REDIS_ADDR` (我们在代码中设置了 Replacer) 或者 `REDIS_ADDR` 来覆盖。
_(注: 我们在 `config.go` 中使用了 `envKeyReplacer` 将 `.` 替换为 `_`，所以 `redis.addr`对应`REDIS*ADDR`)*

## 4. API 文档自动生成 (API Documentation)

### 4.1 痛点 (The Pain)

1.  **文档滞后**: 代码改了，文档没改，前端对着过时文档开发，联调时火葬场。
2.  **维护成本高**: 手写 Markdown 耗时耗力。
3.  **无法测试**: 文档只是纯文本，不能直接发起请求测试。

### 4.2 解决方案 (The Solution)

我们引入了 **Swagger (OpenAPI 2.0)**，让代码即文档。

**代码实现:**
我们在 Handler 上添加注释：

```go
// @Summary Check system health
// @Router /health [get]
func HealthCheck(c *gin.Context) { ... }
```

然后运行 `swag init`，即可自动生成文档。

**访问地址:**
`http://localhost:8080/swagger/index.html`

### 4.3 效果

1.  **自动同步**: 每次发版前运行一次 `swag init`，文档永远与代码保持一致。
2.  **交互式调试**: 可以直接在网页上点击 "Try it out" 发送请求，替代 Postman。

## 5. 数据库版本化迁移 (Database Versioned Migrations)

### 5.1 痛点 (The Pain)

1.  **AutoMigrate 不可控**: `gorm.AutoMigrate` 黑盒执行，不知道具体改了啥。
2.  **不可逆**: 想要回滚一个字段修改？AutoMigrate 做不到。
3.  **数据丢失风险**: 生产环境自动变更 Schema 极度危险。

### 5.2 解决方案 (The Solution)

我们引入了 **golang-migrate**，采用 "Up/Down" 脚本模式管理变更。

**代码实现:**

1.  **Migrations**: `migrations/` 目录下存放 SQL 文件。
    - `000001_init_schema.up.sql`: 建表
    - `000001_init_schema.down.sql`: 删表
2.  **CLI Tool**: `cmd/migrate/main.go` 封装了迁移逻辑。

**使用方式:**

```bash
# 执行升级
go run cmd/migrate/main.go -cmd up

# 回滚一步
go run cmd/migrate/main.go -cmd down
```

### 5.3 生产环境部署

在 K8s 中，我们不再依赖 App 启动时自动迁移。而是将其作为一个 **initContainer** (Job) 在 Deployment 更新前执行。
这样确保了：**先升级数据库，再启动新版代码**。

### 5.4 最佳实践 Q&A

**Q: 系统有几十张表，000001_init_schema.up.sql 会不会太大？**
**A**: `000001` 是系统的"基石" (Baseline Snapshot)，通常会包含项目初始化时的所有表结构，文件大一点是正常的。

后续的变更（如 V1.1 版本）则会是**增量**的：

- `000002_add_kyc_table.up.sql`
- `000003_add_nft_support.up.sql`

**Q: 结构体 (Go Structs) 也要写在一个文件里吗？**
**A**: **绝对不要！**
虽然 SQL 初始化脚本在 001 里可能很大，但 Go 代码必须按领域拆分。
我们目前的 `internal/model/` 目录就是最佳实践：

- `models.go`: 用户与账户核心
- `transaction.go`: 充提记录
- `collection.go`: 归集逻辑
- `outbox.go`: 消息队列表

保持 Go 代码的模块化，与 SQL 迁移脚本的"时间线"管理，是两个维度的概念。

## 6. 总结 (Summary)

至此，我们的钱包服务已经具备了不仅"能用"，而且"好用"、"敢用"的生产级特质：

1.  **结构化日志 (Zap)**: 让问题可被检索。
2.  **优雅停机**: 让部署零故障。
3.  **配置管理 (Viper)**: 让环境可配置。
4.  **API 文档**: 让协作更顺畅。
5.  **数据库迁移**: 让数据更安全。
