# Backend 开发指南

## 1. 环境准备

### 1.1 前置依赖

| 依赖 | 版本要求 | 说明 |
|------|----------|------|
| Go | 1.24+ | 推荐使用 go.mod 指定的版本 |
| Wire | 最新版 | `go install github.com/google/wire/cmd/wire@latest` |
| Git | 任意版本 | 代码管理 |

### 1.2 环境变量配置

在开发环境需要设置以下环境变量：

```bash
# Redis
export COZE_LOOP_REDIS_DOMAIN=localhost
export COZE_LOOP_REDIS_PORT=6379
export COZE_LOOP_REDIS_PASSWORD=

# MySQL
export COZE_LOOP_MYSQL_DOMAIN=localhost
export COZE_LOOP_MYSQL_PORT=3306
export COZE_LOOP_MYSQL_USER=root
export COZE_LOOP_MYSQL_PASSWORD=your_password
export COZE_LOOP_MYSQL_DATABASE=coze_loop

# ClickHouse
export COZE_LOOP_CLICKHOUSE_DOMAIN=localhost
export COZE_LOOP_CLICKHOUSE_PORT=9000
export COZE_LOOP_CLICKHOUSE_USER=default
export COZE_LOOP_CLICKHOUSE_PASSWORD=
export COZE_LOOP_CLICKHOUSE_DATABASE=default

# S3 (MinIO for local dev)
export COZE_LOOP_OSS_PROTOCOL=http
export COZE_LOOP_OSS_DOMAIN=localhost
export COZE_LOOP_OSS_PORT=9000
export COZE_LOOP_OSS_USER=minioadmin
export COZE_LOOP_OSS_PASSWORD=minioadmin
export COZE_LOOP_OSS_REGION=us-east-1
export COZE_LOOP_OSS_BUCKET=coze-loop

# RocketMQ
export COZE_LOOP_RMQ_NAMESRV_DOMAIN=localhost
export COZE_LOOP_RMQ_NAMESRV_PORT=9876
export COZE_LOOP_RMQ_NAMESRV_USER=
export COZE_LOOP_RMQ_NAMESRV_PASSWORD=

# Session
export COZE_LOOP_SESSION_HMAC_KEY=your-secret-key

# 工作目录
export PWD=/path/to/coze-loop/backend
```

### 1.3 本地服务启动

需要本地运行的服务：

1. **MySQL 8.0+**: 主数据库
2. **Redis 7.0+**: 缓存和会话
3. **ClickHouse**: 分析型数据库
4. **RocketMQ**: 消息队列
5. **MinIO** (可选): S3 兼容存储，用于文件上传

---

## 2. 项目结构

```
backend/
├── api/                    # HTTP API 层 (Hertz)
│   ├── handler/            # 业务 Handler
│   └── router/             # 路由注册
├── cmd/                    # 程序入口
│   ├── main.go            # 主程序
│   └── consumer.go        # Consumer 初始化
├── conf/                   # 配置文件
├── infra/                  # 基础设施
│   ├── db/                # MySQL
│   ├── redis/             # Redis
│   ├── mq/                # RocketMQ
│   └── ...
├── kitex_gen/             # Kitex 生成的 RPC 代码
├── loop_gen/              # 本地服务桩代码
├── modules/               # 业务模块
│   ├── data/              # 数据集
│   ├── evaluation/        # 评估
│   ├── foundation/        # 基础服务
│   ├── llm/               # LLM
│   ├── observability/     # 可观测性
│   └── prompt/            # Prompt
└── pkg/                   # 公共包
```

---

## 3. 构建与运行

### 3.1 下载依赖

```bash
cd backend
go mod download
```

### 3.2 生成 Wire 代码

当修改了 `wire.go` 文件后，需要重新生成 Wire 代码：

```bash
cd backend
wire generate ./api/handler/coze/loop/apis/...
```

### 3.3 编译

```bash
# 编译主程序
go build -o coze-loop ./cmd/main.go

# 运行
./coze-loop
```

### 3.4 运行测试

