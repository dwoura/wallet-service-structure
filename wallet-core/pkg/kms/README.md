# KMS (Key Management Service) 模块

本模块定义了 `KeyManager` 接口及其实陈。

## 设计目标

我们采用了 **面向接口编程 (Interface-based Design)** 的方式。这允许我们的钱包业务逻辑层（交易构建、签名请求）与底层的密钥存储方式解耦。

## 实现方案对比

根据安全性从高到低：

| 等级          | 方案                                    | 说明                                           | 本项目支持情况                                 |
| :------------ | :-------------------------------------- | :--------------------------------------------- | :--------------------------------------------- |
| **L1 (最高)** | **CloudHSM / Hardware HSM**             | 使用物理硬件存储私钥。私钥不可导出。           | 接口支持 (需实现 AWS CloudHSM 适配器)          |
| **L2**        | **TEE (Trusted Execution Environment)** | Intel SGX / AWS Nitro Enclaves。内存加密隔离。 | 接口支持 (需实现 Enclave RPC 客户端)           |
| **L3**        | **Cloud KMS (AWS/GCP KMS)**             | 云厂商托管的密钥服务。                         | 接口支持 (需实现 AWS KMS SDK 适配器)           |
| **L4 (最低)** | **Local Encrypted Storage**             | 使用 AES 加密私钥存数据库/文件。               | **本项目主要实现目标** (`LocalKMS` + 持久化层) |

## 当前实现 (`LocalKMS`)

目前代码中的 `LocalKMS` 是一个 **内存版 (In-Memory)** 实现。

- **存储**: `map[string]*keyEntry`
- **生命周期**: 重启后丢失 (非持久化)
- **用途**: 仅用于单元测试、CI/CD 环境以及本地开发调试。

## 未来演进

在接下来的课程中，我们将为 `LocalKMS` 增加 **AES 加密持久化** 功能，使其升级为完整的 **L4 方案**。
这意味着：

1. 这里不存储明文私钥。
2. 启动时需要输入 `MasterKey` (主密钥)。
3. 私钥被 `MasterKey` 加密后存入 PostgreSQL 数据库。
