# 深度解析: 离线签名架构与实现 (Module 8-2)

> **教学目标**: 理解冷热钱包分离架构的设计思想，并掌握使用 `wallet-cli` 进行全流程操作（构建、签名、广播）。

## 1. 架构原理 (Architecture)

### 1.1 为什么需要离线签名？

即使我们有了 Keystore，如果解密过程发生在连网的服务器上，私钥在内存中依然是**暴露**的。
如果服务器被植入木马，攻击者可以 dump 内存拿到私钥；或者在解密的一瞬间 hook 关键函数。

**离线签名 (Offline/Cold Signing)** 实现了**私钥永不触网**：

1.  **观察钱包 (Watch-only / Online)**: 只有地址和公钥，知道你有多少钱，但**没有支配权**。
2.  **冷钱包 (Cold Wallet / Offline)**: 只有私钥，**没有网络**，不知道世界发生了什么，只负责"盖章签字"。

### 1.2 交易生命周期 (Transaction Lifecycle)

| 步骤             | 执行端      | 动作 | 数据输入                                | 数据输出        | 关键点                            |
| :--------------- | :---------- | :--- | :-------------------------------------- | :-------------- | :-------------------------------- |
| **1. Build**     | **Online**  | 构造 | From, To, Amount, (Nonce/Gas from RPC)  | `unsigned.json` | 此时交易未生效，只是一串数据      |
| **2. Transport** | **USB/QR**  | 传输 | `unsigned.json`                         | `unsigned.json` | 物理隔离 (Air-gap)                |
| **3. Sign**      | **Offline** | 签名 | `unsigned.json` + **Keystore(PrivKey)** | `signed.json`   | **显示屏校验** (Verify on Screen) |
| **4. Transport** | **USB/QR**  | 传输 | `signed.json`                           | `signed.json`   | -                                 |
| **5. Broadcast** | **Online**  | 广播 | `signed.json`                           | TxHash          | 只有这一步才真正上链              |

---

## 2. 代码实现解析 (Code Analysis)

### 2.1 交易构造 (`cmd/build_tx.go`)

Online 端的核心难点在于：**如何在没有私钥的情况下，准备好签名所需的一切参数？**

- **Nonce**: 必须通过 RPC (`eth_getTransactionCount`) 查询链上最新的 Nonce。
- **GasPrice**: 必须查询网络当前拥堵情况。
- **ChainID**: EIP-155 要求签名中包含 ChainID，防止重放攻击。

```go
type UnsignedTransaction struct {
    From    string
    To      string
    Amount  string
    Nonce   uint64 // 必须由 Online 端填好
    // ...
    DerivationPath string // 关键: 告诉冷钱包用哪个子私钥签名 (如 "m/44'/60'/0'/0/0")
}
```

### 2.2 离线签名 (`cmd/sign.go`)

这是最核心、最敏感的部分。

**安全机制 1: 屏幕校验 (Verify on Screen)**
代码中必须显式打印交易详情，供人工核对：

```go
fmt.Println("================ 待签名交易 ================")
fmt.Printf("To:     %s\n", unsignedTx.To)
fmt.Printf("Amount: %s\n", unsignedTx.Amount)
// ...
```

**技术实现 2: 路径派生 (Path Derivation)**

```go
// 1. 解密获得 Mnemonic -> Seed -> MasterKey
wallet, _ := bip32.NewMasterKeyFromSeed(seed, ...)

// 2. 根据路径派生子私钥
// 例如 Path: "m/44'/60'/0'/0/0"
derivedKey, _ := wallet.DerivePath(unsignedTx.DerivationPath)
```

**技术实现 3: EIP-155 签名**
使用 `go-ethereum` 的库函数进行标准签名，输出 Hex 格式的 Raw Transaction。

### 2.3 交易广播 (`cmd/broadcast.go`)

Online 端反序列化 Hex 字符串，还原为 `types.Transaction` 对象，通过 RPC 发送。

---

## 3. 操作流程演示 (Step-by-Step Guide)

本节演示如何使用我们将编译好的 `wallet-cli` 来模拟这一过程。

### 第一步: 构造交易 (Online)

在联网机器上，管理员希望发起一笔转账。

```bash
# 生成未签名交易 txn_unsigned.json
./wallet-cli build-tx \
  --from "0x789..." \
  --to "0x123..." \
  --amount "100000000000000000" \
  --nonce 5 \
  --chain-id 1 \
  --output unsigned.json
```

**输出**:

```text
✅ 未签名交易已构造!
文件: unsigned.json
```

### 第二步: 离线签名 (Offline)

将 `unsigned.json` 复制到离线机器（通过 U盘）。

```bash
# 使用本地 Keystore 对交易进行签名
./wallet-cli sign \
  --input unsigned.json \
  --keystore wallet.json \
  --output signed.json
```

**交互输出**:

```text
================ 待签名交易 ================
Chain:      ETH (ID: 1)
From:       0x789...
To:         0x123...
Amount:     100000000000000000
Nonce:      5
Path:       m/44'/60'/0'/0/0
============================================

正在从 wallet.json 加载 Keystore...
请输入 Keystore 密码以确认签名: *******

✅ 签名成功!
TxHash: 0xabc...
已保存到: signed.json
```

### 第三步: 广播交易 (Online)

将 `signed.json` 复制回联网机器。

```bash
# 广播到以太坊网络
./wallet-cli broadcast \
  --input signed.json \
  --rpc https://mainnet.infura.io/v3/YOUR_KEY
```

**输出**:

```text
正在连接 RPC...
正在广播交易 Hash: 0xabc...
✅ 广播成功!
Tx URL: https://etherscan.io/tx/0xabc...
```

---

## 4. 思考与进阶

1.  **多重签名 (Multi-Sig)**:
    - 目前的方案是单签。对于大额资产，通常要求 m-of-n 签名（比如 3个人里至少2个人同意）。这需要智能合约 (Gnosis Safe) 或 MPC 技术支持。

2.  **二维码传输限制**:
    - 如果 Raw Tx 数据量太大，生成的二维码会非常密集。可以使用 `Animated QR` (动态二维码) 解决。

3.  **UTXO 模型 (BTC)**:
    - 如果是 BTC，`UnsignedTransaction` 需要包含所有 Inputs (UTXO) 的详细信息，Online 端需要先凑够 UTXO。