```bash
# 运行所有测试
go test ./...

# 运行指定模块测试
go test ./modules/data/...

# 运行测试并显示覆盖率
go test -cover ./modules/data/...

# 运行测试并输出详细日志
go test -v ./modules/data/...
```

---

## 4. 模块开发指南

### 4.1 添加新的 API 端点

#### 步骤 1: 定义 Handler

在 `api/handler/coze/loop/apis/` 下创建或修改 Handler 文件：

```go
// example_service.go
package apis

type ExampleHandler struct {
    exampleApp ExampleApplication
}

func (h *ExampleHandler) CreateExample(ctx context.Context, c *app.RequestContext) {
    var req *ExampleCreateRequest
    if err := c.BindAndValidate(&req); err != nil {
        // 处理错误
        return
    }

    resp, err := h.exampleApp.Create(ctx, req)
    if err != nil {
        // 处理错误
        return
    }

    c.JSON(http.StatusOK, resp)
}
```

#### 步骤 2: 注册路由

路由由 IDL 注解生成，通常不需要手动修改。但如果是自定义路由，在 `api/router.go` 的 `customizedRegister` 中添加：

```go
func customizedRegister(r *server.Hertz) {
    r.POST("/api/v1/examples", handler.CreateExample)
}
```

#### 步骤 3: 注入依赖

在 `wire.go` 中添加 Provider：

```go
var exampleSet = wire.NewSet(
    NewExampleHandler,
    exampleapp.InitExampleApplication,
)

func InitExampleHandler(
    db db.Provider,
    redis redis.Cmdable,
) (*ExampleHandler, error) {
    wire.Build(exampleSet)
    return nil, nil
}
```

### 4.2 添加新的业务模块

#### 步骤 1: 创建模块目录结构

```
modules/newmodule/
├── application/           # 应用层
│   ├── newmodule_app.go
│   └── wire.go
├── domain/                # 领域层
│   ├── entity/           # 实体
│   ├── service/          # 服务
│   └── repo/             # 仓库接口
├── infra/                # 基础设施
│   └── repo/             # 仓库实现
└── pkg/                  # 模块内公共包
    └── errno/            # 错误码
```

#### 步骤 2: 定义应用层接口

```go
// application/newmodule_app.go
type INewModuleApplication interface {
    Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error)
    Get(ctx context.Context, id int64) (*Entity, error)
    List(ctx context.Context, req *ListRequest) ([]*Entity, error)
}
```

#### 步骤 3: 实现领域层

```go
// domain/entity/entity.go
type Entity struct {
    ID        int64
    Name      string
    CreatedAt time.Time
}

// domain/repo/repo.go
type IRepo interface {
    Create(ctx context.Context, e *Entity) error
    GetByID(ctx context.Context, id int64) (*Entity, error)
}

// domain/service/service.go
type Service struct {
    repo IRepo
}

func (s *Service) Create(ctx context.Context, req *CreateRequest) (*Entity, error) {
    entity := &Entity{
        Name: req.Name,
    }
    if err := s.repo.Create(ctx, entity); err != nil {
        return nil, err
    }
    return entity, nil
}
```

#### 步骤 4: 实现基础设施层

```go
// infra/repo/mysql/repo_impl.go
type RepoImpl struct {
    db db.Provider
}

func (r *RepoImpl) Create(ctx context.Context, e *Entity) error {
    session := r.db.NewSession(ctx)
    return session.Create(e).Error
}
```

#### 步骤 5: Wire 注入

```go
// application/wire.go
//go:build wireinject
// +build wireinject

var newModuleSet = wire.NewSet(
    NewService,
    repo.NewRepoImpl,
    wire.Bind(new(IRepo), new(*repoImpl)),
)

func InitNewModuleApplication(db db.Provider) (INewModuleApplication, error) {
    wire.Build(newModuleSet)
    return nil, nil
}
```

### 4.3 添加新的基础设施组件

#### 添加新的数据库支持

在 `infra/` 下创建新的目录：

