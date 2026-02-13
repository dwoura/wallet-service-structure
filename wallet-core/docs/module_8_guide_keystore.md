# 深度解析: 本地 Keystore 与加密存储 (Module 8-1)

> **教学目标**: 既要理解钱包私钥的安全存储原理 (KDF/AES)，又要掌握如何使用工具生成 keystore 并配置服务。

## 1. 理论基础 (Theoretical Basis)

在区块链世界中，**"Not your keys, not your coins"** 是至理名言。如果私钥以明文形式存储在服务器上（如环境变量、数据库），一旦服务器被攻破，资金将瞬间归零。

为了安全地存储私钥，我们通常采用 **"密码保护的加密文件" (Passphrase Protected Keystore)** 方案。这涉及两个核心密码学概念：

### 1.1 密钥派生函数 (KDF - Key Derivation Function)

用户输入的密码通常较短且熵值低（如 "123456"），容易被暴力破解（Brute Force）或彩虹表攻击。
**KDF** 的作用是将用户的密码转换为一个强壮的、固定长度的加密密钥。

我们主要使用 **Scrypt** 算法，它的特点是 **"内存硬" (Memory-Hard)**。

- 它在计算过程中需要大量内存，使得攻击者难以使用 ASIC 矿机或 GPU 进行大规模并行破解。
- 参数解释:
  - `N`: CPU/内存消耗因子 (262144 -> 2^18)。值越大，计算越慢，破解越难。
  - `r`: 块大小 (8)。
  - `p`: 并行度 (1)。
  - `Salt`: 随机盐值，防止彩虹表攻击。

### 1.2 对称加密 (Symmetric Encryption)

有了 KDF 派生的密钥后，我们使用对称加密算法来加密私钥（或助记词）。

- **AES-256-GCM** (Galois/Counter Mode):
  - **256**: 密钥长度 256位。
  - **GCM**: 不仅提供加密（保密性），还通过 MAC 提供**完整性校验** (Integrity)。如果密文被篡改，解密会直接失败，而不会得到乱码。

---

## 2. 代码实现解析 (Code Analysis)

我们的 `pkg/keystore` 包实现了上述逻辑。

### 2.1 数据结构设计

