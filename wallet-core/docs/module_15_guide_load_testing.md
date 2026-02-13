# Module 15: 压力测试实战 (Load Testing with k6)

本模块我们将使用 **k6** 对微服务架构进行压力测试。

## 1. 为什么需要压力测试?

在功能开发完成后，系统虽然能跑通，但能否抗住高并发流量?

- **瓶颈定位**: 数据库连接池够不够? Redis 缓存有没有生效? Go 协程会不会爆?
- **稳定性验证**: 在高负载下，服务会不会 OOM (内存溢出)?
- **容量规划**: 根据压测结果 (QPS)，预估服务器规格。

## 2. 工具选择: k6

相比 JMeter，k6 使用 **JavaScript** 编写脚本，更贴近开发者习惯，且性能极高 (Go 编写)。

### 2.1 安装 k6

**MacOS (Homebrew):**

```bash
brew install k6
```

**Verify:**

```bash
k6 version
# k6 v0.42.0 (or similar)
```

## 3. 编写压测脚本

脚本位置: `test/load/k6_script.js`

**场景设计 (User Journey):**

1.  **注册 (Register)**: `POST /v1/user/register`
    - 模拟新用户涌入。
    - 压力点: DB Insert, Password Hash (bcrypt)。
2.  **登录 (Login)**: `POST /v1/user/login`
    - 模拟用户鉴权。
    - 压力点: DB Query, Password Verify。
3.  **查余额 (Get Balance)**: `GET /v1/wallet/balance`
    - 模拟高频查询。
    - 压力点: gRPC 调用, Redis 读取 (如果有), DB 读取。

**配置 (Options):**

- `vus: 50`: 模拟 50 个并发虚拟用户。
- `duration: '30s'`: 持续压测 30 秒。
- `thresholds`: 定义成功标准 (如 95% 请求 < 500ms)。

## 4. 执行压测

### 4.1 启动全套服务

确保所有服务都在运行：

```bash
# 推荐使用 Docker Compose 启动依赖 (Postgres, Redis, Kafka)
docker-compose up -d

# 启动微服务 (建议在不同终端分别启动，观察日志)
go run cmd/user-service/main.go
go run cmd/wallet-service/main.go
go run cmd/bc-gateway/main.go
```

### 4.2 运行 k6

```bash
k6 run test/load/k6_script.js
```

## 5. 结果解读

执行完后，你会看到类似如下的报告：

```text
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  running (0m30.0s), 00/50 VUs, 2345 complete and 0 interrupted iterations
  default ✓ [======================================] 50 VUs  30s

     ✓ register status is 200
     ✓ has user_id
     ✓ login status is 200
     ✓ balance status is 200

     checks.........................: 100.00% ✓ 9380       ✗ 0
     http_req_duration..............: avg=35.42ms  min=2.12ms  med=28.45ms  max=342.12ms p(90)=89.12ms  p(95)=120.45ms
     http_reqs......................: 23450   781.666667/s
```

**关键指标:**

1.  **http_reqs (RPS/QPS)**: 每秒请求数 (例如 781.6/s)。数值越高，吞吐量越大。
2.  **http_req_duration (Latency)**:
    - `avg`: 平均延迟。
    - `p(95)`: 95% 的请求都在这个时间内完成。重点关注这个指标 (长尾效应)。
3.  **checks (Success Rate)**: 成功率。如果有 ✗，说明有请求失败 (如超时、500错误)。

## 6. 常见瓶颈与调优

如果 **p(95)** 很高 (>1s) 或 **RPS** 上不去:

1.  **数据库连接池满**:
    - 现象: DB 日志报错 `pool exhausted`。
    - 解决: 调大 `MaxOpenConns`。
2.  **锁竞争**:
    - 现象: `wallet-service` 的 `CreateWithdrawal` 很慢 (SELECT FOR UPDATE)。
    - 解决: 优化事务粒度，引入 Redis 缓存。
3.  **日志打印过多**:
    - 现象: 控制台疯狂刷日志，IO 飙升。
    - 解决: 压测时将日志级别调为 `WARN` 或关闭控制台输出。
