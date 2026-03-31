# Backend 架构文档

## 1. 概述

coze-loop 后端是一个基于 Go 语言构建的 AI 应用后端服务，采用微服务架构风格，集成了多种中间件和云服务。该项目主要用于支持 AI 模型的训练、评估、Prompt 管理和可观测性等功能。

### 1.1 技术栈概览

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| **Web 框架** | Hertz (by ByteDance) | 高性能 HTTP 框架 |
| **RPC 框架** | Kitex (by ByteDance) | 支持 Thrift/Protobuf |
| **数据库** | MySQL + ClickHouse | MySQL 作主存储，ClickHouse 作分析 |
| **缓存** | Redis | 会话、限流、缓存 |
| **消息队列** | RocketMQ | 异步任务处理 |
| **对象存储** | S3 兼容存储 | 文件、日志存储 |
| **依赖注入** | Google Wire | 编译期 DI |
| **LLM 集成** | Eino (by ByteDance) | 多模型统一抽象 |
| **IDL** | Thrift | 服务定义与代码生成 |

---

## 2. 项目结构

```
backend/
├── api/                    # HTTP API 层
│   ├── handler/            # 请求处理器 (Hertz handlers)
│   │   └── coze/loop/apis/
│   ├── router/            # 路由注册
│   │   └── coze/loop/apis/
│   ├── router_gen.go      # 生成的路由注册入口
│   └── api.go             # API 初始化逻辑
├── cmd/                   # 入口点
│   ├── main.go            # HTTP 服务 + Consumer 启动
│   └── consumer.go        # Consumer Worker 初始化
├── conf/                  # 配置文件
│   └── infrastructure.yaml
├── infra/                 # 基础设施层
│   ├── db/                # MySQL 数据库
│   ├── redis/             # Redis 客户端
│   ├── mq/                # RocketMQ 消息队列
│   ├── ck/                # ClickHouse
│   ├── fileserver/        # S3 对象存储
│   ├── middleware/         # HTTP 中间件
│   ├── metrics/           # 指标采集
│   ├── looptracer/       # 链路追踪
│   ├── limiter/           # 限流
│   ├── lock/              # 分布式锁
│   ├── i18n/              # 国际化
│   └── idgen/             # ID 生成器
├── kitex_gen/             # Kitex 生成的代码 (Thrift IDL)
├── loop_gen/              # 本地服务桩代码
├── modules/               # 业务模块
│   ├── data/              # 数据集管理
│   ├── evaluation/        # 评估实验
│   ├── foundation/         # 基础服务 (Auth/User/Space/File)
│   ├── llm/               # LLM 模型管理
│   ├── observability/     # 可观测性 (Trace/Metric/Task)
│   └── prompt/            # Prompt 管理
├── pkg/                   # 公共包
│   ├── conf/              # 配置加载
│   ├── json/              # JSON 处理
│   ├── logs/              # 日志
│   └── ...
└── go.mod / go.sum
```

---

## 3. API 层

### 3.1 路由结构

API 路由由 Hertz IDL 注解自动生成，路由结构如下：

```
/                                   # rootMw: CtxCache, AccessLog, Locale, PacketAdapter
├── /api                            # _apiMw: Session 认证
│   ├── /auth/v1
│   │   └── /personal_access_tokens  # PAT 管理
│   ├── /data/v1
│   │   ├── /datasets               # 数据集 CRUD
│   │   ├── /tags                   # 标签管理
│   │   └── /dataset_items          # 数据项管理
│   ├── /evaluation                 # 评估模块
│   ├── /observability              # 可观测性模块
│   ├── /prompt                     # Prompt 模块
│   └── /llm                        # LLM 管理
├── /ping                           # 健康检查
```

### 3.2 Handler 组织

每个业务模块对应一个 Handler 结构体：

