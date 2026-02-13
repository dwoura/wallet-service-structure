# 模块 6: Kubernetes (K8s) 部署实战指南 (进阶版)

> **核心目标**: 不仅仅是 "能跑"，而是 "生产就绪 (Production Ready)"。我们将探讨 K8s 的核心设计理念、网络模型以及企业级最佳实践。

## 1. 为什么我们需要 K8s? (Docker Compose 不够吗？)

`docker-compose` 适合单机编排，但在多机集群中，我们需要解决以下问题，而这正是 K8s 的强项：

- **调度 (Scheduling):** 哪台机器 CPU 空闲？把容器放到那台机器上去。
- **自愈 (Self-healing):** 机器 A 挂了，自动把上面的容器在机器 B 上重启。
- **弹性伸缩 (Auto-scaling):** 流量来了，自动从 2 个实例扩容到 10 个。
- **服务发现与负载均衡:** 它可以赋予一组 Pods 一个统一的 IP 和 DNS 名，并自动负载均衡。

---

## 2. 核心架构深度解析

### 2.1 Pod: 不仅仅是容器

- **概念:** Pod 是 K8s 的原子单位，它是一个 "逻辑主机"。
- **设计模式:** 为什么 Pod 可以包含多个容器？(Sidecar 模式)
  - 主容器 (`wallet-server`) 处理业务。
  - 辅助容器 (Sidecar) 可以做日志收集 (Filebeat)、流量代理 (Envoy/Istio)。它们共享同一个网络 (localhost) 和存储卷。

### 2.2 Deployment: 声明式管理的魅力

- **ReplicaSet:** Deployment 并不直接管理 Pod，它管理 ReplicaSet，ReplicaSet 再管理 Pod。
- **滚动更新 (Rolling Update):**
  - 当你更新镜像版本时，K8s 并不是一次性杀掉所有 Pod。
  - 它会先启动一个新的 Pod (v2)，等它 Ready 了，再杀掉一个旧的 Pod (v1)。
  - **零停机 (Zero Downtime)** 部署！

### 2.3 Service: 稳定的访问入口

Pod 的 IP 是临时的 (重启就变)，Service 提供了一个 **虚拟 IP (ClusterIP)**。

- **原理:** K8s 通过 `kube-proxy` (iptables/IPVS) 将发往 Service IP 的流量转发到对应的后端 Pods。
- **类型:**
  - **ClusterIP:** 默认，仅集群内部访问。
  - **NodePort:** 在每个节点上通过静态端口 (比如 30080) 暴露服务。
  - **admin:** 配合云厂商 (AWS/GCP) 的负载均衡器暴露服务。
  - **Ingress:** HTTP/HTTPS 层面的路由 (7层负载均衡)，通常用于暴露 Web 服务。

---

## 3. 企业级配置扩展点 (我们即将添加的)

普通的 `deployment.yaml` 可能只有几十行，但生产级配置需要考虑更多：

### 3.1 存活与就绪探针 (Probes) - **拒绝 "假死"**

- **Liveness Probe (存活探针):** "你还活着吗？"
  - 如果失败：K8s 会**重启**该容器。
  - 场景：死锁，进程卡死但 PID 还在。
- **Readiness Probe (就绪探针):** "你能接客了吗？"
  - 如果失败：K8s 会**停止发送流量**给该 Pod (从 Service 的 Endpoints 中摘除)。
  - 场景：程序启动需要加载 10秒的大数据，这期间不能接请求。

### 3.2 资源限制 (Resources) - **防止 "邻里干扰"**

- **Requests (请求量):** 调度依据。节点必须有这么多空闲资源，Pod 才能调度上去。
- **Limits (限制量):** 封顶值。
  - CPU 超限：被限流 (Throttled)。
  - Memory 超限：**OOM Kill** (被杀掉)。

### 3.3 优雅终止 (Graceful Shutdown)

- Pod 被删除时，K8s 发送 `SIGTERM` 信号。
- 应用应捕获该信号，处理完当前请求，断开数据库连接，然后再退出。

---

## 4. 实战：部署生产级全栈环境

