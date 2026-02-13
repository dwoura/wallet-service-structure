# 模块 1 核心概念备忘录：密码学基础

这是模块 1 代码 (`pkg/safe_random`, `pkg/crypto_util`, `pkg/kms`) 背后的核心设计思想。

## 1. 为什么不能用 `math/rand`?

_代码位置: `pkg/safe_random/random.go`_

- **场景**: 生成私钥、生成 Salt、生成 Nonce。
- **错误做法**: 使用 `math/rand`。它是**伪随机**的，种子 (Seed) 相同，结果一定相同。黑客可以预测你的下一个"随机数"。
- **正确做法**: 我们使用了 `crypto/rand`。它直接读取操作系统的熵源 (如 `/dev/urandom`)，这是物理层面的随机，不可预测。

## 2. 对称加密：为何选择 AES-GCM?

_代码位置: `pkg/crypto_util/symmetric.go`_

- **AES-CBC vs AES-GCM**:
  - **CBC**: 需要手动处理填充 (Padding)，容易遭受 "Padding Oracle Attack"。
  - **GCM (Galois/Counter Mode)**: 现代标准。它不仅加密 (保密性)，还自带**认证 (完整性)**。如果你修改了密文中的任何一个比特，解密时会直接报错，而不是吐出乱码。
- **实战**: 我们的 `LocalKMS` 使用 AES-GCM 来加密存储私钥。

## 3. 哈希算法的选择

_代码位置: `pkg/crypto_util/hash.go`_

- **MD5/SHA1**: **已死**。存在碰撞漏洞，严禁用于安全场景。
- **SHA-256**: 行业标准 (Bitcoin 使用)。
- **Keccak-256**: Ethereum 专用标准 (与标准 SHA-3 略有不同)。
- **Blake3**: 新一代高性能哈希，速度极快，适合文件校验或非链上场景。

## 4. KMS (密钥管理系统) 的核心思想

_代码位置: `pkg/kms`_

我们定义了 `KeyManager` 接口：

```go
type KeyManager interface {
    CreateKey() ...
    Sign(keyID, data) ... // 注意：这里不需要传入私钥！
}
```

- **思想**: **私钥不离身**。
- 业务层只知道 `KeyID` ("我有把钥匙叫 key-123")，但永远拿不到钥匙本身。业务层只能请求 KMS："请帮我用 key-123 签个名"。
- 这是所有企业级钱包（交易所、托管）的安全基石。
