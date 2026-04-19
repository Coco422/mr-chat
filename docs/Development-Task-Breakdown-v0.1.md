# MrChat v0.1 开发任务拆解清单

- 状态：执行中
- 日期：2026-04-19
- 更新摘要：已补 `user_groups` / `channels` / limit policies / user adjustments / `llm_request_logs` 设计与首版实现，并已接通首个基于数据库配置的 OpenAI 兼容上游、Swagger UI 与 `stream=true` SSE 主链路；已落地管理控制台重构第一阶段（上游模型发现、模型导入、admin 详情接口与 human-readable 关联返回）以及新版 Chat 工作区 UI；当前剩余交付顺序见 `docs/Delivery-Plan-v0.1-Remaining.md`
- 依赖文档：
  - `docs/Requirements-Baseline-v0.1.md`
  - `docs/Page-and-Route-Spec-v0.1.md`
  - `docs/API-Contract-v0.1.md`
  - `docs/Data-Model-and-State-v0.1.md`
  - `docs/Admin-Console-Refactor-Plan-v0.1.md`
  - `docs/Delivery-Plan-v0.1-Remaining.md`
  - `docs/Implementation-Progress.md`

## 0. 当前实现快照

截至 `2026-04-19`，当前任务推进情况可简化理解为：

- 已落地：
  - `INF-01`、`INF-02`、`INF-03`、`INF-04`、`INF-06`、`INF-07`
  - `DB-01` ~ `DB-04`
  - `AUTH-BE-01`、`AUTH-BE-02`
  - `USER-BE-01`、`USER-BE-02`、`USER-BE-03`
  - `AUTH-FE-01`、`APP-FE-01`、`USER-FE-01`、`USER-FE-02`
  - `ADMIN-BE-03`、`ADMIN-BE-04`、`ADMIN-BE-05`、`ADMIN-BE-06`
  - `MODEL-BE-01`
  - `CHAT-BE-01`、`CHAT-BE-02`、`CHAT-BE-03`、`CHAT-BE-05`
  - `GROUP-BE-01`
  - `LIMIT-BE-01`
  - `LIMIT-BE-02`
- 已进入开发但未完全完成：
  - `INF-08`
  - `AUTH-BE-03`
  - `ADMIN-BE-01`
  - `ADMIN-BE-02`
  - `CHAT-BE-03` ~ `CHAT-BE-08`
  - `ADMIN-BE-07` 的“停用/删除前检查”后续仍待补

当前已经可以稳定支撑：

- 注册 / 登录 / 当前用户 / 个人设置 / 安全设置 / 用量页联调
- Admin 上游 / 模型 / 用户调额 / 审计日志联调
- Admin 渠道 / 用户组 / 分组限额 / 用户限额调整联调
- Admin 上游 / 渠道 / 模型 / 用户组详情联调
- Admin 上游模型发现与模型导入联调
- Admin 模型 human-readable 返回联调
- Admin `references` 轻量选项接口联调
- Chat 非流式主链路联调：`/api/v1/chat/completions -> upstream -> messages -> llm_request_logs`
- Chat 流式主链路联调：`/api/v1/chat/completions(stream=true) -> SSE -> messages -> llm_request_logs`
- Chat 侧模型列表、会话 CRUD、消息列表联调
- Chat 工作区 UI 联调：侧栏会话导航、模型选择、`reasoning_content` 分区、复制 / 重试 / 编辑、流式状态反馈
- Swagger UI 联调入口：`/swagger/index.html`

当前已明确需要继续收敛的一条支线：

- 管理控制台需要从“开发骨架”升级到“可运营维护”
- 重点问题包括：
  - 模型需改成“上游发现 + 导入”
  - 管理台默认不再直出 UUID
  - `channel` 需收敛为高级配置语义
  - 上游 / 渠道 / 模型 / 用户组需补详情、编辑与安全停用策略

## 1. 使用方式

这份清单的目标不是替代详细设计，而是把现有文档压成可直接建 issue / 排 Sprint 的开发任务。

字段含义：

- `任务 ID`：建议直接作为 issue 前缀
- `优先级`：`P0` 表示首发阻塞项，`P1` 表示可后置
- `规模`：`S/M/L`
- `依赖`：开始前应先完成的任务
- `完成标准`：可用于验收

## 2. 建议里程碑

