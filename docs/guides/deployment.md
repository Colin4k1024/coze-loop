# Deployment Guide

本文档描述 coze-loop 项目的部署架构、配置要求和操作流程。

## 1. 部署模式

coze-loop 支持两种部署模式：

1. **Docker Compose** - 适用于开发、测试、小规模部署
2. **Kubernetes (Helm)** - 适用于生产环境，支持高可用和弹性伸缩

## 2. Docker Compose 部署

### 2.1 目录结构

```
release/deployment/docker-compose/
├── docker-compose.yml          # 主编排文件
├── docker-compose-dev.yml      # 开发环境覆盖
├── docker-compose-debug.yml    # 调试配置
├── .env                        # 环境变量配置
├── bootstrap/                  # 各服务初始化脚本
│   ├── app/                   # 应用容器
│   │   ├── entrypoint.sh
│   │   └── healthcheck.sh
│   ├── redis/
│   ├── mysql/                 # MySQL + init
│   ├── clickhouse/            # ClickHouse + init
│   ├── minio/                 # MinIO + init
│   ├── rmq-namesrv/           # RocketMQ NameServer
│   ├── rmq-broker/            # RocketMQ Broker
│   ├── rmq-init/              # RocketMQ 初始化
│   ├── python-faas/            # Python FaaS
│   ├── js-faas/               # JS FaaS
│   └── nginx/                  # Nginx 反向代理
└── conf/                      # 配置文件挂载
    ├── infrastructure.yaml
    ├── observability.yaml
    ├── model_config.yaml
    └── locales/
```

### 2.2 服务组件

| 服务 | 容器名 | 端口 | 说明 |
|------|--------|------|------|
| `app` | coze-loop-app | 8888 | 主应用服务 |
| `redis` | coze-loop-redis | 6379 | Redis 缓存 |
| `mysql` | coze-loop-mysql | 3306 | MySQL 数据库 |
| `mysql-init` | coze-loop-mysql-init | - | MySQL 初始化任务 |
| `clickhouse` | coze-loop-clickhouse | 9000 | ClickHouse |
| `clickhouse-init` | coze-loop-clickhouse-init | - | ClickHouse 初始化 |
| `minio` | coze-loop-minio | 9000 | MinIO 对象存储 |
| `minio-init` | coze-loop-minio-init | - | MinIO 初始化 |
| `rocketmq-namesrv` | coze-loop-rmq-namesrv | 9876 | RocketMQ NameServer |
| `rocketmq-broker` | coze-loop-rmq-broker | - | RocketMQ Broker |
| `rocketmq-init` | coze-loop-rmq-init | - | RocketMQ 初始化 |
| `nginx` | coze-loop-nginx | 8082 | Nginx 反向代理 |
| `coze-loop-python-faas` | coze-loop-python-faas | 8000 | Python FaaS |
| `coze-loop-js-faas` | coze-loop-js-faas | 8000 | JS FaaS (Deno) |

### 2.3 环境变量配置

**文件**: `.env`

#### 镜像配置
```bash
# 应用镜像
COZE_LOOP_APP_IMAGE_REGISTRY=docker.io
COZE_LOOP_APP_IMAGE_REPOSITORY=cozedev
COZE_LOOP_APP_IMAGE_NAME=coze-loop
COZE_LOOP_APP_IMAGE_TAG=1.5.1
COZE_LOOP_APP_OPENAPI_PORT=8888

# Python FaaS 镜像
COZE_LOOP_PYTHON_FAAS_IMAGE_REGISTRY=docker.io
COZE_LOOP_PYTHON_FAAS_IMAGE_REPOSITORY=cozedev
COZE_LOOP_PYTHON_FAAS_IMAGE_NAME=coze-loop-python-faas
COZE_LOOP_PYTHON_FAAS_IMAGE_TAG=1.0.0
```

#### Redis 配置
```bash
COZE_LOOP_REDIS_IMAGE_REGISTRY=docker.io
COZE_LOOP_REDIS_IMAGE_REPOSITORY=library
COZE_LOOP_REDIS_IMAGE_NAME=redis
COZE_LOOP_REDIS_IMAGE_TAG=8.2.0
COZE_LOOP_REDIS_DOMAIN=coze-loop-redis
COZE_LOOP_REDIS_PORT=6379
COZE_LOOP_REDIS_PASSWORD=cozeloop-redis
```