### 步骤 A: 准备配置 (ConfigMap & Secret)

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
```

### 步骤 B: 部署有状态服务 (DB & Redis)

我们添加了 `volumeMounts` 来确保持久化数据。(在下面提供的 yaml 中已启用)

```bash
kubectl apply -f deploy/k8s/db-deployment.yaml
kubectl apply -f deploy/k8s/redis-deployment.yaml
```

### 步骤 C: 部署应用 (生产级配置)

我们更新后的 `deployment.yaml` 包含了以下**生产级特性**：

- **健康检查 (Probes)**: 使用 `/health` 接口替代 TCP检查。
- **优雅停机 (Graceful Shutdown)**: 设置 `terminationGracePeriodSeconds: 60`。
- **资源限制 (Resources)**: 防止 OOM。

同时，`configmap.yaml` 新增了 Viper 所需的环境变量：

- `APP_ENV`: "production"
- `REDIS_MQ_TYPE`: "redis"

```bash
kubectl apply -f deploy/k8s/deployment.yaml
kubectl apply -f deploy/k8s/service.yaml
```

### 步骤 D: 验证 "滚动更新"

1.  修改 `main.go` 里的版本号或日志，重新构建镜像：`docker build -t wallet-server:v2 .`
2.  更新 Deployment：`kubectl set image deployment/wallet-server wallet-server=wallet-server:v2`
3.  观察过程：`kubectl rollout status deployment/wallet-server`

---

## 5. Q&A 扩展

**Q: 既然 Pod 是临时的，数据库数据存哪？**
**A:** 使用 **PV (Persistent Volume)** 和 **PVC (Persistent Volume Claim)**。PVC 就像一张 "硬盘申请单"，K8s 会自动在通过 StorageClass 分配一块硬盘 (如 AWS EBS 或 本地磁盘) 挂载给 Pod。Pod 挂了，硬盘还在，新 Pod 挂载同一块硬盘即可。

**Q: 私钥怎么管理？**
**A:** `Secret` 只是 Base64 编码，不算加密。进阶方案使用 **HashiCorp Vault** 结合 Sidecar 动态注入 Agent 密钥，或者使用云厂商的 KMS 插件。

---

## 6. 实战复盘：刚才发生了什么？ (Troubleshooting Log)

在部署过程中，你可能注意到了几个关键的 "卡顿" 点，这正是 K8s 运维的高频考题：

### 1. "连接被拒绝" (Connection Refused)

- **现象**: `kubectl get nodes` 报错，连不上。
- **原因**: `kubectl` 只是一个客户端，它需要知道去哪里找服务器。你的电脑上可能配置了多个集群 (Context)，默认指向了一个不存在的 `local` 集群。
- **解决**: `kubectl config use-context docker-desktop`。
- **知识点**: **KubeConfig Context**。在管理多集群（比如 开发/测试/生产）时，切换 Context 是基本功。

### 2. "镜像找不到" (ImagePullBackOff / ErrImageNeverPull)

- **现象**: 我不得不运行 `docker tag ...` 命令。
- **原因**: Docker Compose 构建的镜像名叫 `wallet-core-wallet-server` (默认: 文件夹名\_服务名)，但我们的 K8s yaml 写的却是 `wallet-server:latest`。名字不匹配，K8s 就去 DockerHub 找，结果也没找到。
- **解决**: `docker tag` 给镜像起个别名，并在 yaml 里设置 `imagePullPolicy: IfNotPresent` (本地有就用本地的)。
- **知识点**: **镜像命名规范与拉取策略**。生产环境通常使用私有仓库 (Harbor/ACR)，地址会是 `my-registry.com/wallet-server:v1.0`。

### 3. "Pod 重启了几次" (CrashLoopBackOff / Restarts)

- **现象**: `wallet-server` 的状态从 `Running` -> `Error` -> `Running`，`RESTARTS` 计数增加。
- **原因**: **启动顺序**。K8s 是并发启动所有 Pod 的。`wallet-server` 启动太快，尝试连接 DB 时，DB 还没初始化好，导致连接失败程序退出。
- **K8s 的处理**: K8s 检测到 Pod 挂了，立刻**自动重启**它。第二次或第三次重启时，DB 已经好了，连接成功，服务稳定运行。
- **知识点**: **最终一致性 (Eventual Consistency) 与 自愈 (Self-healing)**。
  - 不要试图去 "编排启动顺序" (比如让 App 等 DB 10秒)。
  - 而是让 App 具备 **重试机制**，让 K8s 具备 **重启机制**。系统最终会收敛到稳定状态。这就是这是云原生架构的核心哲学。