| 里程碑 | 目标 | 主要产出 |
|---|---|---|
| `M0` | 工程基础与开发环境 | 项目骨架、环境配置、基础 CI、数据库/Redis 启动能力 |
| `M1` | 数据层与认证基础 | 核心表迁移、鉴权、当前用户与路由守卫 |
| `M2` | 管理配置骨架 | 上游/模型/用户/审计的后台 API 与管理页骨架 |
| `M3` | Chat 后端闭环 | 会话、消息、SSE、路由、结算、日志 |
| `M4` | Chat 前端闭环 | 登录后主 Chat、会话列表、模型选择、流式渲染 |
| `M5` | 额度与兑换码闭环 | 用户用量页、兑换码、管理员调额与充值日志 |
| `M6` | 联调、验收与上线准备 | 关键链路验证、性能回归、安全检查、部署文档 |
| `M7` | P1 能力 | 邀请码、导出、API Key、外部子服务入口 |

## 3. 推荐并行策略

### 3.1 后端主线

- `M0 -> M1 -> M2/M3 -> M5 -> M6`

### 3.2 前端主线

- 在 `M1` 的鉴权与基础契约稳定后，可并行做登录页、壳子和 Chat 主框架
- 在 `M2` 管理 API 稳定后，再推进后台页面
- 在 `M3` SSE 契约稳定后，再推进流式细节和错误收敛

### 3.3 当前最值得先锁定的阻塞项

- `CHAT-BE-04`：多上游 failover、冷却与回切尚未闭环
- `BILL-BE-01`：聊天返回中的 `billing` 仍为 `0`，`quota_logs` 账本尚未打通
- `INF-08` + `AUTH-BE-03`：Redis 降级与登录风控仍未达到“故障不阻断核心链路”的标准
- `BILL-BE-02` ~ `BILL-BE-05`：兑换码链路仍停留在 schema 已有、产品链路未闭环的状态
- `INF-05` + `QA-*`：自动化测试和 CI 基线仍明显偏弱，发布可信度不足

补充说明：

- 当前剩余工作的推荐顺序、Sprint 切法与 issue 队列，不再从 `M0/M1` 初始化任务开始排，而是以 `docs/Delivery-Plan-v0.1-Remaining.md` 为准。

## 4. M0：工程基础与开发环境

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `INF-01` | 初始化后端工程骨架（Gin、GORM、配置、路由、控制器、服务、模型目录） | Backend | P0 | M | 无 | 能启动空服务、初始化 GORM 数据访问层并返回健康检查 |
| `INF-02` | 初始化前端工程骨架（Vue 3 + Vite + Router + 状态管理） | Frontend | P0 | M | 无 | 能启动前端并完成基础路由跳转 |
| `INF-03` | 建立 `.env` 与配置加载规范 | Fullstack | P0 | S | 无 | 本地、测试环境都能通过配置启动 |
| `INF-04` | 准备本地依赖启动方式（PostgreSQL、Redis、可选 Mail mock） | Ops | P0 | S | 无 | 一条命令能起本地依赖 |
| `INF-07` | 建立 goose 迁移目录、命名规范与本地/CI 执行命令 | Backend | P0 | S | `INF-01`、`INF-04` | 本地与 CI 都能用 goose 执行 `status` / `validate` / `up` |
| `INF-08` | 约束 Redis 抽象层、key 命名空间、TTL 与降级策略 | Backend | P0 | M | `INF-01`、`INF-04` | Redis 用途边界明确，缓存/限流/冷却在 Redis 故障时都能自动降级而不阻塞核心链路 |
| `INF-05` | 建立基础 CI（lint/test/build 占位） | Fullstack | P0 | M | `INF-01`、`INF-02` | PR 至少能跑通基础校验 |
| `INF-06` | 统一日志、`request_id`、错误处理中间件 | Backend | P0 | M | `INF-01` | 所有请求都有 `request_id`，错误结构统一 |

