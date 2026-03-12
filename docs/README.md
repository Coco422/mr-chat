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

两份文档存在少量不一致（例如前端栈 Vue vs React、ID 方案等），后续用 ADR（架构决策记录）收敛。

如果只想先对齐认知，建议先看：`docs/Current-Design-Summary.md`。
MVP 上线范围见：`docs/MVP-v0.1-Scope.md`。

已收敛的选型：

- 前端框架：Vue 3 + Vite（见 `docs/adr/ADR-0001-frontend-framework-vue.md`）
- Chat UI 策略：MVP 自研 simple Chat UI，不嵌入外部对话 UI（见 `docs/adr/ADR-0002-chat-ui-strategy-self-built.md`）

## 调研产出（本目录）

调研文档统一放在：

- `docs/research/`：项目调研与专题分析
- `docs/adr/`：关键决策的记录（选型、协议、存储模型等）

建议写法：每个结论都尽量落到“我们要借鉴什么”“我们不做什么”“为什么”“下一步验证方式”。
