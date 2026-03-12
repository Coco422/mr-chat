# ADR-0001: 前端框架选型为 Vue 3 + Vite（SPA）

- 状态：通过
- 日期：2026-03-01
- 相关文档：
  - `docs/Current-Design-Summary.md`
  - `Architecture-Design-OUI-Integration.md`
  - `Plan.md`
  - `docs/research/topics/frontend-performance.md`
  - `docs/research/topics/streaming.md`

## 背景与问题

MrChat 的核心形态是“登录后的对话应用 + 流式输出 + 长对话渲染”。当前文档对前端框架存在分歧（React vs Vue），需要尽早收敛，避免后续实现与组件生态分裂。

我们的核心约束：

- 团队熟悉度优先（开发效率与长期维护）
- 性能优先（流式渲染、长对话、Markdown 渲染控制）
- 后端已有 Go 网关与业务服务，前端不需要承担 BFF 代理流式

## 选项

### 选项 A：Vue 3 + Vite（SPA）

- 优点：
  - 团队熟悉，交付效率高
  - SPA 形态适合“登录后应用”，部署简单（静态资源 + CDN/Go/Nginx）
  - Vue 3 的响应式与组件更新粒度适合做高频局部更新（只更新最后一条 assistant）
- 缺点：
  - 需要重新选定 Vue 生态中的状态管理、路由、Markdown 渲染与虚拟列表方案

### 选项 B：React + Vite（SPA）

- 优点：
  - 可直接借鉴 `new-api/web/` 的实现（SSE、MarkdownRenderer、组件拆分）
- 缺点：
  - 团队不熟悉会拖慢交付与维护

### 选项 C：Next.js（React，SSR/RSC）

- 优点：
  - 适合营销/SEO/落地页与统一的鉴权入口
- 缺点：
  - 聊天核心页面 SSR/RSC 收益有限
  - 增加 Node 运行时与部署复杂度
  - 若用 API Route 代理 SSE 容易引入额外开销与断流风险（需要刻意避免）

## 决策

选择 **Vue 3 + Vite（SPA）** 作为 MrChat 的前端框架。

理由：

- 团队熟悉度带来的交付效率与维护可控性，对早期产品迭代更关键
- MrChat 的核心是“登录后应用”，SSR/RSC 不是必需品
- 性能关键在“流式渲染与长对话”，与框架无强绑定，更多取决于组件/状态/渲染策略

## 影响

- 前端工程以 Vue 3 生态为基准（路由/状态/组件库/虚拟列表/Markdown 渲染方案后续再定 ADR 或在实现中收敛）
- Go 后端继续作为唯一业务入口与网关：SSE/WS 由前端直连 Go，避免前端层 BFF 代理流式
- `Plan.md` 中的 React 相关描述需要更新为 Vue 3

## 后续行动

1. 同步更新文档：`Plan.md`、`docs/Current-Design-Summary.md`，消除 React/Vue 分歧。
2. 追加一个子决策（可选）：UI 组件库与状态管理（例如 Element Plus/Arco/Naive、Pinia 等）。
3. 在前端实现里强制性能约束：
   - 流式阶段只更新最后一条 assistant
   - 节流 flush（按帧或固定间隔）
   - 长列表虚拟化
   - Markdown 分级渲染（流式轻渲染，完成后重渲染）

