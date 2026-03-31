---
artifact: prd
task: open-platform-login
date: 2026-03-31
role: tech-lead
status: draft
---

# 需求简报

## 背景

Coze Loop 目前可能仅支持单一登录方式（推测为内部账号体系）。业务上需要扩展登录能力，支持通过第三方开放平台（如 Gitee）进行 OAuth 登录，满足用户多样化的登录需求。

## 目标

集成 Gitee OAuth 登录能力，作为现有登录体系的补充。

## 约束

1. **保持现有前后端交付体系**：不改变现有 `/team-*` 工作流、文档结构和质量门禁
2. **保持现有 UI 规范**：新增登录入口需符合既有设计系统和前端规范（Tailwind + Design Tokens）
3. **配置外置**：OAuth credentials 由用户提供，存储在后端配置中，不硬编码

## 用户故事

| # | 用户故事 | 验收标准 |
|---|----------|----------|
| 1 | 用户使用 Gitee 账号登录 | 用户可通过 Gitee OAuth 完成登录，进入首页 |
| 2 | 新用户首次通过 Gitee 登录时自动创建账号 | 首次 OAuth 登录时，后端自动创建用户记录，默认角色为 developer（拥有所有权限） |
| 3 | Gitee 登录与现有账号体系共存 | 原有的用户名密码登录方式不受影响 |
| 4 | 登出行为一致 | Gitee 登录用户登出时行为与原有账号一致 |
| 5 | 登录后跳转回首页 | OAuth 登录成功后统一跳转至首页 |

## 范围

### In Scope
- Gitee OAuth 登录集成（后端 + 前端）
- 新用户自动注册，默认角色 developer（全部权限）
- 登录态与会话管理
- 前端登录 UI 改造（新增 Gitee 登录入口）
- 登录日志与审计体系接入

### Out of Scope
- 其他 OAuth 平台（留作后续扩展）
- 账号绑定（一个 Gitee 账号对应一个新账号，不做绑定）
- 第三方账号头像、昵称同步（首期不做）

## 配置预留

后端需预留以下配置项位置（credentials 由用户提供）：

```yaml
# 后端配置文件
oauth:
  gitee:
    enabled: true
    client_id: "${GITEE_CLIENT_ID}"
    client_secret: "${GITEE_CLIENT_SECRET}"
    redirect_uri: "/api/v1/oauth/gitee/callback"
```

## 风险与依赖

| 风险 | 影响 | 缓解 |
|------|------|------|
| OAuth 安全风险 | 高 | 使用标准 OAuth2.0，确保 client_secret 不暴露在前端 |
| 用户冲突（同一邮箱） | 中 | 首次登录自动创建，不做邮箱合并 |
| Gitee API 限流 | 低 | 登录流程仅需 user info，预留重试机制 |
| 登录日志接入 | 中 | 需确认现有日志表结构和字段含义 |

---

# 参与角色清单

| 角色 | 职责 | 输入缺口 |
|------|------|----------|
| tech-lead | 统筹协调、技术方案确认 | 已确认范围和角色 |
| architect | 认证架构设计、现有模块结构分析 | 需要分析现有认证/用户模块 |
| frontend-engineer | 前端登录页 UI 改造 | 需要确认设计系统约束 |
| backend-engineer | OAuth 流程实现、用户注册、日志接入 | 需要分析现有用户模块和日志表 |
| qa-engineer | 测试方案设计 | — |

---

# 待确认项

- [x] Gitee OAuth credentials — **用户（你）提供，后端预留配置位置**
- [x] 平台范围 — **仅 Gitee，不做扩展**
- [x] 默认角色 — **developer，拥有所有权限**
- [x] 登录后跳转 — **跳转至首页**
- [x] 登录日志/审计 — **接入现有登录日志体系**

---

# 企业治理待确认项

> 注：本次为平台功能扩展，不涉及企业内部应用等级判定。

---

# UI 范围、终端假设与质量门禁

| 维度 | 内容 |
|------|------|
| 目标端 | Web |
| 产品类型 | 平台产品 — 登录页 |
| 关键页面 | 登录页（新增 Gitee 入口）、OAuth 回调页 |
| 设计约束 | 必须遵循现有 Tailwind + Design Tokens 体系 |
| 响应式基线 | 移动端适配 |
| 无障碍 | 键盘可访问、屏幕阅读器可理解 |
| 性能基线 | 登录流程 ≤ 3 秒（不含网络延迟） |

---

# 下一步

1. 启动 `/team-plan` 拆解任务
2. architect 分析现有认证/用户模块结构
3. backend-engineer 分析现有登录日志表结构
4. frontend-engineer 确认登录页设计系统约束