#### MySQL 配置
```bash
COZE_LOOP_MYSQL_IMAGE_REGISTRY=docker.io
COZE_LOOP_MYSQL_IMAGE_REPOSITORY=library
COZE_LOOP_MYSQL_IMAGE_NAME=mysql
COZE_LOOP_MYSQL_IMAGE_TAG=9.4.0
COZE_LOOP_MYSQL_DOMAIN=coze-loop-mysql
COZE_LOOP_MYSQL_PORT=3306
COZE_LOOP_MYSQL_USER=root
COZE_LOOP_MYSQL_PASSWORD=cozeloop-mysql
COZE_LOOP_MYSQL_DATABASE=cozeloop-mysql
```

#### ClickHouse 配置
```bash
COZE_LOOP_CLICKHOUSE_IMAGE_REGISTRY=docker.io
COZE_LOOP_CLICKHOUSE_IMAGE_REPOSITORY=clickhouse
COZE_LOOP_CLICKHOUSE_IMAGE_NAME=clickhouse-server
COZE_LOOP_CLICKHOUSE_IMAGE_TAG=latest
COZE_LOOP_CLICKHOUSE_DOMAIN=coze-loop-clickhouse
COZE_LOOP_CLICKHOUSE_PORT=9000
COZE_LOOP_CLICKHOUSE_USER=default
COZE_LOOP_CLICKHOUSE_PASSWORD=cozeloop-clickhouse
COZE_LOOP_CLICKHOUSE_DATABASE=cozeloop-clickhouse
```

#### MinIO 配置
```bash
COZE_LOOP_OSS_IMAGE_REGISTRY=docker.io
COZE_LOOP_OSS_IMAGE_REPOSITORY=minio
COZE_LOOP_OSS_IMAGE_NAME=minio
COZE_LOOP_OSS_IMAGE_TAG=RELEASE.2025-06-13T11-33-47Z
COZE_LOOP_OSS_PROTOCOL=http
COZE_LOOP_OSS_DOMAIN=coze-loop-minio
COZE_LOOP_OSS_PORT=9000
COZE_LOOP_OSS_REGION=us-east-1
COZE_LOOP_OSS_USER=root
COZE_LOOP_OSS_PASSWORD=cozeloop-minio
COZE_LOOP_OSS_BUCKET=cozeloop-minio
```

#### RocketMQ 配置
```bash
COZE_LOOP_RMQ_IMAGE_REGISTRY=docker.io
COZE_LOOP_RMQ_IMAGE_REPOSITORY=apache
COZE_LOOP_RMQ_IMAGE_NAME=rocketmq
COZE_LOOP_RMQ_IMAGE_TAG=5.3.3
COZE_LOOP_RMQ_NAMESRV_DOMAIN=coze-loop-rmq-namesrv
COZE_LOOP_RMQ_NAMESRV_PORT=9876
```

#### Nginx 配置
```bash
COZE_LOOP_NGINX_IMAGE_REGISTRY=docker.io
COZE_LOOP_NGINX_IMAGE_REPOSITORY=library
COZE_LOOP_NGINX_IMAGE_NAME=nginx
COZE_LOOP_NGINX_IMAGE_TAG=1.28.0
COZE_LOOP_NGINX_PORT=8082
COZE_LOOP_NGINX_DATA_VOLUME_NAME=coze-loop-nginx-data
```

#### FaaS 配置
```bash
# Python FaaS
COZE_LOOP_PYTHON_FAAS_DOMAIN=coze-loop-python-faas
COZE_LOOP_PYTHON_FAAS_PORT=8000

# JS FaaS
COZE_LOOP_JS_FAAS_DOMAIN=coze-loop-js-faas
COZE_LOOP_JS_FAAS_PORT=8000

# Deno 配置
DENO_DIR=/tmp/.deno
DENO_NO_UPDATE_CHECK=1
DENO_V8_FLAGS=--max-old-space-size=2048

# FaaS 基础配置
FAAS_WORKSPACE=/tmp/faas-workspace
FAAS_PORT=8000
FAAS_TIMEOUT=30000
FAAS_LANGUAGE=python

# 预装 Python 包版本
NUMPY_VERSION=>=1.24.0
PANDAS_VERSION=>=1.5.0
JSONSCHEMA_VERSION=>=4.0.0
SCIPY_VERSION=>=1.10.0
SKLEARN_VERSION=>=1.3.0
```