| Handler | 职责 | 嵌入的服务接口 |
|---------|------|---------------|
| `FoundationHandler` | Auth, Space, User, File | AuthService, UserService, SpaceService, FileService |
| `LLMHandler` | LLM 模型管理 | LLMManageService, LLMRuntimeService |
| `PromptHandler` | Prompt 管理与调试 | PromptManageService, PromptDebugService, PromptExecuteService |
| `DataHandler` | 数据集管理 | IDatasetApplication, TagService |
| `EvaluationHandler` | 评估实验 | EvaluatorService, EvaluationSetService, EvalTargetService |
| `ObservabilityHandler` | Trace/Metric/Task | ITraceApplication, IObservabilityOpenAPIApplication |

### 3.3 中间件链

**rootMw**: `CtxCacheMW -> AccessLogMW -> LocaleMW -> PacketAdapterMW`

**_apiMw**: `SessionMW`

**_loopMw**: `PatTokenVerifyMW`

---

## 4. 业务模块详解

### 4.1 Foundation 模块 (`modules/foundation/`)

**职责**: 提供基础公共服务，包括认证授权、用户管理、空间管理、文件管理。

**子包结构**:
- `application/`: 应用层实现
  - `auth.go`: 认证应用
  - `authn.go`: 身份验证应用 (API Key)
  - `user.go`: 用户应用
  - `space.go`: 空间应用
  - `file.go`: 文件应用
- `domain/`: 领域层
  - `auth/service/`: 认证服务
  - `authn/`: 身份验证 (API Key 实体、仓库)
  - `user/`: 用户领域 (实体、仓库、服务)
  - `file/`: 文件领域服务
- `infra/repo/`: 数据仓库层
  - `mysql/`: MySQL 实现 (GORM)

**对外接口** (通过 Kitex RPC 调用):
- `AuthService`: 认证
- `AuthNService`: 身份验证
- `UserService`: 用户
- `SpaceService`: 空间
- `FileService`: 文件

### 4.2 Data 模块 (`modules/data/`)

**职责**: 数据集管理，支持 CSV/Parquet 文件导入、数据项版本管理、数据集 Schema 定义。

**子包结构**:
- `application/`: 数据集、数据项、任务、Schema 应用
  - `dataset_app.go`: 数据集管理
  - `item_app.go`: 数据项 CRUD
  - `job_app.go`: IO 任务 (导入/导出)
  - `schema_app.go`: Schema 管理
- `domain/`: 领域层
  - `dataset/entity/`: Dataset, DatasetItem, DatasetVersion, IOJob 实体
  - `dataset/service/`: 数据集服务 (导入/导出/Item 管理)
  - `dataset/repo/`: 数据仓库接口
  - `dataset/component/mq/`: 消息发布
  - `tag/entity/`: TagKey, TagValue 实体
  - `tag/service/`: 标签服务
- `infra/`: 基础设施
  - `repo/dataset/mysql/`: MySQL 实现 + GORM Generate
  - `repo/dataset/oss/`: OSS 存储
  - `repo/dataset/redis/`: Redis 缓存
  - `repo/tag/mysql/`: 标签 MySQL 实现
  - `vfs/`: 虚拟文件系统 (CSV/Parquet 解析)
  - `mq/consumer/`: RocketMQ 消费者

**关键设计**:
- 支持 CSV、Parquet 文件导入
- 数据集版本管理 (snapshot)
- VFS (Virtual File System) 统一文件抽象
- 软删机制

### 4.3 Evaluation 模块 (`modules/evaluation/`)

**职责**: AI 评估实验管理，包括 Evaluator、EvaluationSet、Experiment、EvalTarget。

**子包结构**:
- `application/`: 评估应用
  - `evaluator_app.go`: 评估器管理 (Evaluator)
  - `evaluation_set_app.go`: 评估集管理 (EvaluationSet)
  - `experiment_app.go`: 实验管理 (Experiment)
  - `eval_target_app.go`: 评估目标管理
  - `eval_openapi_app.go`: OpenAPI 接口
- `domain/`: 领域层
  - `evaluation/entity/`: 评估实体
  - `evaluation/service/`: 评估服务
  - `evaluation/repo/`: 仓库
  - `trajectory/`: 轨迹服务
