---
artifact: arch-design
task: project-docs-organization
date: 2026-03-31
role: architect
status: draft
---

# 架构设计

> 本文档基于代码探索初步产出，详细内容待阶段 1 深度探索后补充完善。

## 系统边界

### 外部依赖

| 依赖 | 用途 | 集成方式 |
|------|------|----------|
| LLM Provider (OpenAI/Claude/DeepSeek/Ark/Gemini) | AI Agent 推理 | Eino 框架 |
| MySQL | 关系型数据存储 | GORM |
| ClickHouse | 分析型数据存储 | clickhouse-go |
| Redis | 缓存、限流、分布式锁 | go-redis |
| RocketMQ | 异步消息队列 | rocketmq-client-go |
| Object Storage (S3 compatible) | 文件存储 | 自定义 fileserver infra |

### 系统边界划分

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (React)                      │
│  Rush.js Monorepo — 53 projects                            │
│  loop-base / loop-components / loop-pages / loop-modules   │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP/OpenAPI
┌─────────────────────▼───────────────────────────────────────┐
│                    Backend (Go) — CloudWeGo                  │
│  Hertz (HTTP) + Kitex (RPC)                                 │
│  modules/: data / evaluation / foundation / llm / prompt    │
│  infra/: db / redis / mq / ck / fileserver / looptracer     │
└─────────────────────────────────────────────────────────────┘
```

## 组件拆分

### 前端组件

| 组件层 | 包名 | 职责 |
|--------|------|------|
| 基础组件 | `@cozeloop/loop-base` | 账号、API Schema、Logger、Stores、路由、i18n |
| 业务组件 | `@cozeloop/loop-components` | evaluate / observation / prompt 等业务组件 |
| 页面 | `@cozeloop/loop-pages` | auth / evaluate / observation / prompt / tag |
| 应用 | `apps/cozeloop` | 主应用，React 18 + Rsbuild + Tailwind |

### 后端组件

| 模块 | 路径 | 职责 |
|------|------|------|
| data | `modules/data` | 数据管理 |
| evaluation | `modules/evaluation` | AI Agent 评估（Eino + 运行时） |
| foundation | `modules/foundation` | 基础服务 |
| llm | `modules/llm` | LLM 集成（Eino 封装） |
| prompt | `modules/prompt` | Prompt 管理和版本控制 |
| observability | `modules/observability` | 链路追踪、指标监控 |
| api | `api` | Hertz HTTP handlers + routers |
| infra | `infra/*` | DB/Redis/MQ/ClickHouse 等基础设施 |

## 关键数据流

### Agent 开发-评估流程

```
用户操作（前端）
  → Prompt 管理（prompt module）
  → Agent 配置（data module）
  → LLM 调用（llm module via Eino）
  → 执行日志（looptracer）
  → 评估（evaluation module）
  → 结果存储（ClickHouse + MySQL）
```

### API 路由层

- **HTTP**: `backend/api/` — Hertz 框架，处理 OpenAPI 路由
- **RPC**: `kitex_gen/` — Kitex 生成代码，内部服务间通信
- **IDL**: `idl/` — Thrift 定义，API 契约文档化

## 接口约定

| 类别 | 协议 | 说明 |
|------|------|------|
| 外部 API | HTTP/JSON | Hertz + OpenAPI |
| 内部 RPC | Thrift/Kitex | 微服务间调用 |
| 前端→后端 | HTTP REST | `/api/v1/*` 路径 |

## 技术选型

| 领域 | 选型 | 原因 |
|------|------|------|
| 前端框架 | React 18 | 生态成熟，组件化完善 |
| 前端打包 | Rsbuild (Rspack) | 性能优，Rust-based |
| 前端状态 | Zustand | 轻量、支持多 store |
| 后端框架 | CloudWeGo (Hertz + Kitex) | 字节内部开源，高性能 |
| ORM | GORM | Go 生态主流 |
| LLM 框架 | Eino | 支持多 Provider 统一抽象 |
| DI | Google Wire | 编译期依赖注入 |
| 服务发现 | 基于配置 | 未发现服务网格组件 |

## 风险与约束

| 风险/约束 | 说明 |
|-----------|------|
| 双 monorepo 结构 | 前端 Rush.js + 后端 Go modules，文档需分别说明 |
| 缺乏集中文档 | 需新建 `docs/` 体系 |
| Eino 框架定制 | LLM 集成有业务定制逻辑，需说明扩展点 |
| IDL 代码生成 | API 契约通过 Thrift 管理，但前端 ts 类型需同步维护 |
