---
artifact: delivery-plan
task: open-platform-login
date: 2026-03-31
role: tech-lead
status: draft
---

# 交付计划

## 版本目标

| 项 | 内容 |
|---|---|
| 版本/里程碑 | v1.0 — Gitee OAuth 登录集成 |
| 范围说明 | Gitee OAuth 登录 + 新用户自动注册 + 登录日志审计 |
| 放行标准 | Gitee 登录流程完整、现有登录不受影响、登录日志入库 |

## 工作拆解

### 阶段 1：架构设计与代码骨架（T+1）

| 工作项 | 主责角色 | 依赖 | 说明 |
|--------|----------|------|------|
| 现有认证模块分析 | backend-engineer | intake 产出 | 分析 user service、session、repo 结构 |
| 新登录日志表设计 | backend-engineer | 现有分析 | 设计 login_audit 表和写入逻辑 |
| OAuth 配置结构设计 | backend-engineer | intake 产出 | 在 foundation.yaml 中预留 gitee 配置 |
| 前端登录页结构分析 | frontend-engineer | intake 产出 | 分析 login-panel 组件和 API 调用模式 |
| 前端 OAuth 跳转设计 | frontend-engineer | 前端分析 | 设计 Gitee 按钮样式和跳转逻辑 |

### 阶段 2：后端实现（T+2）

| 工作项 | 主责角色 | 依赖 | 说明 |
|--------|----------|------|------|
| OAuth 配置加载 | backend-engineer | 阶段 1 | foundation.yaml 预留 gitee 配置 |
| Gitee OAuth 工具函数 | backend-engineer | 配置加载 | authorize URL 生成、token 交换、user info 获取 |
| Gitee 登录 API（/auth/gitee） | backend-engineer | 工具函数 | 返回授权 URL 或处理 callback |
| OAuth callback API | backend-engineer | 工具函数 | 交换 code、创建/查找用户、写 session、写日志 |
| 登录日志服务 | backend-engineer | callback | login_audit 表结构、DAO、写入逻辑 |
| Wire DI 注册 | backend-engineer | 各服务 | 将新组件注入到 DI 容器 |

### 阶段 3：前端实现（T+3）

| 工作项 | 主责角色 | 依赖 | 说明 |
|--------|----------|------|------|
| Gitee 登录按钮 | frontend-engineer | 阶段 1 | 在 login-panel 中新增 Gitee 按钮 |
| OAuth 跳转逻辑 | frontend-engineer | 按钮 | 点击后跳转至 Gitee 授权页 |
| Callback 页面 | frontend-engineer | 后端 API | 接收 callback、显示登录结果 |
| 跳转至首页 | frontend-engineer | callback | 登录成功后 redirect 到首页 |

### 阶段 4：集成与自测（T+4）

| 工作项 | 主责角色 | 依赖 | 说明 |
|--------|----------|------|------|
| 端到端流程测试 | qa-engineer | 前后端完成 | Gitee OAuth 完整流程 |
| 现有登录回归 | qa-engineer | 前后端完成 | 确保原有登录不受影响 |
| 日志写入验证 | backend-engineer | 日志服务 | 验证 login_audit 表数据正确 |

## 风险与缓解

| 风险 | 影响 | 缓解措施 | Owner |
|------|------|----------|-------|
| Gitee API 不稳定 | 中 | 预留重试，记录失败原因 | backend-engineer |
| 新用户字段缺失 | 中 | 确认 User 模型是否需要扩展 provider 字段 | backend-engineer |
| Session 与 OAuth 兼容 | 中 | 复用现有 session middleware，复用 LoginByPassword 模式 | backend-engineer |
| 登录日志表缺失 | 高 | 阶段 1 先设计日志表，阶段 2 实现 | backend-engineer |

## 节点检查

| 节点 | 目标 | 通过标准 |
|------|------|----------|
| 方案评审 | intake + plan 确认范围 | PRD 已更新，所有项已确认 |
| 后端完成 | 阶段 2 结束 | OAuth API 可调通，日志可写入 |
| 前端完成 | 阶段 3 结束 | 登录按钮可见，流程可触发 |
| 测试通过 | 阶段 4 结束 | 端到端通过 + 回归通过 |

## 技能装配清单

| 技能 | 用途 | 主责角色 |
|------|------|----------|
| architect | 架构设计、认证模块分析 | architect |
| backend-engineer | OAuth 实现、日志服务 | backend-engineer |
| frontend-engineer | 登录页改造 | frontend-engineer |
| qa-engineer | 测试方案、回归验证 | qa-engineer |
| tech-lead | 统筹协调 | tech-lead |

**company skill**：不涉及 BPMN/HPRMC/集团脚手架等场景，不启用。

## 前端交付物与检查点

| 交付物 | 说明 |
|--------|------|
| `login-panel` 改造 | 新增 Gitee 登录按钮，符合 Design Tokens |
| OAuth callback 页面 | 独立的 callback 处理页 |
| UI 自测证据 | 按钮样式、响应式、无障碍检查 |

## 关键依赖

- 阶段 1 的分析结果直接影响阶段 2/3 的实现准确性
- Gitee OAuth credentials（由用户提供）需在阶段 2 开始前就位

## 阻塞与升级

- 若 Gitee credentials 未到位，影响阶段 2 → 升级 tech-lead
- 若 User 模型需要扩展字段，影响 API 设计 → 升级 architect
- 若现有 session middleware 不兼容 OAuth callback → backend-engineer 确认方案

## 应用等级 / 技术架构等级 / 关键组件偏离

> 本次为登录功能扩展，不触发生产应用等级变更。