## 5. M1：数据层与认证基础

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `DB-01` | 建立 `users`、`auths`、用户分组基础迁移，并明确 `users.user_group_id` 单归属口径 | Backend | P0 | M | `INF-01`、`INF-07` | 可在 PostgreSQL 上通过 goose 成功迁移并回滚 |
| `DB-02` | 建立 `upstreams`、`channels`、`models`、`model_route_bindings` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 路由配置相关表可在 PostgreSQL 上通过 goose 迁移 |
| `DB-03` | 建立 `conversations`、`messages`、`quota_logs`、`llm_request_logs` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 聊天、账本和请求日志表可在 PostgreSQL 上通过 goose 迁移 |
| `DB-04` | 建立 `redeem_codes`、`redeem_redemptions`、`audit_logs` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 兑换与审计表可在 PostgreSQL 上通过 goose 迁移 |
| `AUTH-BE-01` | 实现注册、登录、退出、刷新 token | Backend | P0 | M | `DB-01` | 四个接口按契约工作 |
| `AUTH-BE-02` | 实现 JWT/角色中间件与受保护路由守卫 | Backend | P0 | M | `AUTH-BE-01` | `User/Admin/Root` 权限可控 |
| `AUTH-BE-03` | 实现登录失败次数限制、基础风控日志与限流 | Backend | P0 | M | `AUTH-BE-01`、`DB-04`、`INF-06`、`INF-08` | 登录失败与限流可控，并在 Redis 不可用时降级为单实例内存策略与安全日志 |
| `USER-BE-01` | 实现 `GET /users/me` 与 `PUT /users/me` | Backend | P0 | S | `AUTH-BE-01` | 用户资料可查看与更新 |
| `USER-BE-02` | 实现 `GET /users/me/quota` 与 `GET /users/me/usage` 骨架 | Backend | P0 | M | `DB-03` | 返回真实或占位统计结构 |
| `USER-BE-03` | 实现密码修改与安全信息接口 | Backend | P0 | S | `AUTH-BE-01`、`AUTH-BE-03` | 用户可修改密码，并可查看最近登录等安全信息 |
| `AUTH-FE-01` | 登录/注册页与基础表单校验 | Frontend | P0 | M | `INF-02`、`AUTH-BE-01` | 用户能完成登录注册 |
| `APP-FE-01` | 登录态管理、路由守卫、全局 AppShell | Frontend | P0 | M | `AUTH-FE-01`、`AUTH-BE-02` | 未登录跳 `/login`，已登录进 `/chat` |
| `USER-FE-01` | 实现 `/settings/profile` 页面 | Frontend | P0 | S | `APP-FE-01`、`USER-BE-01` | 用户可维护资料与偏好设置 |
| `USER-FE-02` | 实现 `/settings/security` 页面 | Frontend | P0 | M | `APP-FE-01`、`AUTH-BE-03`、`USER-BE-03` | 用户可修改密码并查看基础安全信息 |

