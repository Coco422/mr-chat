# GitHub 项目调查报告（来自 Grok）

来源：用户提供的 Grok 报告原文（未做二次核验）。

- 调查日期：2026-02-28
- 调查对象：HiveNexus/HiveChat、OVINC-CN/OpenWebUI、QuantumNous/new-api、open-webui/open-webui、BerriAI/litellm

---

**GitHub 项目调查报告**  
**调查日期**：2026 年 2 月 28 日  
**调查对象**：HiveNexus/HiveChat、OVINC-CN/OpenWebUI、QuantumNous/new-api、open-webui/open-webui、BerriAI/litellm  

这些项目全部属于 **LLM（大语言模型）生态工具**，主要解决“如何方便、安全、统一地使用多个 AI 模型”的问题。它们形成了一个完整的生态链：

- **UI 层**（用户聊天界面）：HiveChat、OpenWebUI（官方 + 增强 fork）
- **网关/代理层**（统一 API、路由、计费）：new-api、litellm

下面按项目逐一分析（数据来自 GitHub 主页 + README 实时抓取），最后给出对比与推荐。

## 1. HiveNexus/HiveChat（1.1k stars）

- **定位**：专为**中小团队/企业/学校**设计的集中式 AI 聊天平台。
- **核心亮点**：
  - 管理员一次配置，全团队直接使用（无需每个人填 API Key）。
  - 按**分组**管理用户 + 模型权限 + 月度 token 限额。
  - 支持 DeepSeek、OpenAI、Claude、Gemini、Moonshot、阿里、百度、腾讯、Ollama 等 10+ 家 + 任意 OpenAI 兼容接口。
  - 企业级登录：邮箱 + 企业微信/钉钉/飞书 SSO。
  - 管理员后台 + 用户聊天界面双端。
- **技术栈**：Next.js + Tailwind + Ant Design + PostgreSQL + Drizzle。
- **部署**：Docker Compose / Vercel 一键部署。
- **活跃度**：最后一次大更新在 2025 年 9 月（5 个月前），目前处于维护状态，非高频迭代。
- **适用场景**：公司内部 AI 助手，需要权限管控和配额的团队。

## 2. open-webui/open-webui（官方，125k stars） + OVINC-CN/OpenWebUI（增强 fork，309 stars）

**官方版（125k stars）** 是目前最受欢迎的自托管 LLM Web UI：

- **核心功能**：支持 Ollama、本地模型 + 任意 OpenAI 兼容后端；内置 RAG（9 种向量数据库）、网页搜索、语音对话、图像生成/编辑、模型构建器、Pipelines 插件系统、RBAC 权限、企业 SSO（LDAP/SCIM/Okta 等）。
- **技术栈**：Svelte 前端 + Python 后端，Docker 一键部署。
- **活跃度**：极高，上周还有合并 PR（2026-02-23）。
- **适用人群**：个人、本地玩家、追求极致功能的自托管用户。

**OVINC-CN/OpenWebUI（fork）** 是官方版的**商业化增强版**（中文社区维护）：

- 新增**计费、支付、充值、用户消费报表**（按 token 或请求次数）。
- 支持支付宝 + 易支付，自助充值。
- 适合想把 OpenWebUI 变成“可运营产品”的个人开发者或小团队。
- 最新 release：2026-02-24，dev 分支活跃。
- **注意**：与官方不兼容，需使用他们自己的 Docker 镜像。

## 3. QuantumNous/new-api（18.6k stars）

- **定位**：下一代 **LLM 统一网关 & 资产管理平台**（基于原 One-API 重写增强）。
- **核心亮点**：
  - 支持 OpenAI / Claude / Gemini 格式相互转换（跨协议调用）。
  - 智能路由（权重随机 + 自动重试）、用户级限流、缓存计费。
  - 内置充值（EPay/Stripe）、在线支付、多语言 UI（中/英/法/日）。
  - 数据仪表盘、权限分组、Discord/Telegram 登录。
  - 与原 One-API 数据库完全兼容，可无缝迁移。
- **技术栈**：Go（后端）+ Docker，SQLite/MySQL/PostgreSQL。
- **活跃度**：极高，**18 小时前**还有新 commit（Gemini Veo 更新），每天都在迭代。
- **适用场景**：需要把多家 API 聚合成一个入口的企业/开发者，常与 OpenWebUI 搭配使用。

## 4. BerriAI/litellm（37.3k stars）

- **定位**：**Python SDK + AI Gateway（代理服务器）**，统一调用 100+ 家 LLM。
- **核心亮点**：
  - 一行代码 `litellm.completion()` 就能调用 OpenAI/Anthropic/Gemini/Bedrock/Ollama 等全部模型。
  - 代理模式提供虚拟 Key、预算控制、守卫（guardrails）、负载均衡、详细成本追踪。
  - 支持 MCP 工具调用、A2A 代理协议、企业多租户仪表盘。
  - 性能极强（1k RPS 下 P95 延迟仅 150ms）。
- **技术栈**：Python（82%）+ FastAPI，Docker/Helm 部署。
- **活跃度**：**最高**，当天（2026-02-28）多个 commit + 安全修复。
- **适用场景**：开发者集成（SDK）、生产级代理（Proxy）、任何需要“统一 + 可观测 + 计费”的项目。

## 项目对比总结（2026 年 2 月最新）

| 项目 | Stars | 主要角色 | 计费/支付 | 多模型转换 | 活跃度 | 最佳搭配 | 推荐用户群 |
|---|---:|---|---|---|---|---|---|
| HiveChat | 1.1k | 团队聊天 UI | 原生 | 支持 | 中等（5 月前） | + new-api/litellm 后端 | 中小团队、企业内部 |
| OpenWebUI 官方 | 125k | 万能聊天 UI | 无 | 依赖后端 | 极高 | + litellm/new-api | 个人、本地玩家、公司 |
| OVINC fork | 309 | 商业化 UI | 原生 | 依赖后端 | 高 | 独立使用或 + 后端 | 想赚钱的个人/小团队 |
| new-api | 18.6k | 中文网关 | 原生 | 最强 | 极高 | + OpenWebUI/HiveChat 前端 | 中文社区、企业 |
| litellm | 37.3k | 国际网关/SDK | 原生 | 强 | 最高 | + 任意 UI | 开发者、生产环境 |

## 生态关系与推荐使用组合（2026 年主流方案）

1. **最强自托管方案**：OpenWebUI（官方） + litellm（代理） → 功能最全、性能最好。
2. **中文团队最优**：HiveChat（UI） + new-api（后端） → 权限 + 计费 + 企业登录全都有。
3. **想变现**：OVINC-CN/OpenWebUI 直接部署 + 支付宝充值 → 个人开发者最快上线。
4. **纯开发者集成**：直接用 litellm SDK，几乎零学习成本。
5. **企业生产级**：litellm Proxy（多租户） + OpenWebUI（前端） + Redis/PostgreSQL 集群。

**结论**：

这五个项目构成了 2026 年最成熟的开源 LLM 基础设施闭环。  

- 想**简单好看** → OpenWebUI 官方  
- 想**团队管控** → HiveChat  
- 想**统一后端** → new-api（中文）或 litellm（国际）  
- 想**带支付变现** → OVINC-CN fork  

