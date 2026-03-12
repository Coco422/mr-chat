# 项目调研：new-api（基于本仓库源码）

> 本调研基于本仓库内的 `new-api/` 代码快照（不是在线 README 摘要），重点聚焦：前端与流式体验、网关路径的性能手段。

## 1. 定位与我们关心的价值

new-api 更像“统一 LLM 网关 + 资产管理后台”，但它在两个方向对 MrChat 很有参考价值：

- 网关侧：OpenAI 兼容入口、多格式中继、限流/计费/统计、对多上游的适配层组织方式
- 前端侧：React + Semi UI 的可维护性、SSE 流式渲染与调试面板、Markdown 渲染能力与一些 UX 细节（reasoning/think 标签）

## 2. 后端要点（性能相关）

### 2.1 路由与中继结构

- 核心中继路由集中在 `new-api/router/relay-router.go`
  - `/v1/*`：OpenAI 兼容中继（含 `/v1/chat/completions`、`/v1/responses` 等）
  - `/pg/chat/completions`：Playground 专用入口（UI 用来测试/调试 SSE）
  - `/v1/realtime`：WebSocket realtime（统一走 relay）

### 2.2 中间件链路里值得借鉴的点

`new-api/router/relay-router.go` 的全局 middleware（按顺序）：

- `middleware.CORS()`：跨域
- `middleware.DecompressRequestMiddleware()`：支持压缩请求体（带宽与上行性能）
- `middleware.BodyStorageCleanup()`：清理请求体存储（避免长时间占用）
- `middleware.StatsMiddleware()`：统计/指标采集

在 `/v1` Relay 组上额外加了：

- `middleware.SystemPerformanceCheck()`：系统性能保护（高负载时降级/拒绝）
- `middleware.TokenAuth()`：token 鉴权（网关核心）
- `middleware.ModelRequestRateLimit()`：模型请求速率限制（按模型做保护）
- `middleware.Distribute()`：分发到具体渠道/上游

这些点基本都是“把稳定性与性能保护前置”，对聚合镜像站很关键。

### 2.3 压缩与剖析（可观测）

- 响应压缩：`gin-contrib/gzip` 在多个路由组启用（例如 `new-api/router/api-router.go`、`new-api/router/web-router.go`）
- 请求解压：`middleware.DecompressRequestMiddleware()`（对应实现可从 `new-api/middleware/gzip.go` 继续深挖）
- Profiling：内置 Pyroscope 启动逻辑（`new-api/common/pyro.go`，入口 `new-api/main.go` 调用）

### 2.4 适配与流式扫描

代码里有针对 `/v1/chat/completions` 的流式扫描与测试用例：

- `new-api/relay/helper/stream_scanner_test.go`

对我们来说，价值在于：

- 流式解析的边界条件（断行、DONE、异常 chunk）要在测试里跑过
- 不同上游/不同格式的“usage 统计”与流式结算，通常会卡在这里

## 3. 前端要点（性能与更好的对话）

### 3.1 技术栈

`new-api/web/package.json`：

- React 18 + Vite
- Semi UI（`@douyinfe/semi-ui`）
- `react-markdown` + `rehype-highlight` + `rehype-katex` + `mermaid`
- `sse.js` 用于以 POST 方式建立 SSE 流（比原生 EventSource 更灵活）

### 3.2 SSE 流式请求与健壮性处理

核心在 `new-api/web/src/hooks/playground/useApiRequest.jsx`：

- 同时支持非流式与流式（SSE）请求
- SSE 事件处理包含：
  - 收到 `[DONE]` 正常收尾
  - JSON parse 失败兜底与 UI 错误态收敛
  - readyState/status 异常检测，避免“正常关闭也当成 error”重复上屏
  - `onStopGenerator` 主动停止：关闭 SSE 并把最后一条“未闭合 think 标签”做清理后落库（UI 侧保存）

这套模式建议直接复用到 MrChat 的“聊天流式”链路里（无论后端是 Go 还是 proxy）。

