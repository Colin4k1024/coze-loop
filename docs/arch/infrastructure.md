# Infrastructure Architecture

本文档描述 coze-loop 项目的技术架构和基础设施组件。

## 1. 系统概述

coze-loop 是一个基于微服务架构的 AI 应用平台，后端采用 Go 语言开发，使用 GORM 作为 ORM 框架。主要基础设施组件包括：

- **MySQL** - 关系型数据库，存储业务数据
- **Redis** - 缓存、分布式锁、限流
- **ClickHouse** - OLAP 数据库，存储可观测性数据（trace、metrics）
- **RocketMQ** - 消息队列，用于异步任务和事件驱动
- **MinIO/S3** - 对象存储，存储文件和数据
- **Python/JS FaaS** - 函数即服务，执行用户自定义代码

## 2. 数据库 (MySQL)

### 2.1 ORM 配置

**位置**: `backend/infra/db/`

```go
// 使用 GORM 作为 ORM 框架
import "gorm.io/gorm"
import "gorm.io/driver/mysql"
```

**配置结构** (`backend/infra/db/config.go`):

```go
type Config struct {
    Timeout           time.Duration  // 连接超时
    ReadTimeout       time.Duration  // 读超时
    WriteTimeout      time.Duration  // 写超时
    User              string         // 数据库用户
    Password          string         // 数据库密码
    Loc               string         // 时区
    DBName            string         // 数据库名
    DBCharset         string         // 字符集，默认 utf8mb4
    DBHostname        string         // 主机地址
    DBPort            string         // 端口
    InterpolateParams bool           // 是否启用参数插值
    DSNParams         url.Values     // 额外 DSN 参数
    WithReturning     bool           // 是否支持 RETURNING 子句
}
```

**Provider 接口**:

```go
type Provider interface {
    NewSession(ctx context.Context, opts ...Option) *gorm.DB
    Transaction(ctx context.Context, fc func(tx *gorm.DB) error, opts ...Option) error
}
```

### 2.2 连接池配置

- 使用 GORM 默认连接池
- 支持读写分离配置（通过 `dbresolver` 插件）
- 支持强制读主库选项 `WithMaster()`
- 支持事务嵌套 `WithTransaction(tx)`

### 2.3 数据库初始化

数据库初始化通过 `mysql-init` 容器完成：

**位置**: `release/deployment/docker-compose/bootstrap/mysql-init/`

初始化脚本执行顺序：
1. 等待 MySQL 服务就绪
2. 执行 `init-sql/` 目录下的建表 SQL 文件
3. 执行 `patch-sql/alter_proc.sql` 存储过程
4. 执行 `patch-sql/*_alter.sql` 变更脚本

**建表 SQL 文件示例**:
- `dataset.sql` - 数据集表
- `dataset_item.sql` - 数据集条目表
- `evaluator.sql` - 评估器表
- `evaluator_template.sql` - 评估器模板表
- `prompt.sql` - 提示词表
- `api_key.sql` - API 密钥表

## 3. Redis

### 3.1 使用模式

**位置**: `backend/infra/redis/`

Redis 在系统中承担三种主要职责：

1. **缓存** - 字符串、哈希表操作
2. **分布式锁** - 基于 SETNX 的互斥锁
3. **限流** - 计数器+滑动窗口

### 3.2 客户端封装

```go
// 基于 github.com/redis/go-redis/v9
type Cmdable interface {
    SimpleCmdable      // String/Hash/ZSet/List 操作
    Pipeline() Pipeliner  // 管道操作
}

type SimpleCmdable interface {
    StringCmdable    // Get/Set/Incr/Decr/MGet/MSet
    HashCmdable      // HSet/HGet/HGetAll/HIncrBy
    SortedSetCmdable // ZAdd/ZRange
    ListCmdable      // RPush/LRange/LTrim
    Del/Eval/Expire/Exists
}
```

### 3.3 分布式锁实现

**位置**: `backend/infra/lock/lock.go`

```go
type ILocker interface {
    WithHolder(holder string) ILocker
    Lock(ctx context.Context, key string, expiresIn time.Duration) (bool, error)
    Unlock(key string) (bool, error)
    LockBackoff(ctx context.Context, key string, expiresIn time.Duration, maxWait time.Duration) (bool, error)
    LockBackoffWithRenew(parent context.Context, key string, ttl time.Duration, maxHold time.Duration) (locked bool, ctx context.Context, cancel func(), err error)
    LockWithRenew(parent context.Context, key string, ttl time.Duration, maxHold time.Duration) (locked bool, ctx context.Context, cancel func(), err error)
    BackoffLockWithValue(ctx context.Context, key, val string, expiresIn time.Duration, backoff time.Duration) (bool, string, error)
    UnlockForce(ctx context.Context, key string) (bool, error)
}
```

