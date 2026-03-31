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

- **slug**: `project-docs-organization`
- **目标**: 梳理整个项目的设计文档体系（前端 + 后端 + 使用文档）
- **范围**: docs/ 目录结构、架构文档、使用指南；不做代码修改
- **状态**: 已完成（/team-intake → /team-plan → /team-execute → /team-review → /team-release）

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
