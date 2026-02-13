# 模块 2 核心概念深度解析：从随机数到钱包地址

在模块 2 中，我们实现了一个命令行钱包 (`wallet-core`)。为了让你真正理解这部分代码，我将代码背后的**密码学标准**和**数据流向**进行了拆解。

## 1. 整体流程图 (The "Big Picture")

我们的 `wallet-cli new` 命令执行了以下 4 个步骤，这也几乎是所有现代区块链钱包（MetaMask, Ledger, Trust Wallet）的标准创建流程：

```mermaid
graph TD
    A[熵 (Entropy)] -->|BIP-39| B[助记词 (Mnemonic)]
    B -->|BIP-39 + Salt| C[种子 (Seed)]
    C -->|BIP-32| D[主私钥 (Master Key)]
    D -->|BIP-44| E[子私钥 (Child Private Key)]
    E -->|ECDSA| F[公钥 (Public Key)]
    F -->|Encoding| G[钱包地址 (Address)]
```

---

## 2. 核心标准详解

### 2.1 BIP-39: 让人类读懂私钥

_代码位置: `pkg/wallet/bip39/mnemonic.go`_

- **问题**: 计算机喜欢的随机数是这样的 `0x1a2b3c...` (128位或256位)，人类根本记不住。
- **解决**: BIP-39 定义了一个包含 2048 个单词的词库。
- **原理**:
  1.  生成的随机数（熵）。
  2.  每 11 个位 (bits) 对应词库中的一个单词 ($2^{11} = 2048$)。
  3.  128位熵 = 12个单词；256位熵 = 24个单词。
- **实战中的我们**:
  ```go
  mnemonic, _ := service.GenerateMnemonic(256) // 生成 24 个词
  // 输出: "witch collapse practice feed shame open despair creek road again ice least..."
  ```

### 2.2 BIP-32: 分层确定性钱包 (HD Wallet)

_代码位置: `pkg/wallet/bip32/wallet.go`_

- **问题**: 如果每个地址都需要备份一个新的私钥，当你交易多了，通过备份私钥来管理资产将是一场噩梦。
- **解决**: BIP-32 允许我们通过**这一个种子 (Seed)**，数学上推导出**无数个**子私钥。只要你记住了助记词（即记住了种子），你就拥有了所有子账号的控制权。
- **核心概念**: `m` 代表主密钥 (Master)，`/` 代表派生下一层。

### 2.3 BIP-44: 多币种路径标准

_代码位置: `cmd/wallet-cli/cmd/new.go`_

- **问题**: HD 钱包可以派生无数子钥匙，但哪一个是 BTC 的？哪一个是 ETH 的？如果大家都乱用，钱包软件就无法互通。
- **解决**: BIP-44 规定了通用的派生路径格式：
  `m / purpose' / coin_type' / account' / change / address_index`
  - **BTC 路径**: `m/44'/0'/0'/0/0` (CoinType = 0)
  - **ETH 路径**: `m/44'/60'/0'/0/0` (CoinType = 60)
  - **TRON 路径**: `m/44'/195'/0'/0/0` (CoinType = 195)

- **代码中的体现**:
  ```go
  btcPath := "m/44'/0'/0'/0/0"
  btcKey, _ := wallet.DerivePath(btcPath) // 自动推导到第 5 层
  ```

---

## 3. 地址生成原理 (Address Generation)

_代码位置: `pkg/wallet/address/`_

拿到公钥后，不同链有不同的生成规则：

| 币种         | 算法步骤                                                         | 关键技术点                                                      | 实战代码                                  |
| :----------- | :--------------------------------------------------------------- | :-------------------------------------------------------------- | :---------------------------------------- |
| **Bitcoin**  | 1. SHA256(PubKey)<br>2. RIPEMD160(Result)<br>3. Base58Check 编码 | **双重哈希 + Base58**: 为了更短且防输错 (包含校验和)            | `btcutil.NewAddressPubKeyHash`            |
| **Ethereum** | 1. Keccak256(PubKey)<br>2. 取后 20 字节<br>3. Hex 编码           | **EIP-55**: 利用大小写 (Checksum) 来防止地址输错。如 `0xAbC...` | `address/eth.go` 中的 `toChecksumAddress` |

---

## 4. 下一步：从工具到系统

到目前为止，我们写的都是**单机工具**（生成器）。
但在交易所或钱包后台中，我们不能手动敲命令生成钱包。

**模块 3 (中心化钱包系统)** 我们将把这些库集成到一个 Web 服务中：

1.  用户注册 -> 分配一个 DB 记录。
2.  调用 `GenerateAddress` -> 为用户生成专属充值地址。
3.  把地址存入 **Postgres 数据库**。
4.  开始监听链上数据。

这才是真正的后端开发开始的地方。
