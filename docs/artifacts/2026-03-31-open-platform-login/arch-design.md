---
artifact: arch-design
task: open-platform-login
date: 2026-03-31
role: architect
status: draft
---

# 架构设计 — Gitee OAuth 登录集成

> 本文档基于代码探索初步产出，详细实现细节在阶段 1 分析后补充。

## 系统边界

### 新增外部依赖

| 依赖 | 用途 | 集成方式 |
|------|------|----------|
| Gitee OAuth API | 第三方登录授权 | HTTPS REST API |

### 新增组件

| 组件 | 类型 | 职责 |
|------|------|------|
| Gitee OAuth Handler | API Handler | OAuth 授权跳转、callback 处理 |
| Gitee OAuth Service | Domain Service | authorize URL 生成、token 交换、user info 获取 |
| Login Audit Service | Domain Service | 登录日志写入 |
| login_audit 表 | 数据库表 | 存储登录日志 |

### 现有组件（复用）

| 组件 | 复用方式 |
|------|----------|
| User Entity / GORM Model | 不变，直接用唯一标识（gitee_id）关联或新建用户 |
| Session Middleware | 复用，登录成功后下发 session cookie |
| User Repo | 复用，新增按 gitee_id 查询接口 |
| Password Hashing (Argon2id) | 复用，但 Gitee 用户无密码 |

## 数据流

### Gitee OAuth 登录完整流程

```
1. 用户点击"使用 Gitee 登录"
   → 前端跳转 GET /api/foundation/v1/oauth/gitee/authorize
   → 后端生成 state 参数，存储重定向 URL
   → 后端拼接 Gitee authorize URL，返回给前端
   → 前端 window.location.href = authorizeURL

2. 用户在 Gitee 授权页面完成授权
   → Gitee 回调至 /api/foundation/v1/oauth/gitee/callback?code=xxx&state=yyy
   → 前端中转（可选）或直接由后端接收

3. 后端处理 callback
   → 验证 state 防止 CSRF
   → 用 code 换取 access_token（Gitee API）
   → 用 access_token 获取用户信息（Gitee API）
   → 查询用户：gitee_id 已注册？ 是 → 查到已有用户；否 → 创建新用户（developer 角色）
   → 写入 session cookie（复用 LoginByPassword 模式）
   → 写入 login_audit 日志
   → 返回重定向至首页

4. 前端接收 session cookie
   → 前端 redirect 至首页
   → 用户已登录状态
```

## 接口约定

### 新增 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/foundation/v1/oauth/gitee/authorize` | 返回 Gitee 授权 URL |
| GET | `/api/foundation/v1/oauth/gitee/callback` | Gitee OAuth 回调 |

### 现有 API（不变）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/foundation/v1/users/login_by_password` | 现有密码登录（不变） |
| POST | `/api/foundation/v1/users/logout` | 登出（不变） |
| GET | `/api/foundation/v1/users/session` | 获取当前用户（不变） |

## 数据模型

### User 模型扩展（待确认）

现有 `User` GORM model 字段：
```
ID, Name, UniqueName, Email, Password, Description, IconURI, UserVerified, CountryCode, SessionKey, DeletedAt, CreatedAt, UpdatedAt
```

**方案 A（推荐）**：新增 `gitee_id` 唯一索引字段
```
gitee_id  string  // varchar(64)，唯一索引，可为空
```
新用户注册时若通过 Gitee，gitee_id 填入 Gitee 用户 ID。

**方案 B**：新建 `user_oauth` 关联表
```
user_oauth: user_id, provider(gitee), provider_user_id, created_at
```
适合多平台扩展，但更复杂。

**推荐方案 A**，本次仅 Gitee，够用且简单。

### login_audit 表

```sql
CREATE TABLE login_audit (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id       BIGINT NOT NULL COMMENT '用户ID，-1 表示未注册',
    login_type    TINYINT NOT NULL COMMENT '1=密码登录 2=Gitee OAuth',
    provider      VARCHAR(32) DEFAULT NULL COMMENT 'provider: gitee',
    ip            VARCHAR(64) COMMENT '登录 IP',
    user_agent    VARCHAR(512) COMMENT '浏览器 UA',
    success       TINYINT NOT NULL COMMENT '1=成功 0=失败',
    fail_reason   VARCHAR(256) DEFAULT NULL COMMENT '失败原因',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
);
```

## 配置结构

### foundation.yaml 新增配置

```yaml
oauth:
  gitee:
    enabled: true
    client_id: "${GITEE_CLIENT_ID}"
    client_secret: "${GITEE_CLIENT_SECRET}"
    redirect_uri: "/api/foundation/v1/oauth/gitee/callback"
    scopes:
      - "user_info"
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `GITEE_CLIENT_ID` | Gitee OAuth App Client ID |
| `GITEE_CLIENT_SECRET` | Gitee OAuth App Client Secret |

## 技术选型

| 项 | 选型 | 说明 |
|---|------|------|
| OAuth 库 | 标准 golang.org/x/oauth2 + 自定义 Gitee provider | 不引入重型依赖 |
| Session | 复用现有 HMAC-SHA256 session | Cookie 方式，与现有登录一致 |
| 用户创建 | 新建用户，角色 developer | 自动注册 |
| 日志写入 | 同步写入 login_audit | 在 session 下发成功后写入 |

## 关键设计决策

1. **state 参数**：使用 HMAC-SHA256 签名（复用 session key），防止 CSRF
2. **Gitee ID 存储**：在 User 表新增 gitee_id 字段，避免新建关联表
3. **登录日志**：独立 login_audit 表，不复用现有 audit service（因为现有 audit service 是业务审计，不适合登录日志）
4. **错误处理**：Gitee API 失败时返回 500，前端展示通用错误，不泄漏内部细节

## 风险与约束

| 风险 | 说明 |
|------|------|
| Gitee API 限流 | 登录流程仅 2 次 API 调用（token + userinfo），预留重试 |
| CSRF | 通过 state 参数 HMAC 签名防止 |
| 用户首次登录无账号 | 自动创建，角色 developer |
| 现有 session 机制不变 | 复用 LoginByPassword 成功后的 session 下发模式 |
