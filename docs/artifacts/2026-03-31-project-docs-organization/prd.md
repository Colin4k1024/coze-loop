---
artifact: prd
task: project-docs-organization
date: 2026-03-31
role: tech-lead
status: draft
---

# 需求简报

## 背景

Coze Loop 是 AI Agent 开发与运维平台，采用双 monorepo 结构（前端 Rush.js + 后端 Go），目前缺乏集中归档的设计文档体系。现有架构信息散落在 README、Wiki、代码注释中，没有统一的 `docs/` 目录和 ADR 记录。

## 目标

梳理整个项目的设计文档体系，包括：
1. 前端架构与模块文档
2. 后端架构与模块文档
3. 平台使用文档
4. 建立 docs/ 目录结构，支持后续 ADR 归档

## 范围

### In Scope
- 前端：模块划分、设计系统、组件架构、页面路由
- 后端：模块划分、API 架构、基础设施层（DB/Redis/MQ/ClickHouse）、LLM 集成
- 整体：技术栈概览、部署架构、基础设施
- 文档结构：`docs/` 目录规划

### Out of Scope
- 代码修改
- 代码层面的重构或优化
- 新功能开发

## 关键约束

- 仅涉及文档梳理和创建，不改动任何业务代码
- 文档需与现有代码结构保持一致

---

# 参与角色清单

| 角色 | 职责 | 输入缺口 |
|------|------|----------|
| tech-lead | 统筹协调、文档结构规划 | 需要确认文档输出范围和优先级 |
| architect | 输出架构设计文档（前后端） | 需要深入理解模块边界和依赖关系 |
| frontend-engineer | 前端文档梳理与编写 | 需要提供组件/页面结构和设计系统说明 |
| backend-engineer | 后端文档梳理与编写 | 需要提供模块划分、API 契约、数据流说明 |

---

# 待确认项

- [ ] 文档优先级：先做架构总览还是先按模块深挖？
- [ ] 是否需要建立 ADR 记录已有技术决策（如 LLM 框架选型 Eino、RPC 选型 Kitex）？
- [ ] 前端设计系统文档是否需要截图或色板等视觉辅助？
- [ ] docs/ 目录结构是否需要与现有 `release/`、`conf/` 目录做关联索引？

---

# 企业治理待确认项

- [ ] 本项目是否属于企业内部应用？如是，应用等级（T1-T4）为何？
- [ ] 是否需要接入集团统一监控/日志/配置等公共能力？

> 注：本任务为文档梳理性质，不涉及代码发布，暂不触发企业内控强制约束。

---

# 领域技能包启用建议

本次任务不涉及 BPMN、HPRMC、海尔脚手架、GitLab 发布、Langfuse 追踪、DDD 业务服务配置等场景，不启用领域扩展技能包。

---

# UI 范围、终端假设与质量门禁

| 维度 | 状态 |
|------|------|
| 是否涉及 UI 变更 | 否 |
| 目标端 | 纯文档 |
| 产品类型 | 平台产品文档 |
| 体验门禁 | 不适用（无 UI 变更） |
| 性能基线 | 不适用（无 UI 变更） |

---

# 初步 docs/ 目录结构建议

```
docs/
├── README.md                     # 文档门户索引
├── artifacts/                    # 任务级产出（本次重点）
│   └── 2026-03-31-*/            # 本次任务产出
├── arch/                         # 架构文档
│   ├── README.md                 # 架构总览
│   ├── frontend.md               # 前端架构
│   ├── backend.md                # 后端架构
│   ├── infrastructure.md         # 基础设施（DB/Redis/MQ/ClickHouse）
│   └── deployment.md             # 部署架构
├── adr/                          # Architecture Decision Records
│   └── README.md                 # ADR 索引
└── guides/                       # 使用指南
    ├── README.md
    ├── frontend-dev.md           # 前端开发指南
    ├── backend-dev.md            # 后端开发指南
    └── deployment.md             # 部署指南
```

---

# 下一步

1. 确认文档优先级和范围
2. 启动 `/team-plan` 拆解具体工作任务
3. 由 architect、frontend-engineer、backend-engineer 分别输出各自负责的文档