## 6. M2：管理配置骨架

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `ADMIN-BE-01` | 实现上游 CRUD API | Backend | P0 | M | `DB-02`、`AUTH-BE-02` | 可增删改查上游 |
| `ADMIN-BE-02` | 实现模型 CRUD API、可见用户组配置与渠道路由绑定保存 | Backend | P0 | L | `DB-02`、`ADMIN-BE-01` | 模型、可见用户组和优先级绑定可维护 |
| `ADMIN-BE-03` | 实现用户查询与人工调额 API | Backend | P0 | M | `DB-03`、`DB-04` | 可按用户调额并写账本/审计 |
| `ADMIN-BE-04` | 实现审计日志查询 API | Backend | P0 | S | `DB-04` | 后台可查关键操作日志 |
| `ADMIN-BE-05` | 实现上游模型发现接口与标准化候选模型返回 | Backend | P0 | M | `ADMIN-BE-01`、`CHAT-BE-03` | 可从指定 upstream 拉取候选模型，并统一成管理台可导入格式 |
| `ADMIN-BE-06` | 实现模型导入接口与 admin human-readable 关联返回 | Backend | P0 | L | `ADMIN-BE-02`、`ADMIN-BE-05`、`GROUP-BE-01` | 已完成第一版：模型导入接口可用，模型返回已补用户组名、渠道名、上游名与 summary，不再只暴露 UUID |
| `ADMIN-BE-07` | 补上游/渠道/模型/用户组单资源详情接口 | Backend | P0 | M | `ADMIN-BE-01`、`ADMIN-BE-02`、`GROUP-BE-01` | 已完成第一版：详情接口可用；后续继续补停用/删除前检查与编辑体验 |
| `ADMIN-BE-08` | 建立管理资源安全停用/删除策略 | Backend | P0 | M | `ADMIN-BE-07`、`DB-02` | 优先支持停用；删除需经过引用检查并明确语义 |
| `GROUP-BE-01` | 实现 `user_groups` CRUD 与单用户归属维护 API | Backend | P0 | M | `DB-01`、`AUTH-BE-02` | 管理员可维护用户组与用户归属，并供模型可见性与限额策略使用 |
| `LIMIT-BE-01` | 实现用户组模型限额模板 API | Backend | P0 | M | `DB-01`、`DB-02`、`AUTH-BE-02` | 可批量维护默认模板与模型覆盖规则 |
| `LIMIT-BE-02` | 实现用户限额使用统计与 direct adjustment API | Backend | P0 | M | `LIMIT-BE-01`、`DB-03`、`AUTH-BE-02` | 可查询 hour/week/lifetime 使用与剩余额度，并记录单用户调整 |
| `MODEL-BE-01` | 实现 `GET /models` 用户侧模型列表 API | Backend | P0 | S | `ADMIN-BE-02`、`GROUP-BE-01`、`AUTH-BE-02` | `/chat` 首屏可返回当前用户有权限看到的模型列表 |
| `ADMIN-FE-01` | 管理后台壳子与导航 | Frontend | P0 | M | `APP-FE-01` | `/admin/*` 页面框架可用 |
| `ADMIN-FE-02` | 上游管理页 | Frontend | P0 | M | `ADMIN-BE-01` | 能配置和修改上游 |
| `ADMIN-FE-03` | 模型管理页 | Frontend | P0 | M | `ADMIN-BE-02`、`GROUP-BE-01` | 能维护模型、可见组与优先级 |
| `ADMIN-FE-04` | 用户管理页 | Frontend | P0 | M | `ADMIN-BE-03`、`GROUP-BE-01` | 能查用户、调额并维护用户组归属 |
| `ADMIN-FE-05` | 审计日志页 | Frontend | P0 | S | `ADMIN-BE-04` | 能筛查关键审计记录 |
| `ADMIN-FE-07` | 模型发现、勾选导入与路由规则编辑页 | Frontend | P0 | L | `ADMIN-BE-05`、`ADMIN-BE-06` | 管理员可从 upstream 拉模型并导入，路由规则可视化编辑 |
| `ADMIN-FE-08` | 上游/渠道/模型/用户组详情与编辑页 | Frontend | P0 | M | `ADMIN-BE-07` | 管理台从“新增 + 列表”升级到可维护资源 |
| `ADMIN-FE-09` | 管理台 human-readable 展示清理 | Frontend | P0 | M | `ADMIN-BE-06`、`ADMIN-FE-08` | 除审计日志外，主界面不再直接展示 UUID |

## 7. M3：Chat 后端闭环

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `CHAT-BE-01` | 实现会话 CRUD API | Backend | P0 | M | `DB-03`、`AUTH-BE-02` | 会话列表可按最近时间分页返回，并支持创建、重命名、软删 |
| `CHAT-BE-02` | 实现消息列表 API | Backend | P0 | S | `DB-03`、`AUTH-BE-02` | 能分页返回消息 |
| `CHAT-BE-03` | 实现 OpenAI 兼容上游客户端 | Backend | P0 | M | `ADMIN-BE-01` | 已完成：可按数据库中的 upstream 配置调用一个 OpenAI 兼容上游完成非流式请求 |
| `CHAT-BE-04` | 实现模型路由器、优先级 Failover 与上游冷却 | Backend | P0 | L | `CHAT-BE-03`、`ADMIN-BE-02`、`INF-08` | 故障时自动切换下一个上游，并按服务商记录失败次数与冷却状态；Redis 故障时退化为单实例内存冷却 |
| `CHAT-BE-05` | 实现 Chat Completions 接口（SSE/非流式）与断连取消 | Backend | P0 | L | `CHAT-BE-03` | 已完成第一版：非流式与 SSE `POST /api/v1/chat/completions` 已可用，客户端断连可取消上游请求；独立 stop API 仍待后续演进 |
| `CHAT-BE-06` | 实现消息持久化与状态流转 | Backend | P0 | M | `CHAT-BE-05`、`DB-03` | 部分完成：user/assistant 消息已落库，完整状态机仍待扩展 |
| `CHAT-BE-07` | 实现 usage 采集与本地估算回退 | Backend | P0 | M | `CHAT-BE-05` | 部分完成：优先采用上游 usage，回退估算已接入但尚未进入结算闭环 |
| `CHAT-BE-08` | 实现请求日志、路由日志、错误码收敛 | Backend | P0 | M | `INF-06`、`CHAT-BE-04` | 部分完成：`llm_request_logs` 已记录真实请求链路，路由日志与聚合报表待补齐 |