```
infra/newdb/
├── newdb.go          # Provider 定义和实现
└── config.go         # 配置结构
```

```go
// newdb.go
type Provider interface {
    NewSession(ctx context.Context) *gorm.DB
}

type provider struct {
    db *gorm.DB
}

func (p *provider) NewSession(ctx context.Context) *gorm.DB {
    return p.db.WithContext(ctx)
}

func NewDBFromConfig(cfg *Config) (Provider, error) {
    // 实现
}
```

#### 添加新的消息队列 Consumer

```go
// infra/mq/consumer/example_consumer.go
type ExampleConsumer struct {
    app ExampleApplication
}

func (c *ExampleConsumer) ConsumerCfg(ctx context.Context) (*mq.ConsumerConfig, error) {
    return &mq.ConsumerConfig{
        Addr:         getMqAddr(),
        Topic:        "example_topic",
        ConsumerGroup: "example_group",
    }, nil
}

func (c *ExampleConsumer) HandleMessage(ctx context.Context, msg *mq.MessageExt) error {
    // 处理消息
    return nil
}

// 在 consumer.go 中注册
func MustInitConsumerWorkers(...) []mq.IConsumerWorker {
    workers, err := exampleconsumer.NewConsumerWorkers(loader, exampleApplication)
    // ...
}
```

### 4.4 添加新的 LLM Provider

在 `modules/llm/domain/service/llmfactory/` 添加新的工厂方法：

```go
// factory.go
func (f *FactoryImpl) CreateLLM(ctx context.Context, model *entity.Model, opts ...entity.Option) (llminterface.ILLM, error) {
    switch model.Provider {
    case "openai":
        return openai.NewLLM(ctx, model, opts...)
    case "claude":
        return claude.NewLLM(ctx, model, opts...)
    // 添加新的 Provider
    default:
        return nil, errorx.NewByCode(llm_errorx.ModelInvalidCode, "unsupported provider")
    }
}
```

---

## 5. Wire 依赖注入

### 5.1 基本概念

Wire 是 Google 提供的编译期依赖注入工具，通过 `//go:build wireinject` 标记注入点。

### 5.2 Provider Set

使用 `wire.NewSet` 创建 Provider 集合：

```go
var foundationSet = wire.NewSet(
    NewFoundationHandler,
    foundationapp.InitAuthApplication,
    foundationapp.InitAuthNApplication,
    foundationapp.InitSpaceApplication,
    foundationapp.InitUserApplication,
    wire.Bind(new(authservice.Client), new(*loauth.LocalAuthService)),
    loauth.NewLocalAuthService,
)
```

### 5.3 绑定接口

```go
wire.Bind(new(InterfaceType), new(*ConcreteType))
```

### 5.4 重新生成

修改 `wire.go` 后运行：

```bash
wire generate ./api/handler/coze/loop/apis/...
```

---

## 6. 配置管理

### 6.1 配置文件格式

使用 YAML 格式，例如 `conf/infrastructure.yaml`：

```yaml
infra:
  redis:
    host: "localhost"
    port: 6379
    password: ""
  rds:
    user: "root"
    password: "password"
    host: "localhost"
    port: "3306"
    db: "coze_loop"
```

### 6.2 加载配置

```go
// 使用 viper 加载
cfgFactory := viper.NewFileConfigLoaderFactory(viper.WithFactoryConfigPath("conf"))
loader, err := cfgFactory.NewConfigLoader("infrastructure.yaml")

var config ComponentConfig
err = loader.UnmarshalKey(ctx, "infra", &config)
```

### 6.3 热加载

Viper 支持配置热加载：

```go
v.WatchConfig()
v.OnConfigChange(func(e fsnotify.Event) {
    logs.Info("config changed: %s", e.Name)
})
```

---

## 7. 测试指南

### 7.1 测试组织

```
modules/
└── data/
    ├── application/
    │   ├── dataset_app.go      # 实现
    │   └── dataset_app_test.go  # 测试
    └── domain/
        └── dataset/
            └── service/
                ├── service.go
                └── service_test.go
```

