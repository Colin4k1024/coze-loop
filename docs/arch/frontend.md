# Coze Loop 前端架构文档

## 1. 项目概述

Coze Loop 前端是一个基于 **React 18** + **Rsbuild** + **Tailwind CSS** + **Rush Monorepo** 技术栈的企业级前端应用。前端代码位于 `/frontend` 目录下，采用 **pnpm workspace** 管理多包依赖。

### 1.1 技术栈概览

| 类别 | 技术选型 | 版本 |
|------|----------|------|
| UI 框架 | React | 18.2.0 |
| 构建工具 | Rsbuild | ~1.1.0 |
| CSS 框架 | Tailwind CSS | ~3.3.3 |
| 状态管理 | Zustand | ^4.4.7 |
| 路由 | React Router | v6.22.0 |
| Hooks 库 | ahooks | ^3.7.8 |
| 国际化 | i18next / intl-messageformat | >=19.0.0 |
| 包管理器 | pnpm | 8.15.8 |
| Monorepo 工具 | Rush.js | 5.147.1 |
| 测试框架 | Vitest | ~3.0.5 |
| 代码规范 | ESLint + Prettier + Stylelint | - |

### 1.2 目录结构

```
frontend/
├── apps/
│   └── cozeloop/              # 主应用入口
├── packages/
│   ├── loop-base/             # 基础能力包
│   ├── loop-components/        # 业务组件包
│   ├── loop-config/           # 配置包
│   ├── loop-modules/          # 业务模块包
│   └── loop-pages/            # 页面包
├── config/                    # 共享配置
│   ├── eslint-config/
│   ├── postcss-config/
│   ├── stylelint-config/
│   ├── tailwind-config/
│   └── ts-config/
├── infra/                     # 基础设施
│   ├── eslint-plugin/
│   ├── idl/
│   ├── plugins/
│   └── utils/
└── scripts/                   # 构建脚本
```

---

## 2. Monorepo 结构

Coze Loop 前端采用 **Rush.js** 进行 Monorepo 管理。项目依赖通过 `workspace:*` 协议引用本地包，而非发布到 npm 的版本。

### 2.1 包层级划分

| 层级 | 包范围 | 说明 |
|------|--------|------|
| **Level 1** | `config/*`, `infra/*` | 共享配置和基础设施 |
| **Level 2** | `packages/loop-base/*` | 基础能力（hooks、stores、i18n、路由、权限等） |
| **Level 3** | `packages/loop-components/*` | 业务组件（Prompt/Evaluation/Observation 组件） |
| **Level 4** | `packages/loop-pages/*` | 页面级组件（Auth/Eval/Observation/Prompt/Tag Pages） |
| **Level 5** | `apps/cozeloop` | 主应用入口 |

### 2.2 主要 Package 清单

#### 2.2.1 loop-base（基础能力层）

`packages/loop-base/` 下包含 22 个子包，提供前端应用的基础能力：

| 包名 | 说明 |
|------|------|
| `@cozeloop/account` | 账户相关状态和逻辑 |
| `@cozeloop/api-schema` | API Schema 定义（基于 IDL 自动生成） |
| `@cozeloop/base-hooks` | 基础 React Hooks |
| `@cozeloop/bot-env` | Bot 环境配置 |
| `@cozeloop/bot-env-adapter` | Bot 环境适配器 |
| `@cozeloop/bot-flags` | 特性开关/Feature Flags |
| `@cozeloop/bot-md-box-adapter` | Markdown Box 适配器 |
| `@cozeloop/bot-typings` | Bot 类型定义 |
| `@cozeloop/components` | 通用 UI 组件库 |
| `@cozeloop/env-adapter` | 环境适配器 |
| `@cozeloop/fetch-stream` | 流式请求处理 |
| `@cozeloop/guard` | 权限守卫组件 |
| `@cozeloop/i18n` | 国际化资源管理 |
| `@cozeloop/intl` | 运行时国际化（基于 i18next） |
| `@cozeloop/logger` | 日志模块 |
| `@cozeloop/loop-lng` | 语言配置 |
| `@cozeloop/route-base` | 路由工具 |
| `@cozeloop/stores` | Zustand 状态管理 |
| `@cozeloop/tea-adapter` | TEA 埋点适配器 |
| `@cozeloop/toolkit` | 通用工具函数 |