### 3.3 “更好的对话”：reasoning/think 的呈现与自动折叠

Playground 的消息渲染做了两层兼容：

- 如果上游走 OpenAI 类增量字段：支持 `delta.reasoning_content` / `delta.reasoning`（在 SSE 里分离更新）
- 如果模型把思考包在 `<think>...</think>`：会在渲染时提取并放到“思考过程”区域

相关代码：

- 消息解析与抽取：`new-api/web/src/helpers/utils.jsx`（`processThinkTags` / `processIncompleteThinkTags`）
- 消息展示：`new-api/web/src/components/playground/MessageContent.jsx` + `ThinkingContent.jsx`
- 自动折叠逻辑：`new-api/web/src/hooks/playground/useApiRequest.jsx`（`applyAutoCollapseLogic`）

这块是 MrChat 可以做差异化的核心：把“思考/正文/引用/工具调用”等拆成可折叠的结构化块。

### 3.4 Markdown 渲染与流式动画（体验点）

`new-api/web/src/components/common/markdown/MarkdownRenderer.jsx`：

- Markdown 能力很全：GFM、数学公式、代码高亮、Mermaid、代码复制、HTML 预览（iframe sandbox）
- 支持“流式新文字淡入”：
  - rehype 插件：`new-api/web/src/helpers/render.jsx`（`rehypeSplitWordsIntoSpans`）
  - 用 `previousContentLength` 判断哪一段是新增内容，只对新增词加动画 class

这对“流式时看起来更顺滑”很有帮助，但也带来性能成本（rehype + 分词 + span 拆分）。MrChat 需要明确性能预算，并考虑：

- 只对最后一条 assistant 消息启用动画
- 大段内容达到阈值后关闭动画
- Mermaid/高亮等重型渲染按需启用或延迟渲染

### 3.5 渲染性能：React.memo + 精准对比

`new-api/web/src/components/playground/OptimizedComponents.js`：

- 对 MessageContent/Actions/Settings/Debug 面板做 `React.memo` 包装
- 对 MessageContent 的 props 采用“字段级对比”，减少无意义 rerender

注意：对某些大对象直接 `JSON.stringify` 对比会引入 CPU 开销，MrChat 如果走长对话/高频流式，建议：

- 让上层 state 拆分更细（messages / settings / debug 分 store）
- 通过 stable references + selector 来避免 stringify

### 3.6 “聚合镜像”的 UI 形态：iframe 嵌入外部 Chat

`new-api/web/src/pages/Chat/index.jsx`：

- 会把某个聊天应用 URL 模板（包含 `{address}` / `{key}` 占位）与用户 key 拼装后，用 `iframe` 打开

`new-api/web/src/pages/Setting/Chat/SettingsChats.jsx`：

- 提供“聊天应用配置”的可视化编辑，最终存到 `Chats` 这个 option（JSON）

这对我们很直接：

- MrChat 可以先做一个“Portal”，把 OpenWebUI/HiveChat 等作为外部应用嵌入
- 同时逐步替换成我们自己的高性能对话 UI（从 0 开发成本更低）

## 4. 对 MrChat 的可落地借鉴清单（第一批）

- SSE hook 的整体结构（启动/停止/异常收敛/完成时落库）
- reasoning_content + <think> 双兼容的渲染策略
- MarkdownRenderer 的能力拆分（先保留基础：GFM + code + math，再按需启用 mermaid/html preview）
- “聚合镜像 Portal” 的 iframe 拼装模型（地址 + key 注入 + 多应用配置）
- 网关路由的 middleware 顺序：解压/统计/性能保护/鉴权/限流/分发

## 5. 待补充（需要继续读的后端点）

- 计费与结算：预扣/缓存计费/usage 的来源与一致性策略
- 分发与重试：`middleware.Distribute()` + relay 内部的路由策略（权重/失败剔除）
- 指标/Profiling：StatsMiddleware 与 pyroscope 的接入点（落到我们的可观测方案）