### 2.4 启动 Docker Compose

```bash
# 进入部署目录
cd release/deployment/docker-compose

# 启动所有服务（使用 app 和 nginx profile）
docker-compose --profile app --profile nginx up -d

# 仅启动基础设施（不启动 app）
docker-compose --profile redis --profile mysql --profile clickhouse up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f app
```

### 2.5 服务依赖关系

启动顺序（通过 `depends_on` 和 `condition` 控制）：

```
mysql (健康检查)
  └── mysql-init (执行完)
        └── app (健康检查)

redis (健康检查)
  └── app

clickhouse (健康检查)
  └── clickhouse-init (执行完)
        └── app

minio (健康检查)
  └── minio-init (执行完)
        └── app

rocketmq-namesrv (健康检查)
  └── rocketmq-broker (健康检查)
        └── rocketmq-init (执行完)
              └── app

python-faas (健康检查)
  └── app

js-faas (健康检查)
  └── app

nginx (依赖 app 健康)
  └── 反向代理前端
```

## 3. Kubernetes Helm 部署

### 3.1 目录结构

```
release/deployment/helm-chart/
├── umbrella/              # 统一 Chart（包含所有组件）
│   ├── Chart.yaml
│   ├── values.yaml
│   ├── conf/              # 配置文件
│   ├── templates/         # K8s 资源模板
│   └── examples/
│       └── minikube/     # Minikube 示例
└── charts/               # 独立 Chart
    ├── app/              # 主应用 Chart
    ├── mysql/
    ├── redis/
    ├── clickhouse/
    ├── minio/
    ├── rmq-namesrv/
    ├── rmq-broker/
    ├── python-faas/
    ├── js-faas/
    └── nginx/
```

### 3.2 Umbrella Chart 配置

**文件**: `release/deployment/helm-chart/umbrella/values.yaml`

```yaml
service:
  type: ClusterIP
  port: 8888
  targetPort: 8888

image:
  registry: "docker.io"
  repository: "cozedev"
  image: "coze-loop"
  tag: "1.5.1"
  pullPolicy: Always
  pullSecrets: "coze-loop-image-secret"

deployment:
  replicaCount: 1
  terminationGracePeriodSeconds: 5

liveness:
  startSeconds: 10
  intervalSeconds: 10
  timeoutSeconds: 30
  shutdownFailureTimes: 30

env:
  redis:
    domain: "coze-loop-redis"
    port: "6379"
    password: "cozeloop"
  mysql:
    domain: "coze-loop-mysql"
    port: "3306"
    user: "root"
    password: "cozeloop"
    database: "coze-loop"
  clickhouse:
    domain: "coze-loop-clickhouse"
    port: "9000"
    user: "default"
    password: "cozeloop"
    database: "coze-loop"
  oss:
    protocol: "http"
    domain: "coze-loop-minio"
    port: "9000"
    region: "us-east-1"
    user: "root"
    password: "cozeloop"
    bucket: "coze-loop"
  rmq:
    namesrv:
      domain: "coze-loop-rmq-namesrv"
      port: "9876"
      user: ""
      password: ""
  pythonFaas:
    port: 8000
  jsFaas:
    port: 8000
```

### 3.3 自定义配置 (custom)

通过 `custom` 配置项覆盖默认基础设施：

```yaml
custom:
  image:
    registry: ""
    pullSecrets: ""
  redis:
    disabled: true    # 使用外部 Redis
    domain: "external-redis.example.com"
    port: "6379"
    password: "external-password"
  mysql:
    disabled: true
    domain: "external-mysql.example.com"
    port: "3306"
    user: "root"
    password: "external-password"
    database: "coze_loop"
  clickhouse:
    disabled: true
    domain: "external-clickhouse.example.com"
    port: "9000"
    user: "default"
    password: "external-password"
    database: "coze_loop"
  oss:
    disabled: true
    protocol: "https"
    domain: "s3.amazonaws.com"
    bucket: "my-bucket"
  rmq:
    disabled: true
    namesrv:
      domain: "external-rmq.example.com"
      port: "9876"
```

