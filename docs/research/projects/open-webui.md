# 项目调研：open-webui（基于 README/官方文档，待进一步补全）

信息来源：open-webui GitHub README + 官方文档（读取时间：2026-03-01）。当前仓库内没有 open-webui 源码，后续如需确认具体实现（前端渲染/流式策略/数据模型），仍需拉取源码分析。

项目地址：`https://github.com/open-webui/open-webui`

## 0. 已知信息（先沉淀）

### 0.1 定位与核心能力（README/文档摘要）

- 自托管 AI 平台，支持离线运行
- 支持 Ollama 与 OpenAI 兼容 API
- 对话体验能力：
  - 多模型并行对话（multiple models concurrently）
  - Markdown 与 LaTeX 渲染
  - RBAC（角色权限控制）
- 生态与扩展：
  - Pipelines 插件系统
  - “Native Python Function Calling Tool”
- RAG：
  - 9 种向量数据库（Chroma、Milvus、Qdrant、Pinecone、Weaviate、Redis、Neo4j 等）
  - 多种文档解析引擎
  - 15+ Web Search Providers
- 可观测与扩展性（文档描述）：
  - OpenTelemetry
  - 多副本水平扩展 + Redis session management + WebSocket communication

### 0.2 许可与品牌限制（高优先级风险点）

官方文档明确：

- `v0.6.5` 及以前：BSD-3-Clause
- 从 `v0.6.6`（2025-04-19）开始：采用 “Open WebUI License”，在 BSD-3-Clause 基础上额外增加了关于 Branding 的限制

这会直接影响 MrChat 的策略选择：

- 如果我们想“二开 open-webui 并更换品牌”，需要先确认许可证是否允许，或者考虑 Enterprise 许可
- 更现实的路线是：把 open-webui 当作“外部应用”接入（iframe/反向代理），MrChat 自己做网关与计费等

## 1. 为什么要看

open-webui 在“对话 UI 产品化”这条线上几乎是事实标准：

- 功能面非常全（RAG、搜索、语音、图片、工具/插件、权限/SSO）
- UI/交互迭代快，社区活跃
- 前后端分离，且能对接 OpenAI 兼容网关（正好对上 MrChat/new-api/litellm）

## 2. 我们关心的调研问题（按优先级）

### 2.1 前端性能与体验

- 长对话列表是否做了虚拟滚动？怎么做的？有哪些边界 bug？
- 流式渲染：token 级更新的节流策略是什么？如何避免“每个 chunk 都触发全树渲染”？
- 消息结构：是否区分正文/推理/工具调用/引用？如何折叠与定位？
- Markdown 渲染：用什么库？如何处理代码高亮、数学、图表、附件？
- 资源加载：首屏包体、按路由拆包、缓存策略、字体/图片策略

### 2.2 “更好的对话”数据模型

- conversation/message 的存储结构（DB schema / JSONB / blob）
- 搜索：全文索引怎么做？是否支持跨对话检索？
- 分享与权限：share link 的安全边界、token 化策略
- 导出：是否支持 Markdown/JSON/可复盘格式？导出包含哪些 metadata？

### 2.3 插件与可扩展

- Pipelines 插件系统的边界：UI 插件？后端插件？数据流怎么穿透？
- 工具调用/函数调用：如何对不同 provider 做兼容

## 3. 我们可能要借鉴的点（先列候选）

- Chat UX：消息分块、引用与溯源、快捷操作、上下文管理
- 插件体系：可控、安全、可运营的扩展机制
- RAG 设计：把 RAG 做成“可插拔能力”而不是耦合在聊天里

## 4. 待补全的动作

- 拉取源码后：
  - 固定入口点（前端 store、消息渲染、stream handler、markdown renderer）
  - 跑一轮基本性能 profiling（长对话、流式、搜索）
  - 输出“可直接抄作业”的实现清单（组件/协议/状态管理）

## 5. 对 MrChat 的直接启发（第一批）

- RAG 与 Pipelines 的产品形态值得借鉴，但最好做成“可插拔能力”，避免把对话内核做重
- open-webui 的许可证与 Branding 条款需要尽早纳入技术路线讨论（直接 fork 可能不适合）