#### 2.2.2 loop-components（业务组件层）

`packages/loop-components/` 下包含 13 个子包：

| 包名 | 说明 |
|------|------|
| `@cozeloop/adapter-interfaces` | 业务适配器接口定义 |
| `@cozeloop/biz-components` | 业务通用组件 |
| `@cozeloop/biz-config` | 业务配置 |
| `@cozeloop/biz-hooks` | 业务 Hooks |
| `@cozeloop/components-with-adapter` | 带适配器的组件封装 |
| `@cozeloop/evaluate-adapter` | 评测模块适配器 |
| `@cozeloop/evaluate-components` | 评测相关组件 |
| `@cozeloop/observation-adapter` | 观测模块适配器 |
| `@cozeloop/observation-components` | 观测相关组件 |
| `@cozeloop/prompt-components` | Prompt 相关组件（v1） |
| `@cozeloop/prompt-components-v2` | Prompt 相关组件（v2） |
| `@cozeloop/shared-components` | 共享业务组件 |
| `@cozeloop/tag-components` | 标签相关组件 |

#### 2.2.3 loop-pages（页面层）

`packages/loop-pages/` 下包含 5 个子包：

| 包名 | 说明 |
|------|------|
| `@cozeloop/auth-pages` | 登录/鉴权页面 |
| `@cozeloop/evaluate-pages` | 评测页面 |
| `@cozeloop/observation-pages` | 观测页面 |
| `@cozeloop/prompt-pages` | Prompt 管理页面 |
| `@cozeloop/tag-pages` | 标签管理页面 |

#### 2.2.4 loop-modules（业务模块层）

`packages/loop-modules/` 下包含：

| 包名 | 说明 |
|------|------|
| `@cozeloop/evaluate` | 评测业务模块（聚合包） |

#### 2.2.5 loop-config（配置层）

| 包名 | 说明 |
|------|------|
| `@cozeloop/rsbuild-config` | Rsbuild 构建配置 |
| `@cozeloop/tailwind-config` | Tailwind CSS 配置 |
| `@cozeloop/tailwind-plugin` | Tailwind 插件 |

### 2.3 工作空间依赖关系

依赖方向遵循：**config/infra -> loop-base -> loop-components -> loop-pages -> apps/cozeloop**

各包通过 `workspace:*` 协议引用本地依赖，例如：

```json
{
  "dependencies": {
    "@cozeloop/components": "workspace:*",
    "@cozeloop/stores": "workspace:*",
    "@cozeloop/i18n-adapter": "workspace:*"
  }
}
```

---

## 3. 主应用入口（apps/cozeloop）

### 3.1 应用结构

```
apps/cozeloop/
├── src/
│   ├── main.tsx           # 入口文件（初始化逻辑）
│   ├── app.tsx            # 根组件
│   ├── index.tsx          # React DOM 渲染
│   ├── index.css          # 全局样式
│   ├── global.d.ts        # 全局类型声明
│   ├── assets/            # 静态资源（fonts、images）
│   ├── components/        # 主应用特有组件
│   │   ├── basic-layout/ # 基础布局
│   │   ├── breadcrumb/   # 面包屑
│   │   ├── locale-provider/
│   │   ├── navbar/       # 导航栏
│   │   └── user-info-section/
│   ├── constants/         # 常量定义
│   ├── hooks/             # 主应用 Hooks
│   │   ├── use-api-error-toast.tsx
│   │   ├── use-setup-i18n.ts
│   │   └── use-setup-space.ts
│   └── routes/            # 路由配置
│       ├── base-route.tsx     # 基础路由（鉴权）
│       ├── space-route.tsx    # Space 路由
│       ├── enterprise-route.tsx # 企业路由
│       └── index.tsx          # 路由配置汇总
├── rsbuild.config.ts      # Rsbuild 配置
├── vitest.config.ts       # Vitest 配置
├── tsconfig.json          # TypeScript 配置
└── package.json
```

### 3.2 渲染流程

