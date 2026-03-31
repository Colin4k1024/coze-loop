---
name: project-context
description: Coze Loop project overview and current task
type: project
---

# Project Context

## 项目基本信息

| 项 | 内容 |
|---|---|
| 项目名 | Coze Loop — AI Agent 开发与运维平台 |
| 前端 | TypeScript/React, Rush.js monorepo (53 projects), Rsbuild, Tailwind |
| 后端 | Go, CloudWeGo (Hertz + Kitex), GORM, Eino (LLM), RocketMQ, ClickHouse |
| 架构 | 双 monorepo，前端 Rush.js + 后端 Go modules |
| 仓库 | https://github.com/coze-dev/coze-loop |

## 当前任务

- **slug**: `open-platform-login`
- **目标**: 集成 Gitee OAuth 登录能力
- **范围**: Gitee OAuth 登录 + 新用户自动注册（developer 角色）+ 登录日志审计
- **状态**: /team-intake + /team-plan 完成，待 /team-execute

## 新增活跃风险

- Gitee OAuth 集成需要新增 `gitee_id` 字段到 User 模型
- 现有登录日志（login_audit）表和写入逻辑需要新建
- 需确保 OAuth 回调与现有 session middleware 兼容

## 最近发布

| 日期 | 任务 | 状态 |
|------|------|------|
| 2026-03-31 | project-docs-organization | released |

## Tech Stack

### Frontend
- React 18 + TypeScript
- Rsbuild (Rspack-based bundler)
- Tailwind CSS + 自定义 design tokens
- Zustand (状态管理)
- React Router 6
- Rush.js monorepo (53 projects)
- Leveled architecture: loop-base → loop-components → loop-pages → apps

### Backend
- Go 1.24+
- CloudWeGo: Hertz (HTTP), Kitex (RPC)
- GORM + MySQL + ClickHouse
- Redis (缓存/限流/分布式锁)
- RocketMQ (消息队列)
- Eino 框架 (LLM 多 Provider 统一抽象)
- Google Wire (DI)
- OpenTelemetry + looptracer (可观测性)

### Infrastructure
- IDL: Thrift (API 定义)
- Container: Docker, Kubernetes (via release/deployment/)
- File Storage: S3-compatible object storage

## 关键依赖

- Eino 框架的 LLM Provider 抽象（扩展点）
- IDL 代码生成流程（Thrift → Go + TypeScript）
- Rush autoinstaller 和 pnpm workspace

## 活跃风险

- 双 monorepo 结构复杂，文档需分别说明前后端各自的结构
- 缺乏既有 ADR，技术决策历史需追溯代码
- 文档与代码同步需要持续维护机制