**锁特性**:
- Lua 脚本保证 Unlock 原子性
- 支持自动续期（Renew）
- 支持指数退避重试
- 锁持有者标识防止误删

### 3.4 限流器

**位置**: `backend/infra/limiter/`

```go
type IRateLimiter interface {
    AllowN(ctx context.Context, key string, n int, opts ...LimitOptionFn) (*Result, error)
}

type Rule struct {
    Match   string  // 标签匹配表达式
    KeyExpr string  // 限流 key 表达式
    Limit   Limit   // 限流配置
}

type Limit struct {
    Rate   int           // 速率（次/秒）
    Burst  int           // 突发容量
    Period time.Duration // 时间周期
}
```

## 4. 消息队列 (RocketMQ)

### 4.1 架构概览

**位置**: `backend/infra/mq/`

使用 Apache RocketMQ 5.x 作为消息中间件，采用 Push Consumer 模式消费消息。

### 4.2 组件结构

```
mq/
├── factory.go          # 生产者/消费者工厂
├── consumer.go        # 消费者接口
├── producer.go        # 生产者接口
├── message.go         # 消息结构
├── registry/
│   ├── registry.go    # 消费者注册表
│   └── registry_test.go
└── rocketmq/
    ├── factory.go     # RocketMQ 工厂实现
    ├── consumer.go    # RocketMQ 消费者实现
    └── producer.go    # RocketMQ 生产者实现
```

### 4.3 消费者注册机制

```go
type ConsumerRegistry interface {
    Register(worker []IConsumerWorker) ConsumerRegistry
    StartAll(ctx context.Context) error
    StopAll(ctx context.Context) error
}
```

**特性**:
- 支持优雅关闭（通过 shutdownCtx）
- 消费者处理包装（SafeConsumerWrapper）实现 panic 恢复
- 支持并发消费配置

### 4.4 RocketMQ 主题

根据 `observability.yaml` 配置，系统使用以下主题：

| 主题名 | 用途 | 消费者组 |
|--------|------|----------|
| `trace_ingestion_event` | Trace 数据摄入 | trace_ingestion_event_cg |
| `trace_annotation_event` | Trace 标注事件 | trace_annotation_event_cg |
| `observability_span_queue` | 可观测性 Span 队列 | observability_span_queue_cg |
| `trace_to_task` | Trace 到任务转换 | trace_to_task_cg |
| `cozeloop_async_tasks` | 异步任务 | cozeloop_async_tasks_backfill_cg |
| `cozeloop_evaluation_online_expt_eval_result` | 评估实验结果 | cozeloop_evaluation_online_expt_eval_result_cg |
| `cozeloop_evaluation_correction_evaluator_result` | 评估修正结果 | cozeloop_evaluation_correction_evaluator_result_evaluation_cg |

### 4.5 环境变量

RocketMQ NameServer 配置通过环境变量获取：

```go
func getRmqNamesrvDomain() string {
    return os.Getenv("COZE_LOOP_RMQ_NAMESRV_DOMAIN")
}
func getRmqNamesrvPort() string {
    return os.Getenv("COZE_LOOP_RMQ_NAMESRV_PORT")
}
func getRmqNamesrvUser() string {
    return os.Getenv("COZE_LOOP_RMQ_NAMESRV_USER")
}
func getRmqNamesrvPassword() string {
    return os.Getenv("COZE_LOOP_RMQ_NAMESRV_PASSWORD")
}
```

## 5. ClickHouse

### 5.1 配置

**位置**: `backend/infra/ck/`

```go
type Config struct {
    Host              string            // ClickHouse 地址
    Database          string            // 数据库名
    Username          string            // 用户名
    Password          string            // 密码
    DialTimeout       time.Duration     // 拨号超时
    ReadTimeout       time.Duration     // 读超时
    CompressionMethod CompressionMethod // 压缩方法：LZ4/ZSTD/GZIP/Deflate/Brotli
    CompressionLevel  int               // 压缩级别
    Protocol          Protocol          // 协议：HTTP/Native
    Debug             bool              // 调试模式
    HttpHeaders       map[string]string // HTTP 头
    Settings          map[string]any    // ClickHouse 设置
}
```

### 5.2 数据用途

ClickHouse 主要存储可观测性数据：

- **Span 数据** - 分布式追踪的 span 记录
- **Annotation 数据** - 标注信息
- **Metric 数据** - 指标数据

