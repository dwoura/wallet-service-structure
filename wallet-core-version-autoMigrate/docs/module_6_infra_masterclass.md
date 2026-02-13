# Module 6: 基础设施与部署大师课 (Infrastructure & Deployment Masterclass)

> "代码写得再好，跑不起来也是白搭。" —— 只有掌握了部署，你才算真正拥有了全栈能力。

## 1. 为什么需要容器化 (Docker)?

**痛点**: "在我电脑上明明能跑，为什么到服务器上就报错了？"
这通常是因为环境不一致：

- OS 版本不同 (Mac vs Linux)
- 依赖库缺失 (glibc 版本问题)
- Go 版本不一致
- 端口冲突

**解法**: **Docker 镜像 (Image)**
把代码 + 运行环境 (OS/Libs) + 配置文件 打包在一起。

- **标准化**: 无论在哪里跑（AWS, 阿里云, 本地），环境都是一模一样的。
- **隔离性**: 应用之间互不干扰，不用担心端口冲突或依赖冲突。

### 核心概念

- **Dockerfile**: 镜像的"配方"。
- **Image (镜像)**: 按照配方烤出来的"蛋糕" (静态的，只读)。
- **Container (容器)**: 切下来正在吃的一块"蛋糕" (动态的，运行时的进程)。

## 2. 什么是编排 (Orchestration)?

如果你只有一个应用，用 `docker run` 就行了。但我们的钱包系统有：

- `wallet-server` (Go App)
- `Postgres` (数据库)
- `Redis` (缓存)
- `Kafka` (消息队列)
- `Zookeeper` (Kafka 依赖)

一个个手动启动太累了，而且还得管启动顺序（DB 要先于 Server 启动）。

**解法 1: Docker Compose (本地开发神器)**
用一个 `.yml` 文件定义所有服务，`docker-compose up` 一键启动全家桶。

**解法 2: Kubernetes / K8s (生产环境霸主)**
Docker Compose 只能单机跑。如果你的用户量暴增，需要 100 台服务器呢？
K8s 就像一个**"指挥官"**：

- **调度**: "这台机器空闲，把这个容器调度过去跑。"
- **自愈**: "这个容器挂了？马上重启一个新的。"
- **扩容**: "流量大了？自动从 3 个实例加到 10 个。"

## 3. 实战步骤 (Roadmap)

我们将按照企业级标准来走一遍流程：

1.  **编写 Dockerfile**:
    - 使用 **多阶段构建 (Multi-stage Build)**。
    - 第一阶段：用重型的 `golang:alpine` 镜像来编译代码 (Build)。
    - 第二阶段：把二进制文件复制到极小的 `alpine` 镜像中运行 (Run)。
    - _好处_: 镜像体积从 800MB -> 20MB！(面试加分项)

2.  **编写 docker-compose.yml**:
    - 把 `wallet-server` 加入到现有的 compose 文件中。
    - 配置环境变量 (`DB_HOST`, `REDIS_ADDR`) 使得容器内部能互相访问。

3.  **K8s 清单编写 (Deploy)**:
    - 模拟生产环境部署。
    - `ConfigMap`: 管理配置（不把配置写死在镜像里）。
    - `Secret`: 安全存密码。
    - `Deployment`: 定义如何运行应用。
    - `Service`: 定义外部如何访问应用。

准备好了吗？我们将从第一步 `Dockerfile` 开始！