我们参考了 [Ethereum Keystore V3 标准](https://github.com/ethereum/wiki/wiki/Web3-Secret-Storage-Definition)，定义了如下 JSON 结构：

```go
// pkg/keystore/keystore.go

type EncryptedKeyJSON struct {
    Crypto  CryptoJSON `json:"crypto"`
    Id      string     `json:"id"`      // UUID，文件的唯一标识
    Version int        `json:"version"` // 版本号，通常为 3
}

type CryptoJSON struct {
    Cipher       string       `json:"cipher"`       // 加密算法名: "aes-256-gcm"
    CipherText   string       `json:"ciphertext"`   // 核心数据: 被加密后的助记词 (Hex String)
    CipherParams CipherParams `json:"cipherparams"` // 加密参数: IV (初始化向量)
    KDF          string       `json:"kdf"`          // KDF算法名: "scrypt"
    KDFParams    KDFParams    `json:"kdfparams"`    // Scrypt 参数 (N, r, p, salt)
    MAC          string       `json:"mac"`          // 校验码 (用于验证密码是否正确)
}
```

### 2.2 加密流程 (`EncryptMnemonic`)

让我们逐行分析 `EncryptMnemonic` 函数的实现逻辑：

```go
func EncryptMnemonic(mnemonic, password string) (*EncryptedKeyJSON, error) {
    // 1. 生成随机 Salt (32字节)
    // Salt 的作用是让相同的密码在不同文件里生成的 Key 也不同，防止批量破解。
    salt := make([]byte, 32)
    io.ReadFull(rand.Reader, salt)

    // 2. KDF: 密码 + Salt -> 派生密钥 (Derived Key)
    // scrypt.Key(password, salt, N, r, p, keyLen)
    // 我们生成 32 字节的 Key，正好用于 AES-256。
    derivedKey, _ := scrypt.Key([]byte(password), salt, 262144, 8, 1, 32)

    // 3. 准备 AES-GCM
    block, _ := aes.NewCipher(derivedKey)
    gcm, _ := cipher.NewGCM(block)

    // 4. 生成随机 Nonce (IV)
    // GCM 模式下，对于同一个 Key，Nonce 绝对不能重复，否则会泄露密钥流！
    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)

    // 5. 执行加密
    // Seal(dst, nonce, plaintext, additionalData)
    // 结果包含了 tag (MAC)，保证了不可篡改性。
    ciphertext := gcm.Seal(nil, nonce, []byte(mnemonic), nil)

    // 6. 计算 MAC (SHA256(derivedKey + ciphertext))
    mac := sha256.Sum256(append(derivedKey, ciphertext...))

    // 7. 组装 JSON 对象返回
    return &EncryptedKeyJSON{...}, nil
}
```

---

## 3. 实战指南 (Practical Usage)

### 3.1 初始化钱包 (生成 wallet.json)

我们提供了 `wallet-cli` 工具来生成并加密钱包。

**Step 1: 编译工具**

```bash
go build -o wallet-cli ./cmd/wallet-cli
```

**Step 2: 交互式初始化**

```bash
./wallet-cli init -o wallet.json
```

**交互输出示例**:

```text
正在初始化新钱包...
请设置一个强密码来保护您的助记词。
输入密码: *******
确认密码: *******

✅ 钱包已初始化！
文件位置: wallet.json
您的 ID: 5f8d...
⚠️  警告: 请务必记住您的密码！如果丢失密码，您将无法恢复钱包。
```

生成的 `wallet.json` 就是加密后的文件，可以安全地存储在磁盘上。

### 3.2 配置 Wallet Server

`wallet-server` 启动时会自动检测配置中的 `keystore_path`。

**方式 A: Docker Compose (推荐)**

在 `docker-compose.yml` 中：

1.  挂载 `wallet.json` 到容器内。
2.  设置环境变量 `WALLET_PASSWORD`。

```yaml
version: "3.8"
services:
  wallet-server:
    volumes:
      - ./wallet.json:/app/wallet_keystore.json # 挂载文件
    environment:
      - WALLET_KEYSTORE_PATH=/app/wallet_keystore.json
      - WALLET_PASSWORD=your_secure_password # 提供解密密码
```

**方式 B: 本地开发**

```bash
export WALLET_PASSWORD="your_password"
# 确保 config.yaml 中 wallet.keystore_path 指向正确文件
go run cmd/wallet-server/main.go
```

---

## 4. 故障排除 (Troubleshooting)

- **Error: "未找到 Keystore 文件..."**:
  - 检查 `config.yaml` 或环境变量中的路径配置是否正确。
  - 确保先运行了 `./wallet-cli init`。
- **Error: "解密 Keystore 失败..."**:
  - 密码错误：请检查 `WALLET_PASSWORD` 环境变量。
  - 文件损坏：检查 JSON 文件格式是否完整。
- **Error: mac check failed**:
  - 通常意味着密码错误，或者文件内容被篡改。

---

## 5. 面试考点 (Interview Questions)

1.  **问: 为什么不用简单的 SHA256(password) 作为密钥？**
    - **答**: SHA256 计算太快了。普通 GPU 每秒能算数亿次，黑客可以轻易暴力破解 6-8 位的复杂密码。Scrypt 通过消耗大量内存，迫使破解者无法并行加速。

2.  **问: 如果我忘记了 Keystore 密码，能通过技术手段找回吗？**
    - **答**: **绝对不能**。 AES-256 目前被认为是计算安全的。

3.  **问: 这个 Keystore 可以在 MetaMask 导入吗？**
    - **答**: 结构上兼容 V3，但主要字段我们存的是 `mnemonic` (12个单词) 而不是 `private key` (Hex)。直接导入可能需要适配，但核心加密逻辑是一致的。