### 5.3 初始化

**位置**: `release/deployment/docker-compose/bootstrap/clickhouse-init/`

初始化 SQL 文件：
- `observability_spans.sql` - Span 表
- `observability_annotations.sql` - 标注表
- `evaluation.sql` - 评估数据表

## 6. 文件存储 (MinIO/S3)

### 6.1 架构

**位置**: `backend/infra/fileserver/`

```go
type S3Client struct {
    s3  *s3.S3
    cfg *S3Config
}

var (
    _ ObjectStorage      = (*S3Client)(nil)
    _ BatchObjectStorage = (*S3Client)(nil)
)
```

### 6.2 功能接口

```go
// 单对象操作
Stat(ctx context.Context, key string, opts ...StatOpt) (*ObjectInfo, error)
Upload(ctx context.Context, key string, r io.Reader, opts ...UploadOpt) error
Download(ctx context.Context, key string, writer io.WriterAt, opts ...DownloadOpt) error
Read(ctx context.Context, key string, opts ...DownloadOpt) (Reader, error)
Remove(ctx context.Context, key string, opts ...RemoveOpt) error

// 签名 URL
SignDownloadReq(ctx context.Context, key string, opts ...SignOpt) (url string, header http.Header, err error)
SignUploadReq(ctx context.Context, key string, opts ...SignOpt) (url string, header http.Header, err error)

// 批量操作
BatchUpload(ctx context.Context, keys []string, readers []io.Reader, opts ...UploadOpt) error
BatchRead(ctx context.Context, keys []string, opts ...DownloadOpt) ([]Reader, error)
BatchDownload(ctx context.Context, keys []string, writers []io.WriterAt, opts ...DownloadOpt) error
BatchSignDownloadReq(ctx context.Context, keys []string, opts ...SignOpt) ([]string, []http.Header, error)
BatchSignUploadReq(ctx context.Context, keys []string, opts ...SignOpt) ([]string, []http.Header, error)
```

### 6.3 分片上传/下载

- 上传分片大小：`UploadPartSize` (默认配置)
- 下载分片大小：`DownloadPartSize` (默认配置)
- 小文件（<= CacheSizeGT）直接读入内存
- 大文件使用临时文件+分片下载

### 6.4 S3 兼容配置

支持 MinIO 和其他 S3 兼容存储：
- 可配置 Endpoint
- 支持 PathStyle 和 VirtualHost 模式
- 使用 Static Credentials 认证

## 7. 可观测性

### 7.1 链路追踪 (LoopTracer)

**位置**: `backend/infra/looptracer/`

```go
type TracerImpl struct {
    cozeloop.Client
}

type Tracer interface {
    StartSpan(ctx context.Context, name, spanType string, opts ...StartSpanOption) (context.Context, Span)
    GetSpanFromContext(ctx context.Context) Span
    Inject(ctx context.Context) context.Context
    Flush(ctx context.Context)
    InjectW3CTraceContext(ctx context.Context) map[string]string
}
```

**Span 接口**:
```go
type Span interface {
    GetSpanID() string
    GetTraceID() string
    SetInput/SetOutput/SetError/SetStatusCode
    SetUserID/SetMessageID/SetThreadID
    SetPrompt/SetModelProvider/SetModelName
    SetInputTokens/SetOutputTokens
    Finish(ctx context.Context)
}
```

**Trace Context Header**:
- `X-Cozeloop-Traceparent` - W3C TraceParent
- `X-Cozeloop-Tracestate` - W3C TraceState
- `traceparent` / `tracestate` - W3C 标准格式

### 7.2 指标 (Metrics)

**位置**: `backend/infra/metrics/`

```go
type Metric interface {
    Emit(tags []T, values ...*Value)
}

type Value struct {
    suffix string
    mType  MetricType  // Counter/RateCounter/Store/Timer/Histogram
    value  *int64
    valuef *float64
}
```

**指标类型**:
- `Counter` - 累计计数器
- `RateCounter` - 速率计数器
- `Store` - 瞬时值
- `Timer` - 计时器
- `Histogram` - 直方图

### 7.3 日志

**位置**: `backend/infra/middleware/logs/`

结构化日志，支持上下文关联。

### 7.4 可观测性配置

**位置**: `release/deployment/docker-compose/conf/observability.yaml`

关键配置：
- Trace 摄入配置（MQ Producer/Consumer）
- Annotation 事件主题
- Span 队列配置
- 消费者 Worker 数量
- 租户隔离配置

## 8. IDL/Thrift

### 8.1 Thrift 定义结构