- `infra/`: 基础设施
  - `rpc/data/`: 数据模块 RPC 适配器
  - `rpc/prompt/`: Prompt 模块 RPC 适配器
  - `mq/consumer/`: RocketMQ 消费者

**Consumer 配置**: `evaluation.yaml`

**依赖**:
- Data 模块 (Dataset)
- Prompt 模块 (PromptManage, PromptExecute)
- LLM 模块 (LLMRuntime)

### 4.4 LLM 模块 (`modules/llm/`)

**职责**: 多模型统一管理，支持 Eino 框架集成多种 LLM Provider。

**子包结构**:
- `application/`: LLM 应用
  - `manage.go`: 模型管理 (Model CRUD)
  - `runtime.go`: 运行时管理 (Chat/Completion)
- `domain/`: 领域层
  - `entity/`: Model, RuntimeOption, ChatMessage 等实体
  - `service/llmfactory/`: LLM 工厂 (创建不同类型的 LLM)
  - `service/llmimpl/eino/`: Eino 实现
- `infra/repo/`: 模型配置仓库 (MySQL/GORM)

**Eino 集成**:
Eino 是 ByteDance 的 AI LLM 集成框架，本模块通过 `llmfactory.FactoryImpl` 工厂创建 LLM 实例：

```go
// 支持的 Provider (根据 go.mod 推断)
- OpenAI (eino-ext/components/model/openai)
- Claude (eino-ext/components/model/claude)
- DeepSeek (eino-ext/components/model/deepseek)
- Gemini (eino-ext/components/model/gemini)
- Qianfan (eino-ext/components/model/qianfan)
- Qwen (eino-ext/components/model/qwen)
- Ollama (eino-ext/components/model/ollama)
- Ark (eino-ext/components/model/ark)
```

### 4.5 Prompt 模块 (`modules/prompt/`)

**职责**: Prompt 模板管理、调试、执行。

**子包结构**:
- `application/`: Prompt 应用
  - `manage.go`: Prompt 版本管理、发布
  - `debug.go`: Prompt 调试
  - `execute.go`: Prompt 执行
  - `openapi.go`: OpenAPI 接口
- `domain/`: 领域层
  - `entity/`: PromptDetail, PromptBasic, DebugContext, ExecuteContext
  - `service/`: 业务服务 (格式化、标签、执行)
  - `repo/`: 仓库接口
- `infra/`: 基础设施
  - `repo/mysql/`: MySQL 实现
  - `repo/redis/`: Redis 缓存 (Prompt 缓存、Label 版本缓存)
  - `metrics/`: PAAS 覆盖率指标

**模板引擎**: 支持 Go Template 和 Jinja2

### 4.6 Observability 模块 (`modules/observability/`)

**职责**: 可观测性数据管理，包括 Trace、Metric、Task 调度。

**子包结构**:
- `application/`: 可观测性应用
  - `trace.go`: 链路追踪
  - `metric.go`: 指标
  - `task.go`: 任务调度
  - `openapi.go`: OpenAPI
- `domain/`: 领域层
  - `trace/entity/`: Trace 实体
  - `trace/service/`: Trace 服务
  - `component/storage/`: 存储提供者
- `infra/`: 基础设施
  - `storage/`: Trace 存储实现
  - `mq/consumer/`: RocketMQ 消费者

**Consumer 配置**: `observability.yaml`

---

## 5. 基础设施层 (`infra/`)

### 5.1 数据库 (`infra/db/`)

**技术**: GORM + MySQL

**Provider 接口**:
```go
type Provider interface {
    NewSession(ctx context.Context, opts ...Option) *gorm.DB
    Transaction(ctx context.Context, fc func(tx *gorm.DB) error, opts ...Option) error
}
```

**配置选项**:
- `WithMaster()`: 强制读主库
- `WithTransaction(tx)`: 使用已有事务
- `Debug()`: 调试模式
- `WithDeleted()`: 返回软删数据
- `WithSelectForUpdate()`: SELECT FOR UPDATE

### 5.2 Redis (`infra/redis/`)

**技术**: go-redis/v9