### 3.4 部署命令

```bash
# 添加 Helm 仓库（如果使用远程 Chart）
helm repo add coze-loop https://charts.example.com/coze-loop

# 更新 Helm 依赖
helm dependency update

# 部署（使用默认配置）
helm install coze-loop ./umbrella

# 部署（自定义配置）
helm install coze-loop ./umbrella \
  --set image.tag=1.5.2 \
  --set custom.redis.disabled=true \
  --set custom.redis.domain=external-redis.example.com

# 升级部署
helm upgrade coze-loop ./umbrella --set image.tag=1.5.2

# 查看部署状态
helm status coze-loop

# 删除部署
helm uninstall coze-loop
```

### 3.5 Kubernetes 资源

Umbrella Chart 生成以下资源：

- **Deployment**: coze-loop (主应用)
- **Service**: coze-loop (ClusterIP)
- **ConfigMap**: coze-loop-runtime-configmap, coze-loop-locales-configmap
- **Secret**: coze-loop (存储密码等敏感信息)

Init Containers (在 app Pod 内):

- wait-for-redis-init
- wait-for-mysql-init
- wait-for-clickhouse-init
- wait-for-minio-init
- wait-for-rmq-init

## 4. 数据库迁移

### 4.1 MySQL 迁移

MySQL 使用初始化容器执行 SQL 脚本。

**SQL 脚本位置**:
- `bootstrap/mysql/init-sql/` - 初始建表脚本
- `bootstrap/mysql/patch-sql/` - 变更脚本

**执行流程** (mysql-init):

1. 等待 MySQL 服务健康
2. 循环执行 `init-sql/*.sql` 建表
3. 执行 `patch-sql/alter_proc.sql` 存储过程
4. 循环执行 `patch-sql/*_alter.sql` 变更

**关键表结构**:
- `dataset` - 数据集
- `dataset_item` - 数据集条目
- `dataset_version` - 数据集版本
- `evaluator` - 评估器
- `evaluator_template` - 评估器模板
- `evaluator_version` - 评估器版本
- `eval_target` - 评估目标
- `eval_set` - 评估集合
- `prompt` - 提示词
- `api_key` - API 密钥

### 4.2 ClickHouse 迁移

**SQL 脚本位置**:
- `bootstrap/clickhouse-init/init-sql/`

**执行流程** (clickhouse-init):

1. 等待 ClickHouse 服务健康
2. 创建数据库（如果不存在）
3. 循环执行 `init-sql/*.sql` 建表

**关键表结构**:
- `observability_spans` - 追踪 Span 表
- `observability_annotations` - 标注表
- `evaluation` - 评估数据表

## 5. 健康检查

### 5.1 应用健康检查

**端点**: `GET /ping` -> 返回 `pong`

```bash
# Docker Compose 内检查
docker exec coze-loop-app wget -qO- http://localhost:8888/ping | grep pong

# Kubernetes 内检查
kubectl exec -it coze-loop-app -- wget -qO- http://localhost:8888/ping
```

**健康检查脚本** (`bootstrap/app/healthcheck.sh`):
```bash
#!/bin/sh
if wget -qO- http://localhost:8888/ping 2>/dev/null | grep -q pong; then
  exit 0
else
  exit 1
fi
```

### 5.2 各服务健康检查

| 服务 | 检查方式 | 超时 |
|------|---------|------|
| Redis | `redis-cli -a password ping` | 5s |
| MySQL | `mysql -e "SELECT 1"` | 5s |
| ClickHouse | `clickhouse-client --query "SELECT 1"` | 5s |
| MinIO | `mc ready` 或 HTTP 检查 | 5s |
| RocketMQ | 检查 NameServer 端口 | 15s |
| Nginx | `nginx -t` | 3s |

### 5.3 部署验证清单

