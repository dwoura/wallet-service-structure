# 监控与告警接入指南 (Monitoring Guide)

本指南介绍如何使用 Prometheus 和 Grafana 为 `wallet-core` 搭建可观测性系统，并包含常见问题的排查记录。

## 1. 架构概览

- **Metric Source**: `wallet-server` (暴露 `/metrics` 接口)
- **Collector**: Prometheus (Port 9090, 负责拉取数据)
- **Visualizer**: Grafana (Port 3000, 负责展示图表)

## 2. 快速启动

我们的监控栈已经集成在 `docker-compose.yml` 中。

```bash
# 启动所有服务 (包含监控)
docker-compose up -d

# 强行重建 (如果修改了代码没生效)
docker-compose up -d --build --force-recreate wallet-server
```

确认容器运行状态:

```bash
docker ps | grep wallet
# wallet-server      (0.0.0.0:8080->8080/tcp)
# wallet-prometheus  (0.0.0.0:9090->9090/tcp)
# wallet-grafana     (0.0.0.0:3000->3000/tcp)
```

## 3. 访问 Prometheus (Metrics Debug)

- **地址**: [http://localhost:9090](http://localhost:9090)
- **验证**:
  1. 点击顶部菜单 "Status" -> "Targets"。
  2. 确保 `wallet-server` 状态为 **UP**。
- **简单查询**:
  - 输入 `http_requests_total` 查看总请求数。
  - 输入 `rate(http_requests_total[1m])` 查看 QPS。

## 4. 访问 Grafana (Dashboard)

- **地址**: [http://localhost:3000](http://localhost:3000)
- **账号**: `admin` / `admin` (首次登录需重置)

### 4.1 配置数据源 (Data Source)

1.  打开 Configuration (齿轮图标) -> Data Sources。
2.  点击 **Add data source**。
3.  选择 **Prometheus**。
4.  在 URL 栏输入: `http://prometheus:9090` (注意: 必须用 Docker 内部服务名 `prometheus`，不能用 `localhost`)。
5.  点击底部的 **Save & Test**，应显示 "Data source is working"。

### 4.2 导入仪表盘 (Import Dashboard)

1.  点击左侧 "+" 号 -> **Import**。
2.  选择本项目中的 `deploy/grafana/dashboard.json`。
3.  选择对应的 Prometheus 数据源。
4.  点击 **Import**。

## 5. 关键指标说明

| 指标名称                        | 类型      | 说明                | Labels                                                         |
| :------------------------------ | :-------- | :------------------ | :------------------------------------------------------------- |
| `http_requests_total`           | Counter   | 累计请求总数        | `method` (GET/POST), `path` (/api/v1/ping), `status` (200/500) |
| `http_request_duration_seconds` | Histogram | 请求处理耗时分布    | `method`, `path`                                               |
| `go_goroutines`                 | Gauge     | 当前 Goroutine 数量 | -                                                              |
| `go_memstats_alloc_bytes`       | Gauge     | 当前内存使用量      | -                                                              |

---

## 6. 故障排查 (Troubleshooting) - 💡 详细记录

我们在搭建过程中遇到了以下典型问题，记录在此，供查阅。

### 问题 1: `/metrics` 接口返回 404 Not Found

**现象:**
访问 `curl localhost:8080/metrics` 返回 `404 page not found`。

**原因 (Root Cause):**
Docker 镜像缓存 (`Cache`) 导致新增加的 `/metrics` 路由代码没有真正被打进镜像里。虽然你改了代码，但 `docker-compose up -d` 可能复用了旧的层。

**解决方法:**
使用 `--no-cache` 和 `--force-recreate` 强制重构建：

```bash
docker-compose build --no-cache wallet-server
docker-compose up -d --force-recreate wallet-server
```

### 问题 2: Prometheus Target 显示 "Connection Refused" 或 "Down"

**现象:**
Prometheus 界面中 Target 显示红色 `DOWN`，错误信息 `dial tcp 127.0.0.1:8080: connect: connection refused`。

**原因:**
`prometheus.yml` 配置错误。

- **错误配置**: `targets: ['localhost:8080']`。Prometheus 容器里的 `localhost` 指的是容器自己，不是宿主机。
- **正确配置**: `targets: ['wallet-server:8080']`。必须使用 Docker Network 中的服务名称。

### 问题 3: 无法从宿主机访问 `wallet-server`

**现象:**
宿主机执行 `curl localhost:8080/api/v1/ping` 失败，但容器内正常。

**原因:**
`docker-compose.yml` 中忘记暴露端口。

**解决方法:**
在 `wallet-server` 服务下添加端口映射：

```yaml
ports:
  - "8080:8080"
  - "50051:50051"
```

### 问题 4: Grafana 添加数据源失败 "HTTP Error Bad Gateway"

**原因:**
与问题 2 类似，Grafana 容器无法解析 `localhost:9090`。

**解决方法:**
Grafana Data Source URL 必须填写 `http://prometheus:9090` (容器服务名)。
