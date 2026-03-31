# Coze Loop 前端开发指南

## 1. 开发环境准备

### 1.1 环境要求

| 工具 | 版本要求 |
|------|----------|
| Node.js | 18+ (推荐 lts/iron) |
| pnpm | 8.15.8 |
| Rush | 5.147.1 |

### 1.2 安装步骤

#### 1. 安装 Node.js 18+

```bash
# 使用 nvm 安装
nvm install lts/iron
nvm alias default lts/iron
nvm use lts/iron
```

#### 2. 全局安装 pnpm 和 Rush

```bash
npm i -g pnpm@8.15.8 @microsoft/rush@5.147.1
```

#### 3. 克隆代码并安装依赖

```bash
# 克隆仓库
git clone git@github.com:coze-dev/coze-loop.git

# 进入前端目录
cd frontend

# 安装依赖
rush update
```

> **提示**: `rush update` 会自动安装所有 workspace 成员的依赖，耗时可能较长。

### 1.3 启动开发服务器

```bash
cd apps/cozeloop
rushx dev
```

访问 http://localhost:8090/ 查看应用。

**其他开发命令:**

```bash
# 中国区开发模式（离线构建）
rushx dev:cn-boe

# 中国区 Release 模式
rushx dev:cn-release
```

---

## 2. Monorepo 命令

### 2.1 Rush 常用命令

| 命令 | 说明 |
|------|------|
| `rush update` | 更新 lockfile，安装所有依赖 |
| `rush build` | 构建所有包（按拓扑顺序） |
| `rushx <cmd>` | 在当前包运行命令 |
| `rush update-api` | 从 IDL 更新 API Schema |

### 2.2 pnpm workspace 命令

由于使用 pnpm workspace，也可以直接使用 pnpm 命令：

```bash
# 在根目录运行
pnpm --filter @cozeloop/components build

# 在包目录运行
cd packages/loop-base/components
pnpm dev
```

### 2.3 包内命令

各包通用的 npm scripts：

| 命令 | 说明 |
|------|------|
| `rushx dev` | 开发模式 |
| `rushx build` | 生产构建 |
| `rushx build:prod` | 生产构建（区分 region） |
| `rushx lint` | ESLint 检查 |
| `rushx test` | 运行测试 |
| `rushx test:cov` | 运行测试并生成覆盖率 |

---

## 3. 项目结构

### 3.1 目录结构

```
frontend/
├── apps/
│   └── cozeloop/              # 主应用
├── packages/
│   ├── loop-base/             # 基础能力
│   │   ├── account/
│   │   ├── api-schema/
│   │   ├── components/
│   │   ├── guard/
│   │   ├── hooks/
│   │   ├── i18n/
│   │   ├── intl/
│   │   ├── route/
│   │   ├── stores/
│   │   └── toolkit/
│   ├── loop-components/        # 业务组件
│   ├── loop-config/           # 配置
│   ├── loop-modules/           # 业务模块
│   └── loop-pages/             # 页面
├── config/                    # 共享配置
│   ├── eslint-config/
│   ├── postcss-config/
│   ├── tailwind-config/
│   └── ts-config/
└── infra/                     # 基础设施
```

### 3.2 包命名规范