```
index.html (#cozeloop-root)
    ↓
main.tsx (render 函数)
    ├── initIntl() - 初始化国际化
    ├── pullFeatureFlags() - 拉取特性开关
    └── dynamicImportMdBoxStyle() - 动态导入 Markdown 样式
        ↓
    app.tsx (App 组件)
        ├── useSetupI18n() - 设置 i18n
        ├── CozeLoopProvider - 全局 Provider
        ├── Suspense - 路由懒加载
        ├── LocaleProvider - 语言环境
        └── RouterProvider - 路由管理
```

### 3.3 Rsbuild 配置

主应用使用 `@cozeloop/rsbuild-config` 包提供的 `createRsbuildConfig` 函数创建配置：

```typescript
// apps/cozeloop/rsbuild.config.ts
export default createRsbuildConfig({
  server: { port: 8090 },
  dev: { assetPrefix: `http://localhost:8090` },
  html: {
    title: 'Coze Loop',
    template: './src/assets/template.html',
  },
});
```

**关键配置项：**

- **port**: 开发服务器端口 8090
- **CSS 预处理器**: 支持 LESS 和 SASS
- **Tailwind**: 使用 `@tailwindcss/nesting` 和 `@tailwindcss`
- **Chunk 分割策略**: `split-by-experience`，包括 semiStyles、cozeDesign、semiUI、mathjax 等分组
- **alias**: 对 react、react-dom、react-router-dom 等核心库做了路径别名

---

## 4. 设计系统

### 4.1 Tailwind 配置

Tailwind 配置位于 `config/tailwind-config/` 包中，提供了完整的 Coze Design Token 映射。

**配置特点：**

- **Dark Mode**: 使用 `class` 模式，通过 `.dark` 类切换
- **CSS 变量**: 基于 `rgba(var(--xxx), alpha)` 格式的颜色系统
- **语义化颜色**: foreground、background、brand、red、yellow、green 等

### 4.2 Design Token 结构

```javascript
// config/tailwind-config/src/index.js
module.exports = {
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        foreground: { DEFAULT: 'rgba(var(--foreground), 1)', ... },
        background: { DEFAULT: 'rgba(var(--background), 1)', ... },
        brand: { DEFAULT: 'rgba(var(--coze-brand-7), 1)', ... },
        // ... 更多颜色
      },
      spacing: { DEFAULT: 'var(--coze-8)', ... },
      borderRadius: { DEFAULT: 'var(--coze-8)', ... },
      fontSize: { mini: 'var(--coze-10)', base: 'var(--coze-12)', ... },
      // ... 更多 token
    },
  },
};
```

### 4.3 设计 Token 转换

`design-token.ts` 提供了将 JSON 格式 design token 转换为 Tailwind 配置的工具函数：

- `designTokenToTailwindConfig()`: 转换颜色、间距、圆角
- `getTailwindContents()`: 扫描所有 React 包，生成 Tailwind content 路径

---

## 5. 路由系统

### 5.1 路由架构

使用 **React Router v6** 的 `createBrowserRouter` 和嵌套路由。

### 5.2 路由层级

```
/                           # 根路径
├── /auth/*                 # 鉴权路由（登录页等）
└── /console                # 主控制台
    ├── /console/           # 默认跳转
    │   └── enterprise/     # 企业路由（重定向）
    │       └── /:enterpriseID  # 企业 ID
    │           └── /space      # Space 路由（重定向）
    │               └── /:spaceID
    │                   ├── /pe/*          # Prompt Engineering
    │                   ├── /evaluation/*   # 评测
    │                   ├── /observation/*  # 观测
    │                   └── /tag/*         # 标签
```

### 5.3 路由守卫

| 守卫 | 文件 | 职责 |
|------|------|------|
| `BaseRoute` | `base-route.tsx` | 检查登录状态，未登录重定向到 `/auth` |
| `SpaceRoute` | `space-route.tsx` | 检查 Space 有效性 |
| `EnterpriseRoute` | `enterprise-route.tsx` | 处理企业级路由跳转 |
| `GuardProvider` | `@cozeloop/guard` | 权限控制上下文 |

### 5.4 路由配置示例

```typescript
// apps/cozeloop/src/routes/index.tsx
export const routeConfig: RouteObject[] = [
  {
    path: '/auth/*',
    element: <Auth />,  // 懒加载
  },
  {
    path: '/',
    element: <BaseRoute />,
    children: [
      {
        path: 'console',
        element: <Outlet />,
        children: [
          {
            path: 'enterprise/:enterpriseID',
            element: <BasicLayout />,
            children: [
              {
                path: 'space/:spaceID',
                children: [
                  { path: 'pe/*', element: <Prompt /> },
                  { path: 'evaluation/*', element: <Evaluation /> },
                  { path: 'observation/*', element: <Observation /> },
                  { path: 'tag/*', element: <Tag /> },
                ],
              },
            ],
          },
        ],
      },
    ],
  },
];
```

---

## 6. 状态管理

### 6.1 Zustand Stores

状态管理使用 **Zustand**，Store 定义位于 `@cozeloop/stores` 包。

**导出的 Store：**

| Store | 导出 | 说明 |
|-------|------|------|
| `useUIStore` | UI 状态管理 | 包含 BreadcrumbItemConfig |
| `useI18nStore` | 国际化状态 | 语言切换 |
| `useEvaluationFlagStore` | 评测功能开关 | - |

### 6.2 Store 结构

```typescript
// packages/loop-base/stores/src/index.ts
export { useUIStore, UIEvent } from './stores/ui';
export type { BreadcrumbItemConfig } from './stores/ui';
export { useI18nStore } from './stores/i18n';
export type { I18nLang } from './stores/i18n';
export { useEvaluationFlagStore, EvaluationFlagEvent } from './stores/evaluation-flag-store';
```

### 6.3 状态流

```
@cozeloop/stores (Zustand)
    ↓
@cozeloop/hooks (业务 Hooks)
    ↓
页面/组件 (React)
```

---

## 7. 国际化（i18n）

### 7.1 i18n 架构

国际化采用 **i18next** + **intl-messageformat** 方案。

**关键包：**

| 包 | 说明 |
|----|------|
| `@cozeloop/intl` | 运行时国际化（i18next） |
| `@cozeloop/i18n` | 国际化资源管理 |
| `@cozeloop/loop-lng` | 语言配置 |
| `@cozeloop/i18n-adapter` | i18n 适配器（封装 intl） |

### 7.2 初始化流程

```typescript
// apps/cozeloop/src/main.tsx
await initIntl({
  fallbackLng: ['zh-CN', 'en-US'],
});
```

### 7.3 组件使用

```typescript
import { I18n } from '@cozeloop/i18n-adapter';

// 在组件中使用
<I18n>{(t) => t('some_key')}</I18n>

// 或使用 hook
const { t } = useTranslation();
```

---

## 8. 权限系统

### 8.1 Guard 架构

权限管理通过 `@cozeloop/guard` 包实现，提供路由级和组件级权限控制。

### 8.2 核心概念

| 概念 | 说明 |
|------|------|
| `GuardPoint` | 权限点枚举（如 `pe.prompts.create`） |
| `Guard` | 组件级权限控制 |
| `GuardRoute` | 路由级权限控制 |
| `GuardProvider` | 权限上下文 Provider |

### 8.3 使用示例

```tsx
// 组件级
<Guard point={GuardPoint['pe.prompts.create']}>
  <button>创建 Prompt</button>
</Guard>

// 路由级
<GuardRoute point={GuardPoint['pe.prompts.create']}>
  <ProtectedContent />
</GuardRoute>
```

---

## 9. API Schema 管理

### 9.1 IDL 工作流

API 类型定义通过 IDL（Interface Definition Language）自动生成：

```
后端 IDL -> @cozeloop/api-schema (自动生成)
    ↓
各业务包引用
```

### 9.2 Schema 包结构

```typescript
// packages/loop-base/api-schema 的导出
export * from './observation';  // 观测相关 API
export * from './evaluation';   // 评测相关 API
export * from './data';        // 数据相关 API
export * from './llm-manage';  // LLM 管理 API
export * from './foundation';   // 基础 API
export * from './prompt';       // Prompt 相关 API
```

### 9.3 更新 API

```bash
# 在 api-schema 包目录下
rushx update
```

---

## 10. 适配器模式

Coze Loop 前端大量使用 **适配器模式** 来解耦业务逻辑。

### 10.1 适配器接口

`@cozeloop/adapter-interfaces` 定义了各业务模块的接口：

```typescript
export * from './evaluate';   // 评测适配器接口
export * from './prompt';     // Prompt 适配器接口
export * from './observation'; // 观测适配器接口
```

### 10.2 适配器实现

| 适配器 | 实现包 |
|--------|--------|
| Evaluate Adapter | `@cozeloop/evaluate-adapter` |
| Observation Adapter | `@cozeloop/observation-adapter` |
| Biz Hooks Adapter | `@cozeloop/biz-hooks-adapter` |
| Biz Components Adapter | `@cozeloop/biz-components-adapter` |
| Biz Config Adapter | `@cozeloop/biz-config-adapter` |
| Tea Adapter | `@cozeloop/tea-adapter` |

### 10.3 组件封装

`@cozeloop/components-with-adapter` 提供了带适配器封装的组件，便于不同实现间的切换。

---

## 11. 构建系统

### 11.1 Rush 工作流

```bash
# 安装依赖
rush update

# 运行开发服务器
cd apps/cozeloop && rushx dev

# 构建生产版本
cd apps/cozeloop && rushx build

# 更新 API Schema
rushx update-api
```

### 11.2 Rsbuild 配置

- **开发模式**: HMR + Source Map
- **生产模式**: 代码分割 + 哈希命名 + 优化压缩
- **CSS**: PostCSS 处理 Tailwind + autoprefixer

### 11.3 代码分割策略

```javascript
chunkSplit: {
  strategy: 'split-by-experience',
  cacheGroups: {
    semiStyles: { name: 'semi', test: /node_modules\/.*semi.*\.(css|less)/ },
    cozeDesign: { name: 'lib-coze-design', test: /coze-design/ },
    semiUI: { name: 'lib-semi-ui', test: /@douyinfe\/semi-ui/ },
    mathjax: { name: 'lib-mathjax', test: /mathjax-full/ },
  },
}
```

---

## 12. 测试

### 12.1 测试框架

- **Vitest**: 单元测试框架
- **@vitest/coverage-v8**: 代码覆盖率
- **@testing-library/react**: React 组件测试
- **happy-dom**: DOM 模拟

### 12.2 测试命令

```bash
# 运行测试
rushx test

# 带覆盖率
rushx test:cov
```

### 12.3 测试配置

Vitest 配置位于各包的 `vitest.config.ts`，或在 `config/vitest-config/` 包中共享。

---

## 13. 代码规范

### 13.1 ESLint

- **@coze-arch/eslint-config**: 共享 ESLint 配置
- 支持 React + TypeScript
- 路径规范：`rules/import.js`, `rules/ts-standard.js`

### 13.2 Prettier

统一代码格式化，配置见 `.prettierrc.js`

### 13.3 Stylelint

CSS/LESS/SCSS 样式规范，配置在 `config/stylelint-config/`

---

## 14. 关键依赖图

```
apps/cozeloop
├── @cozeloop/auth-pages
├── @cozeloop/evaluate-pages
├── @cozeloop/observation-pages
├── @cozeloop/prompt-pages
├── @cozeloop/tag-pages
├── @cozeloop/components
├── @cozeloop/stores
├── @cozeloop/hooks
├── @cozeloop/account
├── @cozeloop/guard
├── @cozeloop/i18n-adapter
├── @coze-arch/bot-flags
└── zustand, react-router-dom, ahooks, etc.
```

---

## 15. 未确认项

以下信息无法从代码中确认，需要询问项目成员：

1. **rush.json 位置**: 根目录未找到 `rush.json`，可能使用其他方式管理 Rush 配置
2. **CI/CD 配置**: 流水线配置未在代码库中明确展示
3. **Storybook**: 虽然 `observation-components` 包含 Storybook 配置，但未确认是否所有组件都有 Storybook
4. **具体业务组件列表**: 部分包的详细组件导出未逐一确认
