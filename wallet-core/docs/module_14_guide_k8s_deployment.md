# Module 14: K8s 微服务部署实战指南 (Docker Compose vs Kubernetes)

本指南详细记录了如何将拆分后的微服务 (`user-service`, `wallet-service`, `broadcaster-worker`) 部署到 Kubernetes 集群中，并对比它是如何优于 Docker Compose 的。

## 1. 理论对比: Docker Compose vs Kubernetes

| 特性         | Docker Compose                           | Kubernetes (K8s)                                       | 结论                       |
| :----------- | :--------------------------------------- | :----------------------------------------------------- | :------------------------- |
| **定位**     | **单机**容器编排工具                     | **集群**容器编排与调度系统                             | 开发用 Compose，生产用 K8s |
| **扩展性**   | 只能在单台机器上运行，扩展受限于单机资源 | 可跨多台机器（节点）自动调度，支持水平自动扩缩容 (HPA) | K8s 胜出                   |
| **自愈能力** | 容器挂了可以重启 (`restart: always`)     | 强大的健康检查 (Probe)、自动重启、节点故障迁移         | K8s 胜出                   |
| **服务发现** | 基于 DNS (服务名即域名)                  | Service, Ingress, CoreDNS, 负载均衡                    | K8s 更灵活强大             |
| **配置管理** | 环境变量文件 (.env)                      | ConfigMap (明文配置), Secret (加密配置)                | K8s 更安全规范             |
| **滚动更新** | 支持简单的更新，但可能中断服务           | 支持零停机滚动更新 (Rolling Update)，可控制更新速率    | K8s 完胜                   |

**总结**:

- **Docker Compose** 适合 **本地开发** 和 **测试环境**，因为它简单、轻量、启动快。
- **Kubernetes** 是 **生产环境** 的事实标准，它解决了大规模分布式系统的管理、调度、高可用和网络问题。

---

## 2. 实战: 将微服务部署到 K8s (Kind/Minikube)

### 前置条件

- 已安装 Docker
- 已安装 kubectl
- 已启动本地 K8s 集群 (推荐使用 `kind` 或 `minikube`)

### 步骤 1: 构建统一镜像 (One Image, Multiple Entrypoints)

我们在 `Dockerfile` 中更新了构建逻辑，一次构建所有服务的二进制文件。

```bash
# 在项目根目录执行
docker build -t wallet-core:latest .
```

**技巧**: 如果使用 `kind`，需要将镜像加载到集群节点中，否则 K8s 会尝试去 Docker Hub 拉取。

```bash
kind load docker-image wallet-core:latest
```

### 步骤 2: 准备配置 (ConfigMap & Secret)

首先应用配置，确保 Pod 启动时能读取到环境变量。

```bash
kubectl apply -f deploy/k8s/configmap.yaml
kubectl apply -f deploy/k8s/secret.yaml
```

### 步骤 3: 部署基础设施 (DB & Redis)

如果你的 K8s 集群内没有外部数据库，可以使用我们在 K8s 内部署的测试实例。

```bash
kubectl apply -f deploy/k8s/db-deployment.yaml
kubectl apply -f deploy/k8s/db-service.yaml # 确保有 Service
kubectl apply -f deploy/k8s/redis-deployment.yaml
kubectl apply -f deploy/k8s/redis-service.yaml
```

### 步骤 4: 部署微服务

应用我们生成的 Deployment 清单。

```bash
# 1. 用户服务 (User Service)
kubectl apply -f deploy/k8s/user-deployment.yaml

# 2. 钱包服务 (Wallet Service)
kubectl apply -f deploy/k8s/wallet-deployment.yaml

# 3. 广播 Worker (Broadcaster)
kubectl apply -f deploy/k8s/broadcaster-deployment.yaml
```

### 步骤 5: 验证部署

查看 Pod 状态：

```bash
kubectl get pods
```

预期输出：

```text
NAME                                  READY   STATUS    RESTARTS   AGE
broadcaster-worker-zj8kd              1/1     Running   0          10s
user-service-ak9s2                    1/1     Running   0          10s
wallet-service-jd8s1                  1/1     Running   0          10s
postgres-0                            1/1     Running   0          5m
redis-master-0                        1/1     Running   0          5m
```

查看日志 (以用户服务为例)：

```bash
kubectl logs -l app=user-service
```

应看到 `User Service listening on gRPC port :50053`。

### 步骤 6: 服务间调用验证

在 K8s 内部，服务通过 DNS 名称相互访问。

- `user-service` 的地址是: `user-service.default.svc.cluster.local:50053`
- `wallet-service` 的地址是: `wallet-service.default.svc.cluster.local:50052`

你可以启动一个临时 Pod 来测试连接：

```bash
kubectl run -it --rm debug --image=curlimages/curl -- sh
# Inside Pod:
# telnet user-service 50053
```

---

## 3. 下一步: API 网关 (Ingress)

目前我们的服务只暴露在集群内部 (ClusterIP)。要让外部访问，我们需要部署 API 网关或 Ingress Controller。

在 Module 14 的后续阶段，我们将构建 `bc-gateway` (HTTP Server)，它也部署在 K8s 中，通过 gRPC 调用后端的 `user-service` 和 `wallet-service`，并对外暴露 HTTP 接口。