## 8. M4：Chat 前端闭环

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `CHAT-FE-01` | 实现 `/chat` 页面基础布局 | Frontend | P0 | M | `APP-FE-01`、`CHAT-BE-01` | 左栏/主区/输入区完整可用 |
| `CHAT-FE-02` | 实现会话列表、创建、重命名、删除 | Frontend | P0 | M | `CHAT-FE-01`、`CHAT-BE-01` | 基本会话管理完成 |
| `CHAT-FE-03` | 实现消息列表与分页加载 | Frontend | P0 | M | `CHAT-BE-02` | 会话消息可加载与展示 |
| `CHAT-FE-04` | 实现模型选择器与权限过滤展示 | Frontend | P0 | S | `CHAT-BE-04`、`MODEL-BE-01` | 用户只看到可用模型 |
| `CHAT-FE-05` | 实现 SSE 客户端、流式缓冲与节流 flush | Frontend | P0 | L | `CHAT-BE-05` | 流式体验稳定，不卡输入 |
| `CHAT-FE-06` | 实现停止生成、失败重试与错误提示 | Frontend | P0 | M | `CHAT-FE-05` | 用户能主动 stop 并重试 |
| `CHAT-FE-07` | 实现基础 Markdown 渲染与代码块样式 | Frontend | P0 | M | `CHAT-FE-03` | GFM + code block 正常显示 |
| `CHAT-FE-08` | 实现 `reasoning_content` 折叠展示 | Frontend | P0 | M | `CHAT-FE-05` | 推理与正文分区显示 |
| `CHAT-FE-09` | 做首轮性能约束落地 | Frontend | P0 | M | `CHAT-FE-05`、`CHAT-FE-07` | 只更新最后一条 assistant，长列表可用 |

## 9. M5：额度与兑换码闭环

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `BILL-BE-01` | 实现预扣、最终结算、退款逻辑 | Backend | P0 | L | `CHAT-BE-05`、`CHAT-BE-07`、`DB-03` | 聊天请求的账本闭环完成 |
| `BILL-BE-02` | 实现兑换码批量生成 API | Backend | P0 | M | `DB-04`、`AUTH-BE-02` | 管理员可生成批次 |
| `BILL-BE-03` | 实现兑换码兑换 API 与事务幂等 | Backend | P0 | M | `BILL-BE-02`、`DB-04` | 一次性码不可重复兑换 |
| `BILL-BE-04` | 实现账单流水与摘要 API | Backend | P0 | M | `BILL-BE-01`、`DB-03` | 用户用量与账单可查询 |
| `BILL-BE-05` | 实现后台兑换码批次与兑换记录查询 API | Backend | P0 | S | `DB-04`、`AUTH-BE-02` | 后台可查询兑换码批次统计与兑换记录 |
| `USAGE-FE-01` | 实现 `/usage` 页面摘要与流水 | Frontend | P0 | M | `USER-BE-02`、`BILL-BE-04` | 用户可查看额度与用量 |
| `USAGE-FE-02` | 实现兑换码表单与反馈 | Frontend | P0 | S | `BILL-BE-03` | 用户可完成兑换并刷新余额 |
| `ADMIN-FE-06` | 实现兑换码批量生成与查询页 | Frontend | P0 | M | `BILL-BE-02`、`BILL-BE-03`、`BILL-BE-05` | 管理员可管理兑换码批次并查看兑换记录 |