**Provider 接口**: 实现了 go-redis 的 `Cmdable` 接口的子集，包括：
- String 操作: `Get`, `Set`, `Incr`, `Decr`, `MGet`, `MSet`
- Hash 操作: `HGet`, `HSet`, `HDel`, `HGetAll`
- ZSet 操作: `ZAdd`, `ZRange`
- List 操作: `RPush`, `LRange`
- Pipeline 支持

### 5.3 ClickHouse (`infra/ck/`)

**技术**: clickhouse-go/v2 + GORM ClickHouse Driver

**Provider 接口**:
```go
type Provider interface {
    NewSession(ctx context.Context) *gorm.DB
}
```

**用途**: Trace/Metric 等时序数据和分析数据存储

### 5.4 RocketMQ (`infra/mq/`)

**技术**: rocketmq-client-go/v2

**核心组件**:
- `Factory`: 创建 Producer/Consumer
- `ConsumerRegistry`: 消费者注册与管理，支持优雅关闭
- `Consumer`: Push Consumer 实现

**Consumer Worker 接口**:
```go
type IConsumerWorker interface {
    ConsumerCfg(ctx context.Context) (*ConsumerConfig, error)
}
```

**环境变量**:
- `COZE_LOOP_RMQ_NAMESRV_DOMAIN`
- `COZE_LOOP_RMQ_NAMESRV_PORT`
- `COZE_LOOP_RMQ_NAMESRV_USER`
- `COZE_LOOP_RMQ_NAMESRV_PASSWORD`

### 5.5 文件存储 (`infra/fileserver/`)

**技术**: AWS S3 SDK

**接口**:
- `ObjectStorage`: 单对象操作
- `BatchObjectStorage`: 批量操作

**配置**: 通过环境变量读取 S3 兼容存储配置

### 5.6 限流 (`infra/limiter/`)

**技术**: go-redis-rate (Redis 基于令牌桶)

**接口**:
- `IRateLimiterFactory`: 创建限流器
- `IPlainRateLimiterFactory`: 简单限流

### 5.7 中间件 (`infra/middleware/`)

| 中间件 | 文件位置 | 职责 |
|--------|----------|------|
| Session | `session/` | 基于 HMAC-SHA256 的会话验证 |
| CtxCache | `ctxcache/` | 请求上下文缓存 |
| Logs | `logs/` | 访问日志 |
| Validator | `validator/` | 请求参数校验 |

### 5.8 其他基础设施

| 组件 | 路径 | 职责 |
|------|------|------|
| ID 生成器 | `idgen/` | 基于 Redis 的 ID 生成 |
| 分布式锁 | `lock/` | Redis 分布式锁实现 |
| 链路追踪 | `looptracer/` | OpenTelemetry 集成 |
| 指标 | `metrics/` | OpenTelemetry Meter |
| 国际化 | `i18n/` | go-i18n 实现 |

---

## 6. 配置管理 (`pkg/conf/`)

**技术**: Viper

**配置加载**:
```go
// 配置文件查找顺序
1. 绝对路径
2. 当前目录
3. PWD 环境变量目录
4. 递归搜索

// 配置热加载
v.WatchConfig()  // fsnotify 监控文件变化
```

**ConfigLoaderFactory**:
```go
type IConfigLoaderFactory interface {
    NewConfigLoader(name string) (IConfigLoader, error)
}
```

**基础设施配置** (`infrastructure.yaml`):
- Redis
- MySQL (RDS)
- ClickHouse
- S3
- IDGen
- LogLevel

---

## 7. 入口程序 (`cmd/`)

### 7.1 main.go

主程序同时启动 HTTP 服务器和 Consumer Workers：

```go
func main() {
    // 1. 初始化组件 (newComponent)
    //    - Redis, MySQL, ClickHouse, S3, IDGenerator, etc.
    // 2. 初始化 API Handler (api.Init)
    //    - Wire DI 创建各模块 Handler
    // 3. 初始化 Tracer (initTracer)
    //    - LoopTracer 链路追踪
    // 4. 启动 Consumer Workers (registry.StartAll)
    // 5. 启动 HTTP Server (api.Start)
    // 6. 等待信号优雅关闭
}
```

