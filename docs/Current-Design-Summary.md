# MrChat 当前设计摘要（基于现有两份设计文档）

范围：把仓库根目录的两份设计文档快速“对齐认知”，并标出需要用 ADR 收敛的关键分歧。重点只覆盖与“前端/性能/对话体验/网关”强相关的部分。

## 1. 两份文档的关系

- `Architecture-Design-OUI-Integration.md`（v2.0，2026-01-27）
  - 更像“目标架构蓝图”：融合 Open WebUI 的数据模型/权限/审计思路 + new-api 的网关/缓存/中间件经验
- `Plan.md`（v1.0，2026-01-21）
  - 更像“研发计划 + 初版表结构 + 计费流程草案”

建议：后续以 v2.0 作为主线，v1.0 保留做对照与补充。

## 2. 目标系统（从现有设计抽象）

MrChat 目标是一个“聚合镜像的 LLM 对话站”，核心能力分四块：

1. 对话体验：会话/消息、流式、模型切换、分享/导出、（未来）更强的结构化呈现
2. 网关能力：OpenAI 兼容接口、上游适配、多模型统一、可靠的流式转发
3. 多租户与运营：用户/认证、分组与 RBAC、配额/计费、支付（可选）、报表
4. 性能与稳定性：缓存/索引、限流、审计、可观测

## 3. 现有设计里已经比较明确的“好决策”

- 用户与认证分离（User / Auth / ApiKey 分表思想）
- 组与权限（RBAC + Groups），适配团队场景
- 模型配置表 + 适配器模式（provider adapter），向“多上游、多协议”扩展
- 计费流程采用“预扣 + 结算”
- 缓存走多级（内存 L1 + Redis L2 + DB），并配合索引优化
- 流式输出与实时能力：
  - SSE 流式响应（对话 token 流）
  - WebSocket Hub（在线状态/推送/协作类）
- 审计与安全：审计日志、JWT/API Key 双认证、速率限制

## 4. 当前最需要收敛的分歧（建议尽快 ADR 化）

- 前端框架：
  - 已决策：Vue 3 + Vite（见 `docs/adr/ADR-0001-frontend-framework-vue.md`）
- 主键/ID 方案：
  - v1 多处以 int 自增为主
  - v2 以 UUID 为主（更适配多租户与迁移/合并）
- “对话记录沉淀”的产品形态：
  - 仅 DB 存储 + 导出 Markdown
  - 还是把 Markdown 作为一等产物（可版本化、可检索、可复盘）
- 与 OpenWebUI 的关系：
  - 已决策：MVP 不引入外部对话 UI（不做 iframe/Portal），直接自研 simple Chat UI（见 `docs/adr/ADR-0002-chat-ui-strategy-self-built.md`）

## 5. 下一步建议（文档层面）

- 先把调研沉淀持续写进 `docs/research/`，再用 `docs/adr/` 把关键选型敲定
- 目前已补充的调研入口：
  - `docs/research/projects/new-api.md`（基于本仓库源码，偏前端/SSE/Markdown）
  - `docs/research/projects/open-webui.md`（基于 README/官方文档，含许可证风险点）
  - `docs/research/topics/frontend-performance.md` / `docs/research/topics/streaming.md` / `docs/research/topics/licensing.md`
