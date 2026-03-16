# MrChat v0.1 开发任务拆解清单

- 状态：执行清单草案
- 日期：2026-03-17
- 依赖文档：
  - `docs/Requirements-Baseline-v0.1.md`
  - `docs/Page-and-Route-Spec-v0.1.md`
  - `docs/API-Contract-v0.1.md`
  - `docs/Data-Model-and-State-v0.1.md`

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

### 3.3 最值得先锁定的阻塞项

- 数据表迁移方案
- goose 迁移目录、命名规范与执行方式
- GORM 模型与仓储分层约定
- 鉴权方案与 token 刷新方式
- SSE 事件格式
- `quota_logs` 的账本语义
- 模型到上游的路由绑定结构

## 4. M0：工程基础与开发环境

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `INF-01` | 初始化后端工程骨架（Gin、GORM、配置、路由、控制器、服务、模型目录） | Backend | P0 | M | 无 | 能启动空服务、初始化 GORM 数据访问层并返回健康检查 |
| `INF-02` | 初始化前端工程骨架（Vue 3 + Vite + Router + 状态管理） | Frontend | P0 | M | 无 | 能启动前端并完成基础路由跳转 |
| `INF-03` | 建立 `.env` 与配置加载规范 | Fullstack | P0 | S | 无 | 本地、测试环境都能通过配置启动 |
| `INF-04` | 准备本地依赖启动方式（PostgreSQL、Redis、可选 Mail mock） | Ops | P0 | S | 无 | 一条命令能起本地依赖 |
| `INF-07` | 建立 goose 迁移目录、命名规范与本地/CI 执行命令 | Backend | P0 | S | `INF-01`、`INF-04` | 本地与 CI 都能用 goose 执行 `status` / `validate` / `up` |
| `INF-05` | 建立基础 CI（lint/test/build 占位） | Fullstack | P0 | M | `INF-01`、`INF-02` | PR 至少能跑通基础校验 |
| `INF-06` | 统一日志、`request_id`、错误处理中间件 | Backend | P0 | M | `INF-01` | 所有请求都有 `request_id`，错误结构统一 |

## 5. M1：数据层与认证基础

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `DB-01` | 建立 `users`、`auths`、`groups`、`group_members` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 可在 PostgreSQL 上通过 goose 成功迁移并回滚 |
| `DB-02` | 建立 `upstreams`、`models`、`model_route_bindings` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 路由配置相关表可在 PostgreSQL 上通过 goose 迁移 |
| `DB-03` | 建立 `conversations`、`messages`、`quota_logs` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 聊天与账本表可在 PostgreSQL 上通过 goose 迁移 |
| `DB-04` | 建立 `redeem_codes`、`redeem_redemptions`、`audit_logs` 迁移 | Backend | P0 | M | `INF-01`、`INF-07` | 兑换与审计表可在 PostgreSQL 上通过 goose 迁移 |
| `AUTH-BE-01` | 实现注册、登录、退出、刷新 token | Backend | P0 | M | `DB-01` | 四个接口按契约工作 |
| `AUTH-BE-02` | 实现 JWT/角色中间件与受保护路由守卫 | Backend | P0 | M | `AUTH-BE-01` | `User/Admin/Root` 权限可控 |
| `AUTH-BE-03` | 实现登录失败次数限制、基础风控日志与限流 | Backend | P0 | M | `AUTH-BE-01`、`DB-04`、`INF-06` | 登录失败与限流可控，并有安全相关日志可查 |
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
| `ADMIN-BE-02` | 实现模型 CRUD API、可见组配置与路由绑定保存 | Backend | P0 | L | `DB-02`、`ADMIN-BE-01` | 模型、可见组和优先级绑定可维护 |
| `ADMIN-BE-03` | 实现用户查询与人工调额 API | Backend | P0 | M | `DB-03`、`DB-04` | 可按用户调额并写账本/审计 |
| `ADMIN-BE-04` | 实现审计日志查询 API | Backend | P0 | S | `DB-04` | 后台可查关键操作日志 |
| `GROUP-BE-01` | 实现用户组 CRUD 与成员维护 API | Backend | P0 | M | `DB-01`、`AUTH-BE-02` | 管理员可维护组、成员归属，并供模型可见性与路由分组使用 |
| `MODEL-BE-01` | 实现 `GET /models` 用户侧模型列表 API | Backend | P0 | S | `ADMIN-BE-02`、`GROUP-BE-01`、`AUTH-BE-02` | `/chat` 首屏可返回当前用户有权限看到的模型列表 |
| `ADMIN-FE-01` | 管理后台壳子与导航 | Frontend | P0 | M | `APP-FE-01` | `/admin/*` 页面框架可用 |
| `ADMIN-FE-02` | 上游管理页 | Frontend | P0 | M | `ADMIN-BE-01` | 能配置和修改上游 |
| `ADMIN-FE-03` | 模型管理页 | Frontend | P0 | M | `ADMIN-BE-02`、`GROUP-BE-01` | 能维护模型、可见组与优先级 |
| `ADMIN-FE-04` | 用户管理页 | Frontend | P0 | M | `ADMIN-BE-03`、`GROUP-BE-01` | 能查用户、调额并维护用户组归属 |
| `ADMIN-FE-05` | 审计日志页 | Frontend | P0 | S | `ADMIN-BE-04` | 能筛查关键审计记录 |