部署完成后，验证以下项目：

- [ ] 所有容器/Pod 处于 Running/Ready 状态
- [ ] 应用 `/ping` 端点返回 `pong`
- [ ] MySQL 可以正常连接和查询
- [ ] Redis 可以正常连接和操作
- [ ] ClickHouse 可以正常连接和查询
- [ ] MinIO 控制台可访问
- [ ] RocketMQ 控制台可访问（如果启用）
- [ ] Nginx 反向代理正常工作
- [ ] Python/JS FaaS 服务正常响应

## 6. 回滚操作

### 6.1 Docker Compose 回滚

```bash
# 1. 停止当前服务
docker-compose down

# 2. 修改 .env 中的镜像版本
COZE_LOOP_APP_IMAGE_TAG=1.5.0  # 回滚到旧版本

# 3. 重新拉取并启动
docker-compose pull
docker-compose up -d

# 4. 验证回滚
docker exec coze-loop-app wget -qO- http://localhost:8888/ping
```

### 6.2 Kubernetes 回滚

```bash
# 1. 查看部署历史
helm history coze-loop

# 2. 回滚到上一个版本
helm rollback coze-loop

# 或回滚到指定版本
helm rollback coze-loop 3

# 3. 验证回滚
kubectl rollout status deployment/coze-loop
kubectl exec -it coze-loop-app -- wget -qO- http://localhost:8888/ping
```

### 6.3 数据库回滚

数据库变更通过 SQL 脚本执行。如需回滚：

1. **MySQL**: 需要手动执行逆变更 SQL
2. **ClickHouse**: 使用 `ALTER TABLE ... DROP COLUMN` 等逆操作

**注意**: 生产环境数据库回滚需要谨慎，建议：
- 保留变更前的数据库快照
- 在测试环境验证逆变更 SQL
- 在低峰期执行回滚

## 7. 密钥管理

### 7.1 Docker Compose 密钥

通过 `.env` 文件管理：
```bash
# 敏感信息在 .env 中明文存储
# 生产环境建议使用 .env 文件外的密钥管理方案

# 示例：使用 Docker Secrets（需 Docker Swarm 模式）
echo "my-secret-password" | docker secret create coze_loop_mysql_password -
```

### 7.2 Kubernetes Secret

Helm Chart 自动创建 Secret 资源：

```yaml
# Secret 内容
data:
  redis-password: <base64 encoded>
  mysql-user: <base64 encoded>
  mysql-password: <base64 encoded>
  clickhouse-user: <base64 encoded>
  clickhouse-password: <base64 encoded>
  oss-user: <base64 encoded>
  oss-password: <base64 encoded>
  rmq-namesrv-user: <base64 encoded>
  rmq-namesrv-password: <base64 encoded>
```

**创建外部 Secret**:
```bash
kubectl create secret generic coze-loop-secrets \
  --from-literal=redis-password=your-password \
  --from-literal=mysql-password=your-password \
  --from-literal=clickhouse-password=your-password \
  --from-literal=oss-password=your-password \
  --namespace=coze-loop
```

## 8. 持久化存储

### 8.1 Docker Compose 卷

```yaml
volumes:
  redis_data:          # Redis 数据
  mysql_data:          # MySQL 数据
  clickhouse_data:     # ClickHouse 数据
  minio_data:          # MinIO 数据
  minio_config:        # MinIO 配置
  rmq_namesrv_data:    # RocketMQ NameServer 数据
  rmq_broker_data:     # RocketMQ Broker 数据
  nginx_data:          # Nginx 静态资源
  python_faas_workspace:  # Python FaaS 工作区
  js_faas_workspace:      # JS FaaS 工作区
```

### 8.2 Kubernetes 持久卷

建议使用 PersistentVolumeClaim：

```yaml
# values.yaml 中配置
persistence:
  enabled: true
  storageClass: "standard"
  accessMode: ReadWriteOnce
  size: 10Gi
```

## 9. 资源限制

### 9.1 容器资源请求和限制

**Docker Compose** (示例):
```yaml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
        reservations:
          cpus: '0.5'
          memory: 1G
```