## 10. M6：联调、验收与上线准备

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `QA-01` | 跑通关键链路 E2E：注册 -> 登录 -> Chat -> 兑换 -> 继续聊天 | Fullstack | P0 | M | `M1`~`M5` | 与基线验收路径一致 |
| `QA-02` | 验证上游故障 -> 自动切换 -> 恢复后回切 | Backend | P0 | M | `CHAT-BE-04` | 路由日志与行为一致 |
| `QA-03` | 验证 stop/cancel 的结算与退款 | Fullstack | P0 | M | `CHAT-BE-05`、`BILL-BE-01` | 停止生成后余额与流水正确 |
| `QA-04` | 权限与安全检查 | Fullstack | P0 | M | `AUTH-BE-02`、`ADMIN-*` | 普通用户无法访问后台与敏感操作 |
| `QA-05` | 验证 Redis 故障降级 | Fullstack | P0 | M | `INF-08`、`AUTH-BE-03`、`CHAT-BE-04` | Redis 不可用时缓存旁路、限流/冷却退化为单实例策略，登录/Chat/结算/兑换链路不崩溃 |
| `OBS-01` | 补基础观测：请求日志、路由日志、账本日志、审计日志 | Backend | P0 | S | `INF-06`、`CHAT-BE-08` | 线上排障所需日志齐备 |
| `OPS-01` | 补部署与环境文档 | Ops | P0 | S | `INF-04`、`INF-08` | 新环境能按文档部署，并包含 Redis 故障处理与降级说明 |
| `DOC-01` | 回填实现结果到设计文档 | Fullstack | P0 | S | `M6` 完成前 | 文档与实现口径一致 |

## 11. M7：P1 任务池

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `P1-INV-01` | 邀请码绑定 inviter 与邀请记录 | Fullstack | P1 | M | `DB-01`、`AUTH-BE-01` | 注册时可选绑定 inviter |
| `P1-EXP-01` | 会话导出 Markdown | Fullstack | P1 | M | `CHAT-BE-02` | 用户可导出 Markdown |
| `P1-API-01` | 面向外部的 API Key 管理 | Fullstack | P1 | M | `AUTH-BE-02`、`CHAT-BE-03` | 支持 API Key 创建与撤销 |
| `P1-SVC-01` | `service_entries` 表、管理 API 与用户侧可见列表 API | Backend | P1 | M | `DB-04` | 后台能管理外部子服务入口，用户侧可获取可见服务列表 |
| `P1-SVC-02` | `/services` 与 `/services/:id` 页面 | Frontend | P1 | M | `P1-SVC-01` | 用户能看到并进入服务入口 |
| `P1-SVC-03` | `/admin/service-entries` 管理页 | Frontend | P1 | M | `P1-SVC-01` | 管理员能配置 iframe/跳转服务 |

## 12. 当前建议最先创建的 issue 列表

如果现在继续拆 issue，建议优先创建这 12 个：

1. `CHAT-BE-04` 模型路由器的多上游 failover、冷却与回切
2. `CHAT-BE-07` usage 回退口径接入正式结算
3. `BILL-BE-01` 预扣、最终结算、退款闭环
4. `INF-08` Redis key 命名空间、TTL 与降级策略落地
5. `AUTH-BE-03` 登录失败限制、风控日志与基础限流
6. `CHAT-BE-08` 路由日志、错误码和报表字段收敛
7. `ADMIN-BE-08` 管理资源停用 / 删除前检查策略
8. `BILL-BE-02` 兑换码批量生成 API
9. `BILL-BE-03` 兑换码兑换 API 与事务幂等
10. `BILL-BE-05` 后台兑换批次与兑换记录查询 API
11. `ADMIN-FE-06` 兑换码管理页
12. `INF-05` 最小 CI（Go + Web build/typecheck + smoke）

## 13. 当前推荐 Sprint 切法

### Sprint A：把聊天和账本补成“可信链路”

- `CHAT-BE-04`
- `CHAT-BE-07`
- `BILL-BE-01`
- `INF-08`
- `AUTH-BE-03`

目标：

- 上游故障能切换
- `billing` 字段不再固定为零
- Redis 故障不阻塞核心链路
- 登录安全基线到位

### Sprint B：把运营和用户余额路径补成闭环

- `CHAT-BE-08`
- `ADMIN-BE-08`
- `BILL-BE-02` ~ `BILL-BE-05`
- `ADMIN-FE-06`
- `USAGE-FE-01`
- `USAGE-FE-02`
- `CHAT-FE-09`

目标：

- 运营能发码、查码、查账
- 用户能看用量、兑换额度、继续聊天
- Chat 体验完成第一轮性能收口

### Sprint C：把仓库补成“可发布状态”

- `INF-05`
- `QA-01` ~ `QA-05`
- `OBS-01`
- `OPS-01`
- `DOC-01`

目标：

- 形成固定回归路径
- 基础日志、部署文档、设计文档回填完成
- 达到首发验收标准