- **@cozeloop/***: Coze Loop 业务包
- **@coze-arch/***: 共享架构配置包
- **workspace:***: 本地包引用协议

---

## 4. 添加新页面

### 4.1 创建页面包

1. 在 `packages/loop-pages/` 下创建新页面包，如 `my-page/`

2. 创建 `package.json`:

```json
{
  "name": "@cozeloop/my-pages",
  "version": "0.0.1",
  "description": "My page for cozeloop",
  "main": "./src/index.ts",
  "scripts": {
    "build": "exit 0",
    "lint": "eslint ./ --cache",
    "test": "vitest --run --passWithNoTests"
  },
  "dependencies": {
    "@coze-arch/coze-design": "0.0.7-alpha.5f0418",
    "@cozeloop/components": "workspace:*",
    "@cozeloop/guard": "workspace:*",
    "@cozeloop/i18n-adapter": "workspace:*"
  },
  "devDependencies": {
    "@coze-arch/eslint-config": "workspace:*",
    "@coze-arch/ts-config": "workspace:*",
    "react": "~18.2.0",
    "react-dom": "~18.2.0",
    "react-router-dom": "^6.22.0"
  }
}
```

3. 创建 `src/index.ts`:

```typescript
// Copyright (c) 2025 coze-dev Authors
// SPDX-License-Identifier: Apache-2.0
export { MyPage } from './app';
export type { MyPageProps } from './types';
```

4. 创建页面组件

### 4.2 注册路由

在 `apps/cozeloop/src/routes/index.tsx` 中添加路由：

```typescript
import { lazy } from 'react';

// 懒加载新页面
const MyPage = lazy(() => import('@cozeloop/my-pages'));

// 在 routeConfig 中添加
{
  path: 'my-page/*',
  element: <MyPage />,
}
```

### 4.3 添加导航（如需要）

在 `apps/cozeloop/src/components/navbar/` 下添加导航项。

---

## 5. 添加新组件

### 5.1 组件层级

组件按层级位于不同包：

| 层级 | 位置 | 说明 |
|------|------|------|
| 通用组件 | `packages/loop-base/components` | 跨业务通用 |
| 业务组件 | `packages/loop-components/*` | 特定业务域 |
| 页面组件 | 各 `loop-pages/*/src/components` | 单个页面使用 |

### 5.2 创建通用组件

1. 在 `packages/loop-base/components/src/` 下创建组件文件

2. 导出到 `index.ts`:

```typescript
export { MyComponent } from './my-component';
export type { MyComponentProps } from './my-component';
```

### 5.3 创建业务组件

1. 在 `packages/loop-components/biz-components/src/` 创建

2. 或创建独立的业务组件包：

```bash
packages/loop-components/
└── my-business-components/
    ├── package.json
    └── src/
        ├── index.ts
        └── my-biz-component.tsx
```

### 5.4 组件规范

- 使用 TypeScript，定义完整的 Props 类型
- 使用 Tailwind CSS 样式
- 导出 `FC<Props>` 类型
- 提供默认导出

```typescript
// packages/loop-base/components/src/my-component.tsx
import React from 'react';
import classnames from 'classnames';

interface MyComponentProps {
  className?: string;
  title: string;
  onClick?: () => void;
}

export const MyComponent: React.FC<MyComponentProps> = ({
  className,
  title,
  onClick,
}) => {
  return (
    <div
      className={classnames('flex items-center p-4 bg-white rounded-lg', className)}
      onClick={onClick}
    >
      <span>{title}</span>
    </div>
  );
};
```

---

## 6. 添加 API 客户端

### 6.1 API Schema 管理

API 类型定义通过 IDL 自动生成，位于 `@cozeloop/api-schema` 包。

**现有 API 模块:**

| 模块 | 路径 |
|------|------|
| observation | `@cozeloop/api-schema/observation` |
| evaluation | `@cozeloop/api-schema/evaluation` |
| data | `@cozeloop/api-schema/data` |
| llm-manage | `@cozeloop/api-schema/llm-manage` |
| foundation | `@cozeloop/api-schema/foundation` |
| prompt | `@cozeloop/api-schema/prompt` |

### 6.2 使用 API Schema

```typescript
import type { ObservationAPI } from '@cozeloop/api-schema/observation';

// 使用生成的类型
const api: ObservationAPI = {
  // ...
};
```

### 6.3 更新 API

当后端 API 变更时：

```bash
cd packages/loop-base/api-schema

# 从 main 分支拉取 IDL
rushx update

# 或指定分支
# 修改 package.json 中 prethrift 脚本的 --branch 参数
```

---

## 7. 状态管理

### 7.1 使用 Zustand Store

#### 7.1.1 使用现有 Store

```typescript
import { useUIStore } from '@cozeloop/stores';

// 读取状态
const breadcrumbs = useUIStore(state => state.breadcrumbs);