**位置**: `idl/thrift/`

```
idl/thrift/
├── base.thrift           # 基础结构（Base/BaseResp/TrafficEnv）
└── coze/loop/
    ├── apis/            # API 定义
    ├── data/            # 数据模型
    │   └── domain/      # 领域模型
    ├── evaluation/      # 评估相关
    │   └── domain/      # 评估领域模型
    ├── foundation/      # 基础服务（auth/user/space）
    ├── llm/             # LLM 相关
    ├── observability/   # 可观测性（trace/metric）
    └── prompt/          # 提示词管理
```

### 8.2 主要服务定义

| 服务模块 | 用途 |
|---------|------|
| `coze.loop.apis` | 主 API 接口 |
| `coze.loop.data.dataset` | 数据集管理 |
| `coze.loop.evaluation` | 评估实验 |
| `coze.loop.evaluation.evaluator` | 评估器 |
| `coze.loop.prompt` | 提示词管理 |
| `coze.loop.llm` | LLM 运行时 |
| `coze.loop.observability` | 可观测性接口 |
| `coze.loop.foundation` | 基础服务（认证/用户/空间）|

### 8.3 代码生成

Thrift 文件通过代码生成工具生成对应语言的 SDK 代码。具体生成命令和工具链需参考项目构建配置。

## 9. 配置管理

### 9.1 配置目录结构

```
conf/
├── infrastructure.yaml  # 基础设施配置
├── observability.yaml   # 可观测性配置
├── model_config.yaml    # 模型配置
└── locales/            # 国际化文件
    ├── en-US.yaml
    └── zh-CN.yaml
```

### 9.2 基础设施配置 (infrastructure.yaml)

```yaml
infra:
  redis:
    host: "cozeloop-redis"
    port: 6379
    password: "cozeloop-redis"
  rds:
    host: "cozeloop-mysql"
    port: 3306
    user: "root"
    password: "cozeloop-mysql"
    db: "cozeloop-mysql"
  s3_config:
    region: 'us-east-1'
    endpoint: 'http://cozeloop-minio:19000'
    bucket: 'cozeloop-minio'
    access_key: 'root'
    secret_access_key: 'cozeloop-minio'
  ck_config:
    host: "cozeloop-clickhouse:9008"
    username: "default"
    password: "cozeloop-clickhouse"
    database: "cozeloop-clickhouse"
  idgen:
    server_ids:
      - 1
  log_level: "info"
```

### 9.3 环境变量映射

应用程序通过环境变量接收运行时配置：

| 环境变量 | 用途 |
|---------|------|
| `COZE_LOOP_REDIS_DOMAIN/PORT/PASSWORD` | Redis 连接 |
| `COZE_LOOP_MYSQL_DOMAIN/PORT/USER/PASSWORD/DATABASE` | MySQL 连接 |
| `COZE_LOOP_CLICKHOUSE_DOMAIN/PORT/USER/PASSWORD/DATABASE` | ClickHouse 连接 |
| `COZE_LOOP_OSS_PROTOCOL/DOMAIN/PORT/USER/PASSWORD/BUCKET` | 对象存储 |
| `COZE_LOOP_RMQ_NAMESRV_DOMAIN/PORT` | RocketMQ NameServer |
| `COZE_LOOP_PYTHON_FAAS_DOMAIN/PORT` | Python FaaS |
| `COZE_LOOP_JS_FAAS_DOMAIN/PORT` | JS FaaS |

## 10. 中间件

**位置**: `backend/infra/middleware/`

### 10.1 组件列表

| 目录 | 功能 |
|------|------|
| `ctxcache` | 上下文缓存 |
| `errors` | 错误处理 |
| `localrpc` | 本地 RPC 调用 |
| `logs` | 日志中间件 |
| `session` | 会话管理 |
| `validator` | 参数校验 |

### 10.2 会话管理

```go
type SessionMiddleware struct {
    // 会话上下文处理
}
type UserSession struct {
    // 用户会话信息
}
```

## 11. 技术栈总结

| 组件 | 技术选型 | 版本示例 |
|------|---------|---------|
| 语言 | Go | 1.21+ |
| ORM | GORM | v1.x |
| 数据库 | MySQL | 9.4.0 |
| 缓存/锁 | Redis | 8.2.0 |
| OLAP | ClickHouse | latest |
| 消息队列 | RocketMQ | 5.3.3 |
| 对象存储 | MinIO | RELEASE.2025-06-13 |
| IDL | Apache Thrift | - |
| Web 框架 | Gin (推测) | - |
| 链路追踪 | cozeloop-go | - |