### 7.2 Mock 使用

使用 `mockgen` 生成 Mock：

```bash
# 安装 mockgen
go install github.com/golang/mock/mockgen@latest

# 生成 Mock (在对应目录运行)
mockgen -destination=mocks/repo.go -package=mocks . IRepo
```

### 7.3 测试数据库

使用 `sqlmock` 进行数据库测试：

```go
import (
    "github.com/DATA-DOG/go-sqlmock"
    "gorm.io/driver/mysql"
)

func TestExample(t *testing.T) {
    db, mock, _ := sqlmock.New()
    gormDB, _ := gorm.Open(mysql.New(mysql.Config{
        Conn:                      db,
        SkipInitializeWithVersion: true,
    }), &gorm.Config{})

    mock.ExpectQuery("SELECT").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
            AddRow(1, "test"))

    // 执行测试
}
```

### 7.4 Redis 测试

使用 `miniredis` 进行 Redis 测试：

```go
import "github.com/alicebob/miniredis/v2"

func TestRedis(t *testing.T) {
    mr, _ := miniredis.Run()
    defer mr.Close()

    client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
    // 测试代码
}
```

---

## 8. 代码规范

### 8.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写，简短 | `repo`, `service` |
| 结构体 | 大写开头，驼峰 | `DatasetService` |
| 方法 | 大写开头，驼峰 | `CreateDataset` |
| 变量 | 小写开头，驼峰 | `datasetID` |
| 常量 | 全大写，下划线分隔 | `DefaultPageSize` |
| 接口 | 大写开头，以 `I` 开头 | `IApplication` |

### 8.2 错误处理

```go
// 使用 errorx 包装错误
return nil, errorx.Wrapf(err, "create dataset failed, name: %s", name)

// 业务错误码
return nil, errorx.NewByCode(errno.DataNotFoundCode, "dataset not found")
```

### 8.3 Context 传递

```go
// 在入口处传递 context
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
    // ctx 已自动从 Hertz 传递
}

// 传递给下游
result, err := h.app.Create(ctx, req)
```

### 8.4 日志

```go
import "github.com/coze-dev/coze-loop/backend/pkg/logs"

// 调试日志
logs.CtxDebug(ctx, "creating dataset, name: %s", name)

// 信息日志
logs.CtxInfo(ctx, "dataset created, id: %d", id)

// 错误日志
logs.CtxError(ctx, "create dataset failed: %v", err)
```

---

## 9. 数据库迁移

### 9.1 GORM Auto Migrate

在应用启动时自动迁移：

```go
// 在初始化时调用
func autoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &Dataset{},
        &DatasetItem{},
        &DatasetVersion{},
    )
}
```

### 9.2 手动 SQL 迁移

对于复杂的迁移，创建迁移文件：

```sql
-- migrations/20240101_create_dataset.sql
CREATE TABLE IF NOT EXISTS `dataset` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## 10. 常见问题

### 10.1 Wire 编译失败

**问题**: `wire: generate failed`

**解决**:
1. 检查是否有循环依赖
2. 确保所有 Provider 都有构造函数
3. 运行 `go mod tidy` 整理依赖

### 10.2 数据库连接失败

**问题**: `connect refused` 或 `Access denied`

**解决**:
1. 确认 MySQL 服务正在运行
2. 检查环境变量配置
3. 验证用户名密码

### 10.3 Consumer 无法启动

**问题**: Consumer 注册成功但不消费消息

**解决**:
1. 检查 RocketMQ 是否运行
2. 验证 Topic 和 ConsumerGroup 配置
3. 查看 Consumer 日志

### 10.4 LLM 调用失败

**问题**: 模型调用返回错误

**解决**:
1. 检查模型配置是否正确
2. 验证 API Key/Token 是否有效
3. 查看限流配置

---

## 11. 相关文档

- [项目 README](../../README.md)
- [部署文档](../deployment/README.md) (如存在)
- [API 文档](./api.md) (如存在)
- [前端架构文档](./frontend.md) (如存在)