### 7.2 consumer.go

Consumer Worker 初始化：

```go
// 注册三个 Consumer Worker
1. EvalConsumer (evaluation.yaml)     -> Evaluation 模块
2. DataConsumer (data job)            -> Data 模块
3. ObservabilityConsumer (observability.yaml) -> Observability 模块
```

---

## 8. Wire 依赖注入

**技术**: Google Wire (编译期 DI)

**注入点**: `api/handler/coze/loop/apis/wire.go`

**Provider Set**:
```go
foundationSet  // Foundation 模块
llmSet        // LLM 模块
promptSet     // Prompt 模块
dataSet       // Data 模块
evaluationSet // Evaluation 模块
observabilitySet // Observability 模块
```

---

## 9. IDL 与代码生成

### 9.1 Kitex Gen (`kitex_gen/`)

由 Thrift IDL 生成的 RPC 客户端/服务端代码，包含以下服务：

- `coze/loop/apis/`: API 定义
- `coze/loop/observability/`: 可观测性服务
- `coze/loop/evaluation/`: 评估服务
- `coze/loop/data/`: 数据服务
- `coze/loop/foundation/`: 基础服务
- `coze/loop/prompt/`: Prompt 服务
- `coze/loop/llm/`: LLM 服务

### 9.2 Loop Gen (`loop_gen/`)

本地服务桩代码 (Local Service)，用于同进程内的服务调用：

- `coze/loop/foundation/loauth/`: Auth 本地服务
- `coze/loop/foundation/lofile/`: File 本地服务
- `coze/loop/data/lodataset/`: Dataset 本地服务
- `coze/loop/evaluation/loevaluator/`: Evaluator 本地服务
- `coze/loop/prompt/lomanage/`: Prompt 管理本地服务

---

## 10. 关键设计模式

### 10.1 分层架构

```
API Handler (api/)
    ↓
Application (modules/*/application/)
    ↓
Domain (modules/*/domain/)
    ↓
Infrastructure (modules/*/infra/, infra/)
```

### 10.2 本地 RPC 调用

使用 `LocalXXXService` 模式实现同进程内的服务调用，避免网络开销：

```go
// 在 Handler 初始化时绑定
bindLocalCallClient(
    service,
    &localClient,
    NewLocalXXXService,
)
```

### 10.3 Consumer Worker 模式

RocketMQ Consumer 通过 Worker 接口抽象，支持动态注册：

```go
type IConsumerWorker interface {
    ConsumerCfg(ctx context.Context) (*ConsumerConfig, error)
}
```

---

## 11. 环境变量清单

| 变量名 | 用途 |
|--------|------|
| `PWD` | 工作目录 (用于配置查找) |
| `COZE_LOOP_REDIS_DOMAIN/PORT/PASSWORD` | Redis 连接 |
| `COZE_LOOP_MYSQL_DOMAIN/PORT/USER/PASSWORD/DATABASE` | MySQL 连接 |
| `COZE_LOOP_CLICKHOUSE_DOMAIN/PORT/USER/PASSWORD/DATABASE` | ClickHouse 连接 |
| `COZE_LOOP_OSS_PROTOCOL/DOMAIN/PORT/USER/PASSWORD/REGION/BUCKET` | S3 存储 |
| `COZE_LOOP_RMQ_NAMESRV_DOMAIN/PORT/USER/PASSWORD` | RocketMQ |
| `COZE_LOOP_SESSION_HMAC_KEY` | Session 签名密钥 |

---

## 12. 未确认项

1. **IDL 源文件**: 未在代码库中找到 `.thrift` 或 `.proto` IDL 文件，推测 IDL 存储在独立仓库
2. **conf/coze/ 和 conf/stone/ 目录**: 目录结构存在但为空，可能是历史遗留或用于特定部署
3. **评估模块详细配置**: `evaluation.yaml` 和 `observability.yaml` 配置文件内容未确认
