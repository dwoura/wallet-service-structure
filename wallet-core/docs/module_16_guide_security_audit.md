# Module 16: 安全审计指南 (Security Audit)

本模块我们将对钱包系统进行全方位的安全检查。作为金融级应用，安全性是生命线。

## 1. 依赖漏洞扫描 (Dependency Scanning)

Go 官方提供了 `govulncheck` 工具，用于扫描代码中引用的第三方库是否存在已知的 CVE 漏洞。

### 1.1 安装与运行

```bash
# 安装
go install golang.org/x/vuln/cmd/govulncheck@latest

# 扫描当前项目
govulncheck ./...
```

### 1.2 修复策略

如果发现漏洞：

1.  **升级版本**: `go get -u github.com/xxx/xxx`
2.  **寻找替代**: 如果库不再维护，需寻找替代品。

## 2. 代码审计 (Code Audit) checklist

### 2.1 SQL 注入 (SQL Injection)

- **风险**: 拼接 SQL 字符串。
- **检查**: 全局搜索 `fmt.Sprintf` + `Exec/Query`。
- **现状**: 我们使用了 GORM，且通过 `Where("name = ?", name)` 参数化查询，基本杜绝了注入风险。
- **注意**: 避免使用 `db.Raw("SELECT * FROM users WHERE name = '" + name + "'")`。

### 2.2 硬编码密钥 (Hardcoded Secrets)

- **风险**: 私钥、密码写死在代码里，提交到 Git。
- **检查**: 搜索 `password`, `secret`, `private_key`。
- **现状**:
  - `config.yaml` 包含默认密码 (开发环境可接受)。
  - `cmd/broadcaster-worker/main.go` 加载主私钥时，使用了 Keystore + 密码解密，**安全**。
  - **待改进**: 单元测试中有硬编码的 "password123"，生产环境严禁出现。

### 2.3 越权访问 (Broken Access Control)

- **风险**: 用户 A 可以查询用户 B 的余额。
- **检查**: API Handler 中是否校验了 `Ctx UserID` 与 `Request UserID` 的一致性。
- **现状**:
  - `user-service`: `GetUserInfo` 目前只通过 UserID 查询，未校验调用者身份 (由网关层 JWT 负责)。
  - **改进点**: 网关层必须实现 JWT 解析与 UserID 注入，防止 ID 遍历攻击。

## 3. Web 安全

### 3.1 跨域 (CORS)

- 网关需要配置 CORS 中间件，限制允许的 Origin。

### 3.2 限流 (Rate Limiting)

- 防止恶意刷接口 (DDoS)。
- 建议在网关层集成 `golang.org/x/time/rate` 或 Redis 计数器。

## 4. 实战任务

1.  运行 `govulncheck`。
2.  检查 `internal/service/wallet/service.go` 是否存在资金计算精度问题 (Decimal)。
3.  确保 `broadcaster-worker` 的私钥只存在于内存，不落盘 (除 Keystore 外)。