// 更新状态
const setBreadcrumbs = useUIStore(state => state.setBreadcrumbs);
```

#### 7.1.2 创建新 Store

在 `packages/loop-base/stores/src/stores/` 下创建新 store 文件：

```typescript
// packages/loop-base/stores/src/stores/my-store.ts
import { create } from 'zustand';
import { immer } from 'zustand/middleware/immer';

interface MyState {
  count: number;
  increment: () => void;
}

export const useMyStore = create<MyState>()(
  immer((set) => ({
    count: 0,
    increment: () => set((state) => { state.count += 1; }),
  }))
);
```

### 7.2 导出 Store

在 `packages/loop-base/stores/src/index.ts` 中导出：

```typescript
export { useMyStore } from './stores/my-store';
```

---

## 8. 国际化

### 8.1 组件中使用 i18n

```typescript
import { I18n } from '@cozeloop/i18n-adapter';

// JSX 中使用
<I18n>{(t) => t('my_key')}</I18n>

// 或使用 hook
import { useTranslation } from '@cozeloop/i18n-adapter';

function MyComponent() {
  const { t } = useTranslation();
  return <span>{t('my_key')}</span>;
}
```

### 8.2 翻译 Key 管理

翻译资源文件位于各包的 `src/i18n/resource/` 目录：

- `zh-CN.json`: 中文
- `en-US.json`: 英文

### 8.3 添加新翻译

1. 在对应包的 i18n 资源文件中添加：

```json
// zh-CN.json
{
  "my_key": "我的文本"
}

// en-US.json
{
  "my_key": "My Text"
}
```

---

## 9. 权限控制

### 9.1 使用 Guard 组件

```typescript
import { Guard, GuardPoint } from '@cozeloop/guard';

// 组件级控制 - 隐藏无权限内容
<Guard point={GuardPoint['pe.prompts.create']}>
  <button>创建</button>
</Guard>

// 组件级控制 - 只读模式
<Guard point={GuardPoint['pe.prompts.edit']} mode="readonly">
  <Input />
</Guard>
```

### 9.2 权限点定义

权限点通过 `GuardPoint` 枚举定义，按业务域划分：

```typescript
enum GuardPoint {
  'pe.prompts.create' = 'pe.prompts.create',
  'pe.prompts.edit' = 'pe.prompts.edit',
  'pe.prompts.delete' = 'pe.prompts.delete',
  'evaluation.create' = 'evaluation.create',
  // ...
}
```

---

## 10. 样式规范

### 10.1 Tailwind CSS

所有样式优先使用 Tailwind CSS 类：

```tsx
// 推荐
<div className="flex items-center justify-between p-4 bg-white rounded-lg">

// 避免
<div style={{ display: 'flex', padding: '16px' }}>
```

### 10.2 Design Token

使用预定义的设计 Token：

```tsx
// 颜色 - 使用语义化 token
<div className="text-foreground-5 bg-background-2">

// 间距 - 使用 spacing token
<div className="p-4 m-2">

// 圆角
<div className="rounded-lg">
```

### 10.3 响应式设计

```tsx
// 移动优先
<div className="w-full md:w-1/2 lg:w-1/3">

// 隐藏/显示
<div className="hidden md:block">
```

### 10.4 Dark Mode

```tsx
// 自动适配 dark mode
<div className="bg-white dark:bg-gray-900">

// 或使用 CSS 变量
<div style={{ background: 'rgba(var(--background), 1)' }}>
```

---

## 11. 测试

### 11.1 编写测试

在 `src/__tests__/` 或 `*.test.tsx` 文件中编写：

```typescript
// packages/loop-base/components/src/__tests__/my-component.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MyComponent } from '../my-component';

describe('MyComponent', () => {
  it('renders title', () => {
    render(<MyComponent title="Test Title" />);
    expect(screen.getByText('Test Title')).toBeInTheDocument();
  });
});
```

### 11.2 运行测试

```bash
# 运行单个包测试
cd packages/loop-base/components
rushx test

# 运行所有测试
cd apps/cozeloop
rushx test