## 7. M3：Chat 后端闭环

| 任务 ID | 任务 | 端别 | 优先级 | 规模 | 依赖 | 完成标准 |
|---|---|---|---|---|---|---|
| `CHAT-BE-01` | 实现会话 CRUD API | Backend | P0 | M | `DB-03`、`AUTH-BE-02` | 会话列表可按最近时间分页返回，并支持创建、重命名、软删 |
| `CHAT-BE-02` | 实现消息列表 API | Backend | P0 | S | `DB-03`、`AUTH-BE-02` | 能分页返回消息 |
| `CHAT-BE-03` | 实现 OpenAI 兼容上游客户端 | Backend | P0 | M | `ADMIN-BE-01` | 能调用一个上游完成非流式请求 |
| `CHAT-BE-04` | 实现模型路由器、优先级 Failover 与上游冷却 | Backend | P0 | L | `CHAT-BE-03`、`ADMIN-BE-02` | 故障时自动切换下一个上游，并按服务商记录失败次数与冷却状态 |
| `CHAT-BE-05` | 实现 Chat Completions 接口（SSE/非流式）与断连取消 | Backend | P0 | L | `CHAT-BE-03` | 同一入口支持流式/非流式，SSE 能稳定输出并响应 stop |
| `CHAT-BE-06` | 实现消息持久化与状态流转 | Backend | P0 | M | `CHAT-BE-05`、`DB-03` | 消息状态可经历 `pending/streaming/completed/...` |
| `CHAT-BE-07` | 实现 usage 采集与本地估算回退 | Backend | P0 | M | `CHAT-BE-05` | 上游无 usage 时仍可结算 |
| `CHAT-BE-08` | 实现请求日志、路由日志、错误码收敛 | Backend | P0 | M | `INF-06`、`CHAT-BE-04` | 一次聊天可完整追踪请求链路 |

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
| `OBS-01` | 补基础观测：请求日志、路由日志、账本日志、审计日志 | Backend | P0 | S | `INF-06`、`CHAT-BE-08` | 线上排障所需日志齐备 |
| `OPS-01` | 补部署与环境文档 | Ops | P0 | S | `INF-04` | 新环境能按文档部署 |
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

## 12. 建议最先创建的 issue 列表

如果要先起一批 issue，建议优先创建这 12 个：

1. `INF-01` 后端工程骨架（Gin + GORM）
2. `INF-02` 前端工程骨架
3. `INF-04` 本地 PostgreSQL / Redis 启动方式
4. `INF-07` goose 迁移规范与执行命令
5. `DB-01` 用户、认证与分组迁移
6. `DB-02` 模型与上游迁移
7. `AUTH-BE-01` 登录注册刷新
8. `AUTH-BE-02` JWT 与角色守卫
9. `AUTH-BE-03` 登录安全与风控日志
10. `AUTH-FE-01` 登录注册页面
11. `ADMIN-BE-01` 上游 CRUD
12. `ADMIN-BE-02` 模型、可见组与路由绑定

## 13. 推荐第一轮 Sprint 切法

### Sprint 1

- `INF-01` ~ `INF-06`
- `INF-07`
- `DB-01` ~ `DB-04`
- `AUTH-BE-01` ~ `AUTH-BE-03`
- `USER-BE-01`
- `AUTH-FE-01`
- `APP-FE-01`

目标：

- 能登录
- 能拿到当前用户
- 登录安全基线到位
- goose 迁移链路可用
- 工程能跑、表能迁移

### Sprint 2

- `USER-BE-02`
- `USER-BE-03`
- `USER-FE-01`
- `USER-FE-02`
- `ADMIN-BE-01` ~ `ADMIN-BE-04`
- `GROUP-BE-01`
- `MODEL-BE-01`
- `ADMIN-FE-01` ~ `ADMIN-FE-05`
- `CHAT-BE-01`
- `CHAT-BE-02`
- `CHAT-FE-01`
- `CHAT-FE-02`

目标：

- 用户设置页可用
- 管理员可配置基础资源
- 管理员可维护用户组与模型可见性
- 用户能看到会话壳子与会话列表

### Sprint 3

- `CHAT-BE-03` ~ `CHAT-BE-08`
- `CHAT-FE-03` ~ `CHAT-FE-09`
- `BILL-BE-01`

目标：

- 一次完整的流式聊天跑通
- 路由、日志、结算初步闭环

### Sprint 4

- `BILL-BE-02` ~ `BILL-BE-05`
- `USAGE-FE-01`
- `USAGE-FE-02`
- `ADMIN-FE-06`
- `QA-01` ~ `QA-04`

目标：

- 额度、兑换码、账单全部闭环
- 达到首发验收标准
