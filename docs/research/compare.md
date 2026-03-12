# 参考项目对比速览（面向 MrChat 取舍）

> 这份对比只保留与“前端/性能/网关/可运营”强相关的维度，细节见各项目调研文档。

| 项目 | 角色 | 技术栈（已知） | 借鉴重点（对 MrChat） | 主要风险/成本 | 许可证要点 |
|---|---|---|---|---|---|
| new-api | 网关 + 管理后台 | Go + Gin + GORM + React + Semi UI | 中继路由/middleware 组织、SSE 处理、前端渲染优化、Portal iframe 思路 | 领域重心偏网关，聊天产品需另起；二开成本与范围控制 | AGPLv3（`new-api/LICENSE`） |
| open-webui | 对话 UI 平台 | Svelte + Python（文档/README 描述） | 完整对话产品形态、RAG/插件生态、企业级能力 | 功能面大导致系统重；深度二开可能受限 | v0.6.6+ 有 Branding 限制（见 `docs/research/projects/open-webui.md`） |
| OVINC OpenWebUI fork | 对话 UI + 商业化 | 基于 open-webui（fork） | 计费/支付/报表的产品化落地 | 与上游不兼容风险；升级成本 | 待核验 |
| HiveChat | 团队对话 UI | Next.js + Tailwind + Auth.js + Postgres + Drizzle + AntD | 团队/权限/配额的产品设计、MCP/Agent 扩展思路 | 需要读源码确认前端性能与数据模型 | 待核验 |
| litellm | SDK + 代理网关 | Python + Proxy（文档/README 描述） | 多上游统一、虚拟 key/成本追踪/路由回退/可观测思路 | 与 Go 技术栈不同；集成方式需要权衡（独立服务 vs 借鉴设计） | 待核验 |

## 推进路线（已选择自研 Chat UI）

已决策：MVP 走 **自研对话内核路线**（见 `docs/adr/ADR-0002-chat-ui-strategy-self-built.md`），先上线“用户中心 + Chat”。

Portal/外部 UI 嵌入方案作为备选方案保留，用于：

- 临时对比竞品体验
- 需要快速验证某些重功能（RAG/插件生态）但不想立即自研时