# 带覆盖率
rushx test:cov
```

### 11.3 Testing Library

使用 `@testing-library/react` 进行组件测试：

```typescript
import { render, screen, fireEvent } from '@testing-library/react';

fireEvent.click(screen.getByRole('button'));
expect(screen.getByText('Clicked')).toBeInTheDocument();
```

---

## 12. 代码规范

### 12.1 ESLint

```bash
# 检查代码
rushx lint

# 自动修复
rushx lint --fix
```

### 12.2 Prettier

代码格式化：

```bash
# 格式化所有文件
pnpm prettier --write .

# 检查格式
pnpm prettier --check .
```

### 12.3 Stylelint

样式检查：

```bash
pnpm stylelint "**/*.{css,less,scss}"
```

### 12.4 Git Hooks

项目配置了 pre-commit hooks，会自动运行 ESLint 和 Prettier。

---

## 13. 构建和发布

### 13.1 本地构建

```bash
# 开发构建
cd apps/cozeloop
rushx build:prod

# 分析 bundle
rushx analyze
```

### 13.2 环境变量

| 变量 | 说明 |
|------|------|
| `REGION` | 区域：`cn` |
| `CUSTOM_VERSION` | 版本：`inhouse` / `release` |
| `BUILD_TYPE` | 构建类型：`offline` / `online` |

### 13.3 构建输出

构建产物输出到 `apps/cozeloop/dist/` 目录。

---

## 14. 常见工作流

### 14.1 开发新功能

1. **创建分支**
   ```bash
   git checkout -b feat/my-feature
   ```

2. **安装依赖**
   ```bash
   rush update
   ```

3. **开发调试**
   ```bash
   cd apps/cozeloop
   rushx dev
   ```

4. **编写测试和代码**

5. **运行检查**
   ```bash
   rushx lint
   rushx test
   ```

6. **提交代码**
   ```bash
   git add .
   git commit -m "feat: add my feature"
   ```

### 14.2 更新 API Schema

1. 确保后端 IDL 已更新
2. 运行更新命令
   ```bash
   cd packages/loop-base/api-schema
   rushx update
   ```
3. 检查生成的类型是否正确
4. 提交变更

### 14.3 添加新依赖

1. 在对应包的 `package.json` 中添加
2. 运行 `rush update`
3. 如果是新 workspace 包，更新其他包的依赖

---

## 15. 目录约定

### 15.1 组件目录结构

```
my-package/
├── src/
│   ├── index.ts              # 包导出
│   ├── component-a.tsx       # 组件
│   ├── component-b.tsx       # 组件
│   ├── utils.ts              # 工具函数
│   ├── types.ts              # 类型定义
│   ├── hooks/                # 自定义 hooks
│   │   └── use-my-hook.ts
│   └── __tests__/            # 测试文件
│       └── component-a.test.tsx
├── package.json
└── tsconfig.json
```

### 15.2 命名约定

| 类型 | 约定 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `MyComponent.tsx` |
| 工具文件 | camelCase | `formatDate.ts` |
| 测试文件 | 同名 + .test | `MyComponent.test.tsx` |
| 类型文件 | types.ts 或 PascalCase | `MyTypes.ts` |

---

## 16. 故障排除

### 16.1 依赖问题

```bash
# 清理并重新安装
rm -rf node_modules
rush update
```

### 16.2 构建失败

```bash
# 清理构建缓存
cd apps/cozeloop
rm -rf dist node_modules/.cache
rushx build
```

### 16.3 类型错误

```bash
# 重新生成类型
cd apps/cozeloop
rm -rf node_modules/.cache
rushx build:ts
```

---

## 17. 相关资源

| 资源 | 链接 |
|------|------|
| Rush.js 文档 | https://rushjs.io/ |
| Rsbuild 文档 | https://rsbuild.dev/ |
| React Router v6 | https://reactrouter.com/ |
| Tailwind CSS | https://tailwindcss.com/ |
| Zustand | https://zustand.docs.pmnd.rs/ |
| Vitest | https://vitest.dev/ |
