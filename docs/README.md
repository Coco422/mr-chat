# MrChat 文档索引

本仓库当前处于“设计与调研”阶段，核心目标是做一个聚合镜像的 LLM 对话站，重点关注：

- 前端体验与性能（加载速度、流式渲染、长对话可用性）
- 对话体验（更好的呈现与编辑能力、可复盘、可导出）
- 网关能力（统一多模型、多租户、配额与计费、审计）

## 现有设计文档（仓库根目录）

- `Architecture-Design-OUI-Integration.md`
  - 更新时间：2026-01-27（v2.0）
  - 更偏“融合 Open WebUI + new-api 的架构设计”，覆盖多租户/权限/模型适配/实时/缓存/安全等。
- `Plan.md`
  - 创建时间：2026-01-21（v1.0）
  - 更偏“基于 new-api 的 MrChat 研发计划与初版表结构/流程”。

两份旧文档存在少量不一致（例如 ID 方案、MVP 边界、实时能力范围等）。当前建议以统一基线文档为主，再回看旧文档补背景。

如果只想先看当前统一口径，建议先看：`docs/Requirements-Baseline-v0.1.md`。
如果只想快速对齐旧文档关系，可看：`docs/Current-Design-Summary.md`。
MVP 上线范围草案见：`docs/MVP-v0.1-Scope.md`。

已收敛的选型：

- 前端框架：Vue 3 + Vite（见 `docs/adr/ADR-0001-frontend-framework-vue.md`）
- Chat UI 策略：核心 Chat 自研，外部子服务入口作为补充能力（见 `docs/adr/ADR-0002-chat-ui-strategy-self-built.md`）
- 后端框架：Gin（见 `docs/adr/ADR-0003-backend-framework-gin.md`）
- 默认 ORM：GORM（见 `docs/adr/ADR-0006-orm-gorm.md`）
- 主数据库：PostgreSQL（见 `docs/adr/ADR-0004-database-postgresql.md`）
- 数据库迁移工具：goose（见 `docs/adr/ADR-0005-migration-tool-goose.md`）
- Redis 使用策略：仅作缓存/限流/共享短状态，并要求故障可降级（见 `docs/adr/ADR-0007-redis-runtime-strategy.md`）

## 当前推荐阅读顺序

1. `docs/Requirements-Baseline-v0.1.md`
2. `docs/Page-and-Route-Spec-v0.1.md`
3. `docs/API-Contract-v0.1.md`
4. `docs/Data-Model-and-State-v0.1.md`
5. `docs/Development-Task-Breakdown-v0.1.md`
6. `docs/Current-Design-Summary.md`

## 调研产出（本目录）

调研文档统一放在：

- `docs/research/`：项目调研与专题分析
- `docs/adr/`：关键决策的记录（选型、协议、存储模型等）

建议写法：每个结论都尽量落到“我们要借鉴什么”“我们不做什么”“为什么”“下一步验证方式”。
