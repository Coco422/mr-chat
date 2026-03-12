# 专题调研：许可证与复用边界（务必尽早确定）

目标：避免在路线确定后才发现“不能二开/不能换皮/必须开源”，导致返工。

## 1. new-api（本仓库内 `new-api/`）

- `new-api/LICENSE` 为 GNU AGPLv3。
- 含义（对我们最关键的点）：如果 MrChat 直接复用/修改并作为网络服务对外提供，通常需要向网络用户提供对应的源代码（AGPL 的网络传播条款）。

建议：把 new-api 当作“参考实现”学习设计与工程手段；如要复用代码，需要明确开源策略或商业授权策略。

## 2. open-webui（官方文档说明）

官方文档说明（需以最终 license 文本为准）：

- `v0.6.5` 及以前：BSD-3-Clause
- 从 `v0.6.6`（2025-04-19）开始：采用 “Open WebUI License”，在 BSD-3-Clause 之上增加 Branding 限制（例如禁止随意移除/修改品牌标识等）

建议：如果 MrChat 目标是“可运营的聚合镜像站”，优先走“集成”而不是“深度 fork 换皮”：

- 通过 OpenAI 兼容接口接入
- 或以外部应用形式嵌入（iframe/反向代理），避免触碰二开边界

## 3. 其它项目（待补全）

- HiveChat：待确认 LICENSE 与商业使用条款
- OVINC-CN/OpenWebUI：待确认 LICENSE 与与上游差异
- litellm：待确认 LICENSE、商用限制、与其 “Enterprise” 功能边界

