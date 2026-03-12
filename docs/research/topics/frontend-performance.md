# 专题调研：前端性能（对话站）

## 1. 目标与性能预算（建议先定）

建议用“对话体验”视角定预算，而不是只盯首屏：

- 首屏：可交互 < 1.5s（4G/中端机），包体与关键 CSS 可控
- 首 token：点击发送后 < 800ms（取决于上游，但 UI 不该拖后腿）
- 流式渲染：滚动不掉帧，输入框不被卡顿影响（避免 setState 触发全树 rerender）
- 长对话：1000+ 条消息仍可用（滚动、搜索、定位、展开折叠）

## 2. 性能“重灾区”清单（对话站常见）

- Markdown 渲染过重（highlight/katex/mermaid/html preview）
- 流式 chunk 到来频率高，导致 React 频繁 commit
- debug/trace 面板与聊天渲染耦合在同一个 state 树里
- 长消息列表未虚拟化，滚动触发大量 DOM
- 图片/附件预览同步渲染，抢占主线程

## 3. 可复用的工程手段（可直接落到实现）

### 3.1 状态与渲染拆分

- 把聊天消息、设置、调试信息拆成独立 store（避免“任何一个变化都让 Chat rerender”）
- “流式只更新最后一条 assistant 消息”：其它历史消息尽量不可变
- 消息组件做 `React.memo`，对比尽量字段级，不要 `JSON.stringify` 大对象

注：本专题里引用了 new-api 的 React 实现作为样例（例如 `React.memo`），但结论同样适用于 Vue 3。Vue 侧的等价抓手通常是：

- 避免让高频流式更新污染全局响应式依赖（把“最后一条 assistant”做成独立的 `shallowRef`/store slice）
- 组件拆分 + 精准 props，必要时用 `v-memo`/`defineProps` + `markRaw` 控制不必要的深层追踪
- 长列表用虚拟滚动组件，避免一次性挂载大量 DOM

参考：new-api 对 Message/Panel 的 memo 化在 `new-api/web/src/components/playground/OptimizedComponents.js`。

### 3.2 流式更新节流与批处理

- 把上游 delta 先 append 到 buffer，再按帧（`requestAnimationFrame`）或固定间隔（例如 33ms）批量 flush 到 React state
- debug 面板更新用更慢的节流（例如 200-500ms），避免和正文抢资源

参考：new-api 的 SSE 处理在 `new-api/web/src/hooks/playground/useApiRequest.jsx`（可在 MrChat 中加入节流/批处理增强）。

### 3.3 长列表虚拟化

如果 Semi UI 的 `Chat` 组件无法满足 1000+ 消息场景，建议：

- 自建消息列表 + 虚拟滚动（例如 `react-virtuoso` / `react-window`）
- “定位到某条消息、展开思考过程、复制代码”等交互要和虚拟列表兼容

### 3.4 Markdown 渲染分级

建议分三档能力，避免一开始把渲染做成“全都开”：

- L0：GFM + code block（无高亮）+ inline code
- L1：高亮 + 数学（katex）
- L2：mermaid + HTML preview + 重型 artifact

流式阶段只做 L0/L1，消息完成后再升级到 L2（延迟渲染/懒加载）。

参考：new-api 的 MarkdownRenderer 在 `new-api/web/src/components/common/markdown/MarkdownRenderer.jsx`，以及流式淡入动画插件 `new-api/web/src/helpers/render.jsx`（`rehypeSplitWordsIntoSpans`）。

### 3.5 “更好的对话”不等于更重的 UI

- reasoning/think 内容默认折叠，减少首屏信息密度与渲染压力
- 图片/附件预览按需展开

参考：new-api 的 reasoning/think 处理在：

- `new-api/web/src/helpers/utils.jsx`（think 标签抽取）
- `new-api/web/src/components/playground/ThinkingContent.jsx`（折叠面板）

## 4. 验证方式（建议每个迭代都跑）

- React Profiler：看流式阶段 commit 次数、最慢组件
- Performance 面板：主线程任务分布（是否被 markdown/mermaid 卡住）
- 长对话压测脚本：模拟 1000 条消息 + 连续流式 2 分钟