**Kubernetes** (values.yaml):
```yaml
resources:
  limits:
    cpu: "2"
    memory: "4Gi"
  requests:
    cpu: "500m"
    memory: "1Gi"
```

### 9.2 建议资源配置

| 服务 | CPU 请求 | CPU 限制 | 内存请求 | 内存限制 |
|------|---------|---------|---------|---------|
| app | 500m | 2 | 1Gi | 4Gi |
| redis | 250m | 1 | 512Mi | 2Gi |
| mysql | 500m | 2 | 1Gi | 4Gi |
| clickhouse | 1 | 4 | 2Gi | 8Gi |
| minio | 250m | 1 | 512Mi | 2Gi |
| rocketmq | 500m | 2 | 1Gi | 4Gi |

## 10. 网络配置

### 10.1 Docker Compose 网络

```yaml
networks:
  coze-loop-network:
    driver: bridge
```

所有服务通过服务名进行内部通信：
- `coze-loop-mysql:3306`
- `coze-loop-redis:6379`
- `coze-loop-clickhouse:9000`
- `coze-loop-minio:9000`
- `coze-loop-rmq-namesrv:9876`
- `coze-loop-app:8888`

### 10.2 Kubernetes Service

服务间通过 Kubernetes Service DNS 通信：
- `coze-loop-mysql:3306`
- `coze-loop-redis:6379`
- `coze-loop-clickhouse:9000`
- `coze-loop-minio:9000`
- `coze-loop-rmq-namesrv:9876`
- `coze-loop:8888`

## 11. 日志配置

### 11.1 应用日志

应用日志通过结构化日志输出到 stdout，在 Docker/Kubernetes 环境中被容器运行时收集。

**日志级别配置** (`infrastructure.yaml`):
```yaml
infra:
  log_level: "info"  # debug/info/warn/error
```

### 11.2 各服务日志位置

| 服务 | 日志位置 |
|------|---------|
| app | stdout (容器日志) |
| redis | `/var/log/redis/redis.log` |
| mysql | `/var/log/mysql/` |
| clickhouse | `/var/log/clickhouse-server/` |
| minio | stdout |
| rocketmq | `/store/log` |

### 11.3 日志收集建议

生产环境建议配置日志收集：
- **Docker**: 使用 `docker logs` 或日志驱动
- **Kubernetes**: 使用 Fluentd/Fluent Bit + Elasticsearch/Loki
- **云环境**: 使用云厂商日志服务（CloudWatch Logs、Azure Monitor 等）

## 12. 常见问题

### 12.1 服务无法启动

1. **检查依赖服务**: `docker-compose ps` 或 `kubectl get pods`
2. **查看日志**: `docker-compose logs <service>` 或 `kubectl logs <pod>`
3. **检查资源**: 磁盘空间、内存、CPU 是否充足
4. **检查端口**: 端口是否被占用

### 12.2 数据库连接失败

1. 确认 MySQL 服务健康
2. 检查用户名/密码是否正确
3. 检查数据库是否已创建
4. 检查防火墙/网络安全策略

### 12.3 镜像拉取失败

1. 检查镜像仓库是否可访问
2. 确认镜像 tag 是否存在
3. 配置镜像代理/镜像加速器
4. 配置 Image Pull Secrets:
```yaml
image:
  pullSecrets: "coze-loop-image-secret"
```

### 12.4 PVC 挂载失败

1. 检查 StorageClass 是否可用
2. 确认 PVC 请求的存储大小是否超过可用配额
3. 检查 PV/PVC 状态: `kubectl get pvc`

## 13. 安全建议

1. **修改默认密码**: 生产环境务必修改 `.env` 或 values.yaml 中的默认密码
2. **启用 TLS**: MySQL、Redis、RocketMQ 等组件建议启用 TLS 加密
3. **网络隔离**: 使用 Kubernetes NetworkPolicy 限制 Pod 间通信
4. **密钥管理**: 使用外部密钥管理服务（Vault、AWS Secrets Manager 等）
5. **镜像安全**: 使用私有镜像仓库，定期扫描镜像漏洞
6. **最小权限**: 数据库用户使用最小必要权限
7. **日志脱敏**: 确保敏感信息不在日志中明文输出
