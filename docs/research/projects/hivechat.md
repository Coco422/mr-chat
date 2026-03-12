# 项目调研：HiveChat（基于 GitHub README，待进一步补全）

信息来源：GitHub README（读取时间：2026-03-01）。当前仓库内没有 HiveChat 源码，后续如需深入前端性能与数据模型，仍需拉取源码分析。

项目地址：`https://github.com/HiveNexus/HiveChat`

## 0. 已知信息（先沉淀）

### 0.1 产品定位与亮点（从 README 摘要）

- 团队/企业/校园场景的统一聊天平台：管理员配置一次，团队成员直接用
- 支持按“分组”做模型与额度管理（README 描述包含月度 token 限额）
- 支持多家模型与 OpenAI 兼容后端（README 列表包含：OpenAI/Claude/Gemini/DeepSeek/火山/阿里/Qwen/智谱/百度/腾讯/硅基流动/Ollama/OpenRouter）
- 对话体验：支持 DeepSeek “思考链”展示、LaTeX 与 Markdown 渲染、图像理解
- 扩展：支持 MCP Server（SSE mode）、支持 Agents、支持云存储（S3/阿里云/腾讯云/七牛/Cloudflare R2）

### 0.2 技术栈（README 明示）

- Next.js
- Tailwindcss
- Auth.js
- PostgreSQL
- Drizzle ORM
- Ant Design

## 1. 为什么要看

从外部报告看，HiveChat 更偏“团队对话平台”，这对 MrChat 的多租户/分组/配额是强相关的：

- 管理员一次配置，全员使用（避免每人填 key）
- 分组、权限、配额、SSO 等产品形态更接近商业化落地

## 2. 我们关心的调研问题（按优先级）

- 对话产品形态：
  - 会话/文件夹/标签/分享的交互
  - 多模型切换与上下文管理怎么做得不打断用户
- 团队管理：
  - 分组模型、权限粒度、配额策略（月度/每日/模型维度）
  - 邀请、审计、管理员可观测（谁在用、花了多少）
- 前端工程：
  - Next.js 的渲染策略（SSR/CSR）、首屏与路由拆包
  - 长对话性能、流式渲染策略
- 数据模型：
  - message 存储、搜索与导出（是否支持 Markdown）

## 3. 待补全的动作

- 拉取源码后输出：
  - “团队场景对话产品”的关键交互清单
  - 可直接借鉴的 UI/信息架构（IA）
  - 对多租户与配额模型的落地建议（含表结构建议）

## 4. 对 MrChat 的直接启发（第一批）

- “管理员一次配置，全员使用”可以作为 MrChat 的主路径之一：把 key 配置与计费/审计收敛到平台侧
- MCP Server（SSE mode）值得专项调研：它对“工具调用/Agent 扩展/工作流”会产生架构影响
- DeepSeek 思考链展示这类 UX 细节非常贴合“更好的对话”目标（与 new-api 的 reasoning/think 处理可以对照）
