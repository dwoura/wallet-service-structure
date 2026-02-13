# Go 工程化最佳实践指南 (The Ultimate Guide)

这是基于 Go 社区标准 (`golang-standards/project-layout`) 和 Google/Uber 内部规范整理的工程目录结构与测试指南。

**最后更新**: 2026-02-13

---

## 1. 目录结构最佳实践 (Directory Structure)

### 1.1 推荐的目录结构

```text
wallet-core/
├── cmd/                  # [入口] 一个目录对应一个二进制文件
│   └── wallet-server/    # 主 HTTP/gRPC 服务
│   └── wallet-cli/       # 运维/工具脚本
├── internal/             # [核心代码] 禁止外部 import
│   ├── handler/          # [API 层] 解析参数 -> 调用 Service -> 返回 DTO
│   │   ├── request/      # 请求 DTO
│   │   ├── response/     # 响应 DTO
│   ├── service/          # [业务层] 复用逻辑 (App 与 Admin 共享)
│   ├── model/            # [数据层] DB Entities (GORM)
│   ├── server/           # [传输层] HTTP/gRPC Server 生命周期管理
│   └── config/           # [配置] Viper
├── pkg/                  # [公共库] 甚至可以独立出去发版给别人用的代码
│   ├── database/         # DB 连接器
│   ├── logger/           # 日志封装
│   └── bip39/            # 算法库
├── tests/                # [集成测试] E2E Tests
└── scripts/              # [构建脚本] Shell/Python (仅用于 CI/Build)
```

### 1.2 关键命名决策 (Naming Decisions)

#### 为什么叫 `internal/server`?

- **依据**: 封装 Server 的启动、监听、优雅停机逻辑。
- **职责**: 它是传输层 (Transport Layer) 的容器，负责组装 Router 和 gRPC Server。它不是 "Service" (业务逻辑)，也不是 "App" (整个应用的大管家)，所以命名为 `server` 是准确的。

#### App (C端) vs Admin (后台)

- **架构模式**: **单体 (Monolith)**。
- **逻辑分离**:
  - **Handler 层隔离**: `internal/handler/app` vs `internal/handler/admin`。
  - **Service 层共享**: 底层业务逻辑 (如 `GetBalance`) 必须共享，保证数据一致性。
  - **Router 隔离**: 使用 `r.Group("/api")` 和 `r.Group("/admin")` 分别挂载不同的中间件 (Middleware)。

**推荐结构**:

```text
internal/
├── handler/
│   ├── app/          # [App API] 面向移动端/网页用户
│   │   └── wallet.go # 只有查询余额、充值、提现
│   └── admin/        # [Admin API] 面向运营/管理员
│       └── audit.go  # 只有管理员能调用的: 审计、冻结账号、手动归集
├── service/          # [Shared Logic] 业务逻辑层是共享的！
│   └── wallet.go     # 包含 GetBalance, FreezeAccount 等所有逻辑
```

**路由注册示例**:

```go
func NewHTTPRouter() *gin.Engine {
    // ...
    // 1. APP 路由组 (普通用户)
    appGroup := r.Group("/api/v1")

    // 2. Admin 路由组 (管理员，强权限)
    adminGroup := r.Group("/admin/v1")
    adminGroup.Use(middleware.AdminAuth())
    // ...
}
```

---

## 2. API 设计规范 (API Design)

### 2.1 DTO (Data Transfer Object) vs Model

**原则**: **严禁在 API 接口中直接暴露 `internal/model` (DB Entities)。**

- **危险性**: 数据库模型包含 `PasswordHash` 等敏感字段，且绑定结构可能导致客户端被动修改数据库字段。
- **最佳实践**: 使用 DTO。
  - **Request**: `internal/handler/request/xxx.go` (带 `binding` 校验标签)
  - **Response**: `internal/handler/response/xxx.go` (带 `json` 格式化标签)

**推荐结构**:

```text
internal/
├── handler/
│   ├── request/      # [NEW] 专门存放请求参数结构体
│   │   └── wallet_request.go
│   ├── response/     # [NEW] 专门存放响应结构体
│   │   └── wallet_response.go
│   └── wallet_handler.go
├── model/            # 仅存放数据库表结构 (User, Account)
```

---

## 3. Cmd 与脚本规范 (CLI & Scripts)

### 3.1 `cmd` 目录结构

Go 项目的标准是 "One Directory, One Binary"。

- `cmd/wallet-server`: 后端服务入口。
- `cmd/migrate`: 数据库迁移工具入口。

### 3.2 脚本写在哪里？

**场景 A: 生产运维工具 (推荐)**

- **方案**: 集成进 `cmd/wallet-cli`。
- **优点**: 能直接复用 `internal/service` 的业务代码，享受 Go 的类型安全。
- **示例**:
  ```bash
  ./wallet-cli admin create --user root    # 业务脚本
  ./wallet-cli contract sync-events        # 数据修复脚本
  ```

**场景 B: CI/构建/Docker 辅助**

- **方案**: 放进 `scripts/` (Shell/Python)。
- **优点**: 简单，无需编译。
- **示例**: `build.sh`, `deploy.py`。

---

## 4. 测试最佳实践 (Testing Strategy)

### 4.1 单元测试 (Unit Tests)

- **位置**: `internal/service/xxx_test.go` (与源码同级)
- **原则**: Mock 掉所有外部依赖 (DB, Redis, HTTP Client)，只测逻辑分支。

**示例**: `internal/service/address_test.go`

```go
func TestGetDepositAddress(t *testing.T) {
    // 1. Mock Database
    mockRepo := new(mocks.MockRepository)
    mockRepo.On("FindUser", 1).Return(&User{ID: 1}, nil)

    // 2. Call Service
    svc := NewAddressService(mockRepo)
    addr, err := svc.GetDepositAddress(1, "BTC")

    // 3. Assert
    assert.NoError(t, err)
    assert.Equal(t, "1A1zP1...", addr)
}
```

### 4.2 集成测试 (Integration Tests)

- **位置**: `tests/integration/xxx_test.go` (独立目录)
- **原则**: 启动真实环境 (Docker Compose)，测试 API 端到端连通性。

**示例**: `tests/integration/wallet_api_test.go`

```go
func TestDepositAPI(t *testing.T) {
    // 1. 准备数据 (可以直接操作 DB)
    db.Create(&User{Username: "test_user"})

    // 2. 发送真实 HTTP 请求
    resp := http.Post("http://localhost:8080/api/v1/deposit", ...)

    // 3. 验证响应
    assert.Equal(t, 200, resp.StatusCode)
}
```

---

## 5. 总结

遵循以上规范，我们可以确保项目在扩展成百上千个接口时，依然保持清晰、可维护。
