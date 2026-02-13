# Module 14: 微服务拆分 - 踩坑与故障排查指南

在将单体架构拆分为微服务的过程中，我们遇到了一些典型的问题。本指南记录了这些问题及其解决方案，作为后续开发的参考。

## 1. 结构体重复定义冲突

### 问题描述

编译时报错：`Withdrawal is redeclared in this block`。
原因是在 `internal/model/models.go` 和 `internal/model/transaction.go` 中都定义了 `Withdrawal` 结构体。

### 解决方案

保留业务逻辑更相关的定义（`transaction.go`），删除 `models.go` 中的冗余定义。

**经验总结**:

- 在进行代码迁移或重构时，务必检查目标包中是否已存在同名结构体。
- 保持 `model` 层的单一数据源（SSOT）原则。

## 2. Go 接口与指针的类型陷阱 (Type Mismatch)

### 问题描述

在 `cmd/wallet-service/main.go` 中，`loadMasterKey` 函数最初的签名是：

```go
// 错误写法
func loadMasterKey() (*bip32.ExtendedKey, error) { ... }
```

编译器报错：

> `*bip32.ExtendedKey` does not implement `bip32.ExtendedKey` (type `*bip32.ExtendedKey` is pointer to interface, not interface)

### 原因分析

`bip32.ExtendedKey` 本身是一个 **Interface** (接口)。
在 Go 语言中，接口变量已经包含了底层数据的指针（(value, type) tuple）。
声明 `*Interface`（指向接口的指针）通常是错误的且没有必要的，除非你需要修改接口变量本身指向的对象。

错误链条：

1. `NewSQLAddressService` 需要 `bip32.ExtendedKey` (接口)。
2. `loadMasterKey` 返回了 `*bip32.ExtendedKey` (指向接口的指针)。
3. Go 不会自动解引用接口指针来匹配接口。

### 解决方案

修改函数签名，直接返回接口类型：

```go
// 正确写法
func loadMasterKey() (bip32.ExtendedKey, error) { ... }
```

调用时也无需解引用：

```go
// 错误: *masterKey
// 正确: masterKey
addrSvc, err := service.NewSQLAddressService(..., masterKey, ...)
```

## 3. 依赖注入参数顺序错误

### 问题描述

在初始化 `NewSQLAddressService` 时，传递的参数顺序与函数签名不匹配，导致编译错误或运行时 panic。

### 解决方案

严格对照函数签名检查参数顺序。

```go
// 签名
func NewSQLAddressService(db *gorm.DB, rdb *redis.Client, masterKey bip32.ExtendedKey, network *chaincfg.Params, cache cache.Cache)

// 调用
service.NewSQLAddressService(db, rdb, masterKey, netParams, c)
```

**经验总结**:

- 当函数参数超过 3 个时，确实容易出错。
- 更好的设计是使用 `Option` 模式或 `Config` 结构体来传递参数，例如 `NewSQLAddressService(config AddressServiceConfig)`。

## 4. 包导出 (Export) 问题

### 问题描述

`wallet-core/pkg/cache` 包中，我们想使用 `NewChainCache`，但发现该函数未导出（首字母小写）或根本不存在，只导出了 `NewMemoryCache`。

### 解决方案

检查 `pkg/cache` 源码，确认导出的构造函数。暂时使用 `NewMemoryCache` 作为替代，确保存储层可用。

**经验总结**:

- 在拆分包（Package Split）时，要注意哪些函数需要 `public` (首字母大写) 给外部 cmd 使用。

## 5. 广播服务 (Broadcaster) 的逻辑重复

### 问题描述

在 `internal/service/broadcaster_service.go` 中，发现 `deriveHotWalletKey` 方法被定义了两次，或者是与 `wallet-service` 中的逻辑重复。

### 解决方案

删除重复代码。在微服务拆分后，`Broadcaster` 应该独立拥有它需要的逻辑，或者引用共享的 `pkg`。

## 6. 数据库连接 (Connection Refused)

### 问题描述

启动 `broadcaster-worker` 时报错：`dial tcp 127.0.0.1:5433: connect: connection refused`。

### 原因分析

本地开发环境（如 Mac 宿主机）直接运行二进制文件时，它尝试连接 `localhost:5433`。如果 Postgres 是通过 Docker 运行的，需确保 Docker 端口映射正确，且容器正在运行。

### 解决方案

- 确保运行 `docker-compose up -d postgres`。
- 检查 `config.yaml` 中的 DB 端口配置。

---

**总览**: 微服务拆分不仅仅是复制粘贴代码，更多是处理 **依赖解耦**、**类型系统适配** 和 **配置隔离** 的过程。
