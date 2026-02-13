# 模块 6: CI/CD 流水线与自动化 (GitHub Actions)

> **目标**: 并不是 "写好代码" 就结束了。我们需要确保这些代码在别人的环境里也能跑，而且不会把之前的代码搞挂。

## 1. 什么是 CI/CD？

- **CI (Continuous Integration - 持续集成):**
  - **核心动作**: 每次你 Push 代码，我就自动运行测试。
  - **目的**: 尽早发现 Bug。防止 "在我这就好好的" 这种问题。
- **CD (Continuous Delivery - 持续交付):**
  - **核心动作**: 测试通过后，自动构建镜像、自动部署到环境。
  - **目的**: 快速发布，减少手动操作的风险。

---

## 2. 我们的流水线配置 (`.github/workflows/ci.yml`)

我为你创建的那个 YAML 文件，就是给 GitHub 的 "说明书"。让我们逐行拆解：

```yaml
name: Wallet Core CI # 流水线的名字

# [触发器]: 什么时候运行？
on:
  push:
    branches: ["main", "master"] # 当 push 到 main 分支时
  pull_request: # 当有人提 PR 时 (Code Review)
    branches: ["main", "master"]

jobs:
  # === JOB 1: 测试 (Test) ===
  test:
    name: Test & Lint
    runs-on: ubuntu-latest # 在 GitHub 提供的 Ubuntu 虚拟机上跑
    steps:
      # 1. 把代码拉下来
      - name: Checkout code
        uses: actions/checkout@v4

      # 2. 装好 Go 1.24 环境
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.24"

      # 3. 检查 go.mod 和 go.sum 是否完整
      - name: Verify dependencies
        run: go mod verify

      # 4. 真的跑测试！(-v 显示详细日志)
      - name: Run Unit Tests
        run: go test -v ./...

  # === JOB 2: 构建 (Build) ===
  build:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: test # [关键]: 只有上面的 'test' 任务成功了，才跑这个！
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      # 模拟构建 Docker 镜像，确保 Dockerfile 写得没问题
      - name: Build Docker image
        run: docker build . --file Dockerfile --tag wallet-server:latest
```

---

```

---

## 3. 核心概念深度解析 (Critical Concepts)

为了让你也能自己写出这样的文件，我们需要理解以下几个关键点：

### 3.1 缩进 (Indentation)
**YAML 的灵魂是缩进。**
*   它不像 Go 用 `{}` 包裹代码块，而是用**空格**。
*   通常使用 **2 个空格**。
*   **严禁使用 Tab 键！** (这是 YAML 最大的坑)。
*   如果你发现 CI 报错 `syntax error`，90% 是因为缩进不对。

### 3.2 插件机制 (`uses` vs `run`)
*   **`run: ...`**: 这就是最普通的命令行。你想想自己在终端里怎么敲命令 (比如 `go test`, `docker build`)，这里就怎么写。
*   **`uses: ...`**: 这是 GitHub Actions 的杀手锏 —— **复用别人的轮子** (Actions Marketplace)。
    *   比如 `actions/checkout@v4`: 别人写好的复杂脚本，用来安全地拉取代码。
    *   比如 `actions/setup-go@v5`: 别人写好的脚本，用来下载、安装、配置 Go 环境。
    *   **原则**: 能用现成的 `uses`，就不要自己写复杂的 `run` 脚本。

### 3.3 任务依赖 (`needs`)
*   默认情况下，`jobs` 里的任务是**并行运行**的 (Parallel)。
    *   如果去掉 `needs: test`，那么 `test` 和 `build` 会同时开始跑。
    *   但这不合理！如果测试都挂了，构建镜像纯属浪费时间。
*   **`needs: [test]`**: 强制要求：必须等 `test` 任务变绿 (Success) 了，我 `build` 任务才开始跑。这叫**串行运行** (Sequential)。

---

## 4. 为什么这一步很重要？(Web2 vs Web3)

在 Web3 领域，尤其是涉及资金的钱包项目，**安全性 > 一切**。

- **没有 CI 的裸奔**:
  - 你改了一行代码，觉得没问题，直接发版了。
  - 结果这行代码导致 `BIP39` 生成助记词的逻辑变了，所有新用户的私钥都错了。
  - 后果：资金永久丢失。

- **有 CI 的流程**:
  - 你改了代码，Push 上去。
  - GitHub Actions 自动运行 `pkg/wallet/bip39/mnemonic_test.go`。
  - 测试失败！❌ 红色警告。
  - 你甚至无法合并代码到 `main` 分支。
  - **风险被拦截在上线之前**。

---

## 5. 下一步如何验证？

虽然我刚才帮你生成了文件，但 **CI 是要在 GitHub 服务器上跑的**。

你需要做的是：

1.  **Commit & Push**: 把代码推送到你的 GitHub 仓库。
2.  **观察**: 打开仓库页面，点击顶部的 `Actions` 标签。
3.  **结果**: 你会看到一个黄色的圆圈正在转（Running），几分钟后变成绿色的勾（Success）。

这就是现代软件工程的魅力：**把重复的工作交给机器人，人只负责创造。** 🤖
```
