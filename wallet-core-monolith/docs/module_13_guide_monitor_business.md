# Go 后端开发指南：业务指标监控 (Business Metrics) 实战手册

> **"If you can't measure it, you can't improve it."** — Peter Drucker

对于区块链钱包这类资金敏感系统，监控不仅仅是看服务器 CPU 爆没爆，更重要的是看 **业务是否健康**。
本指南将带你从零开始，理解“埋点”概念，并手把手教你如何用 Prometheus 监控核心业务。

---

## 1. 什么是“埋点”？ (What is Instrumentation?)

**埋点**（Instrumentation）就是在代码的关键位置“插眼”。

想象你在经营一家银行：

- **日志 (Log)** 是录像带：出事了回去翻，看细节。
- **监控 (Metric)** 是仪表盘：实时显示当前有多少人排队，金库还剩多少钱。

**主要区别**：

- 日志：`User A deposited 100 BTC` (文本，体积大，难统计)
- 监控：`deposit_count += 1`, `deposit_amount += 100` (数字，体积小，查询快)

---

## 2. Prometheus 三大核心神器 (Metric Types)

Prometheus 提供了几种数据类型，针对不同的场景：

### 2.1 Counter (只增不减的计数器)

**场景**：累计发生了多少次？

- **例子**：用户注册人数、充值笔数、错误发生次数。
- **特点**：只能 `Inc()` (加1) 或 `Add(x)`。重启后归零（Prometheus 会处理这个问题，画图时用 `rate()` 函数看增长率）。

### 2.2 Gauge (可增可减的仪表盘)

**场景**：现在的状态是多少？

- **例子**：地址池剩余数量、当前由 Goroutine 数量、内存占用。
- **特点**：可以 `Inc()`, `Dec()`, `Set(x)`。

### 2.3 Histogram (直方图/柱状图)

**场景**：耗时分布、大小分布。

- **例子**：API 请求耗时（P99是多少？）、归集任务耗时。
- **特点**：它会把数据分到不同的“桶” (Bucket) 里，比如 `0.1s`, `0.5s`, `1s`。

---

## 3. 实战：如何添加一个新指标？

假设我们要监控 **"用户修改密码的次数"**。

### 第一步：定义指标 (pkg/monitor/business.go)

在 `BusinessMetrics` 结构体中加一个字段，并在 `Init` 中初始化。

```go
package monitor

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type BusinessMetrics struct {
    // ... 原有指标 ...

    // [NEW] 定义一个 Counter
    UserPasswordChangedTotal prometheus.Counter
}

func InitBusinessMetrics() {
    Business = &BusinessMetrics{
        // ... 原有初始化 ...

        // [NEW] 初始化
        UserPasswordChangedTotal: promauto.NewCounter(prometheus.CounterOpts{
            Name: "wallet_user_password_changed_total", // 这里的名字要符合 Prometheus 规范 (下划线命名)
            Help: "The total number of password changes",
        }),
    }
}
```

### 第二步：埋点 (internal/handler/user.go)

找到修改密码的业务代码，在成功的位置“打点”。

```go
func ChangePassword(c *gin.Context) {
    // 1. 校验旧密码...
    // 2. 更新数据库...

    if err == nil {
        // [NEW] 埋点：计数器 +1
        monitor.Business.UserPasswordChangedTotal.Inc()
    }

    response.Success(c, nil)
}
```

### 第三步：验证数据

启动服务，访问 `http://localhost:8080/metrics`。
你应该能搜索到 `wallet_user_password_changed_total`。

- 初始值可能是 0 (或者没显示，直到第一次触发)。
- 触发一次修改密码接口，再刷新页面，值应该变成 1。

---

## 4. 高级技巧：使用标签 (Labels)

如果你想区分 **BTC 充值** 和 **ETH 充值**，不需要定义两个变量，用 **Label** 即可。

### 定义带 Label 的指标 (`*Vec`)

```go
// CounterVec = Vector (向量)
DepositAmountTotal *prometheus.CounterVec
```

初始化时指明 Label 的 key：

```go
DepositAmountTotal: promauto.NewCounterVec(prometheus.CounterOpts{
    Name: "wallet_deposit_amount_total",
    Help: "Total deposit amount",
}, []string{"currency"}) // <--- Label Key
```

### 使用时填入 Label Value

```go
// 记录 BTC 充值
monitor.Business.DepositAmountTotal.WithLabelValues("BTC").Add(0.5)

// 记录 ETH 充值
monitor.Business.DepositAmountTotal.WithLabelValues("ETH").Add(10.0)
```

**Prometheus 里的样子**：

```text
wallet_deposit_amount_total{currency="BTC"} 0.5
wallet_deposit_amount_total{currency="ETH"} 10.0
```

---

## 5. Grafana 可视化配置

有了数据，怎么画出高大上的 Dashboard？

### 5.1 设置数据源

确保 Grafana 已连接到 Prometheus。

### 5.2 常用查询语句 (PromQL)

**A. 查看最近 5 分钟的注册速率 (每秒注册多少人)**

```promql
rate(wallet_user_registered_total[5m])
```

**B. 查看今日充值总额 (按币种分组)**
Counter 是累加的，所以直接查就是总额。如果中间重启过，Prometheus 会处理，但通常用 `increase()` 看增量更直观。

```promql
sum by (currency) (wallet_deposit_amount_total)
```

**C. 归集任务的 P99 耗时 (99% 的任务都快于这个时间)**

```promql
histogram_quantile(0.99, sum(rate(wallet_sweeper_job_duration_seconds_bucket[5m])) by (le))
```

---

## 6. 避坑指南

1.  **Label 不要太多**: 不要把 `UserID` 或 `TxHash` 作为 Label！
    - ❌ `WithLabelValues(tx_hash)` -> 导致指标数量爆炸 (Cardinality Explosion)，内存撑爆。
    - ✅ Label 应该是有限集合 (如 币种、错误码、主机名)。
2.  **金额精度**: Prometheus 的值默认是 `float64`。对于超大金额 (如 ETH 精度 18 位)，累加可能会有微小精度丢失。对于监控大盘通常可以接受，但不要用于财务对账。
3.  **命名规范**: 乃至 `system_metric_name`。通常以 `_total`, `_seconds`, `_count` 结尾。
