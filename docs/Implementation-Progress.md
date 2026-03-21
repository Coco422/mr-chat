# MrChat 实现进度快照

- 状态：首轮实现进行中
- 日期：2026-03-21
- 当前阶段：`M0/M1` 已完成，`M2` 已进入“管理台可维护化”收敛阶段，`M3` 已接通首个真实上游并进入聊天闭环细化

## 1. 当前结论

仓库已经不是“只有设计文档”的状态，当前已经进入“可持续开发 + 可联调验证”的实现阶段。

截至 2026-03-18，项目状态可以概括为：

- 后端工程骨架、配置加载、goose 迁移入口、基础鉴权和用户中心 API 已可用
- 核心表结构已经落到 PostgreSQL，并支持服务启动自动迁移
- 管理侧已具备上游、渠道、模型、用户组、用户调额、用户限额调整、审计日志的首版 API
- 管理侧已补齐上游/渠道/模型/用户组详情接口，并进入“模型发现 + 导入 + human-readable 返回”的过渡阶段
- Chat 侧已具备模型列表、会话 CRUD、消息分页查询、非流式与 SSE `/chat/completions`、用户限额前置校验、消息持久化和 `llm_request_logs` 落库
- Swagger UI 已接入，可直接通过 `/swagger/index.html` 查看当前接口
- 前端已提供无样式联调壳子，可直接对接登录、设置、用量、Chat 和主要 Admin 页面

## 2. 已落地能力

### 2.1 工程与数据层

- 后端入口：`cmd/api`
- 迁移入口：`cmd/migrate`
- goose SQL 已建立并应用到：
  - `users` / `auths`
  - `upstreams` / `channels` / `models` / `model_route_bindings`
  - `user_groups` / `user_group_model_limit_policies` / `user_limit_adjustments`
  - `conversations` / `messages`
  - `quota_logs` / `llm_request_logs`
  - `redeem_codes` / `redeem_redemptions` / `audit_logs`
- 服务启动默认执行 pending migrations

### 2.2 已实现 API

- Auth
  - `POST /api/v1/auth/signup`
  - `POST /api/v1/auth/signin`
  - `POST /api/v1/auth/signout`
  - `POST /api/v1/auth/refresh`
- User / Billing
  - `GET /api/v1/users/me`
  - `PUT /api/v1/users/me`
  - `GET /api/v1/users/me/security`
  - `PUT /api/v1/users/me/password`
  - `GET /api/v1/users/me/quota`
  - `GET /api/v1/users/me/usage`
  - `GET /api/v1/billing/summary`
  - `GET /api/v1/billing/logs`
- Chat
  - `GET /api/v1/models`
  - `GET /api/v1/conversations`
  - `POST /api/v1/conversations`
  - `PUT /api/v1/conversations/:id`
  - `DELETE /api/v1/conversations/:id`
  - `GET /api/v1/conversations/:id/messages`
  - `POST /api/v1/chat/completions`
- Swagger
  - `GET /swagger/index.html`
  - `GET /swagger/doc.json`
- Admin
  - `GET /api/v1/admin/upstreams`
  - `GET /api/v1/admin/upstreams/:id`
  - `GET /api/v1/admin/upstreams/:id/discovered-models`
  - `POST /api/v1/admin/upstreams`
  - `PUT /api/v1/admin/upstreams/:id`
  - `GET /api/v1/admin/channels`
  - `GET /api/v1/admin/channels/:id`
  - `POST /api/v1/admin/channels`
  - `PUT /api/v1/admin/channels/:id`
  - `GET /api/v1/admin/models`
  - `GET /api/v1/admin/models/:id`
  - `POST /api/v1/admin/models/import`
  - `POST /api/v1/admin/models`
  - `PUT /api/v1/admin/models/:id`
  - `GET /api/v1/admin/user-groups`
  - `GET /api/v1/admin/user-groups/:id`
  - `POST /api/v1/admin/user-groups`
  - `PUT /api/v1/admin/user-groups/:id`
  - `GET /api/v1/admin/user-groups/:id/limits`
  - `PUT /api/v1/admin/user-groups/:id/limits`
  - `GET /api/v1/admin/users`
  - `PUT /api/v1/admin/users/:id/group`
  - `PUT /api/v1/admin/users/:id/quota`
  - `GET /api/v1/admin/users/:id/limit-usage`
  - `GET /api/v1/admin/users/:id/limit-adjustments`
  - `POST /api/v1/admin/users/:id/limit-adjustments`
  - `GET /api/v1/admin/audit-logs`

### 2.3 已可联调页面

以下页面当前均为“无样式功能壳子”，只保留最小逻辑与接口对接：

- `/login`
- `/signup`
- `/settings/profile`
- `/settings/security`
- `/usage`
- `/chat`
- `/chat/:conversationId`
- `/admin/upstreams`
- `/admin/channels`
- `/admin/models`
- `/admin/user-groups`
- `/admin/users`
- `/admin/audit-logs`

## 3. 已验证结果

### 3.1 命令级验证

- `go test ./...` 通过
- `pnpm install --frozen-lockfile` 已补齐前端依赖
- `pnpm typecheck` 通过
- `pnpm build` 通过
- `make swagger` 能生成并更新 `internal/http/swagger/`
- `go run ./cmd/migrate status` 能识别局域网 PostgreSQL 上的 pending migration
- `go run ./cmd/migrate up` 已将局域网 PostgreSQL 升到 version `6`
- `go run ./cmd/api` 启动时日志显示 `goose: no migrations to run. current version: 6`
- 新增 admin detail / import / human-readable 接口编译与烟雾验证通过

### 3.2 局域网 PostgreSQL 与 `newapi` 烟雾验证

已在局域网 PostgreSQL + 局域网 `newapi` 环境验证通过以下链路：

- 新建临时 admin/user 测试账号
- 提升测试 admin 角色后重新登录
- Admin 创建 upstream
- Admin 创建 channel
- Admin 创建 user group
- Admin 创建 model，并保存 `visible_user_group_ids + channel_id + upstream_id` 路由绑定
- Admin 为 user group 批量写入默认模板和模型覆盖规则
- Admin 将普通用户分配到该 user group
- Admin 查询该用户在指定模型下的 limit usage，返回 `policy_source = model_override`
- Admin 创建单用户 direct adjustment，并能从调整列表和 audit logs 中查到
- 普通用户重新拉取 `/models` 后，能看到被 user group 放行的模型
- 普通用户通过 `POST /api/v1/chat/completions` 发起真实非流式请求，服务端按数据库中的 `upstream + channel + route_binding` 解析上游
- 普通用户通过 `POST /api/v1/chat/completions` + `stream=true` 发起真实 SSE 请求，已收到 `response.start`、`reasoning.delta`、`response.completed` 和 `[DONE]`
- 管理员通过 `GET /api/v1/admin/upstreams/:id/discovered-models` 已能从局域网 `newapi` 拉取候选模型，并识别本地已导入模型
- 管理员通过 `POST /api/v1/admin/models/import` 已能对已存在模型返回 `skipped_existing`，避免重复创建
- 管理员通过 `GET /api/v1/admin/models/:id` 已能拿到 `visible_user_groups`、`visibility_summary` 和 hydrated `route_bindings`
- 管理员通过 `GET /api/v1/admin/channels/:id` 与 `GET /api/v1/admin/user-groups/:id` 已能读取单资源详情
- 上游读接口当前会对 `auth_config` 中的敏感字段做脱敏，不再直接回传明文 `api_key`
- 会话、消息、usage 和 `llm_request_logs` 已成功落库，`limit_usage` 已体现请求次数与 token 消耗增量
- Chat 最小联调页已切到 `fetch + SSE`，并支持最基本的“停止生成”

本轮烟雾验证的关键结果：

- `policy_count = 2`
- `policy_source = model_override`
- `visible_model_count >= 2`
- `request_log_count = 3`
- `latest_request_status = completed`
- `latest_request_total_tokens = 539`
- `stream_request_status = completed`
- `stream_finish_reason = length`
- `stream_total_tokens = 109`
- `discovered_model_total = 2`
- `discovery_already_imported = 2`
- `import_requested = 1`
- `import_skipped_existing = 1`
- `model_detail_visibility_summary = lan-group-1773811857`
- `model_detail_route_summary = lan-channel-1773811857 -> lan-newapi-1773811857 (priority 1)`
- `masked_upstream_auth_config = true`
- `remaining_hour_requests = 2`
- `adjustment_count = 1`
- `model_visible = true`
- `audit_total = 1`

本轮用于真实联调的上游信息：

- 上游类型：OpenAI 兼容接口
- 目标服务：局域网 `newapi`
- Base URL：`172.16.99.204:3398`
- 当前验证到的可用模型：`Qwen/Qwen3.5-122B-A10B`

当前已知现象：

- 该上游在当前测试模型下会稳定返回 `reasoning_content`
- 当 `max_tokens` 较小甚至较大时，响应可能在推理阶段被截断，出现 `finish_reason = length` 且 `content = ""`
- 这不影响当前“路由 -> 请求 -> SSE -> 落库 -> 限额统计”闭环验证，但后续需要继续做模型参数策略和更细的前端 `reasoning_content` 展示优化

## 4. 当前任务映射

### 4.1 可视为已完成或已达到当前阶段验收

- `INF-01`
- `INF-02`
- `INF-03`
- `INF-04`
- `INF-06`
- `INF-07`
- `DB-01`
- `DB-02`
- `DB-03`
- `DB-04`
- `AUTH-BE-01`
- `AUTH-BE-02`
- `USER-BE-01`
- `USER-BE-02`
- `USER-BE-03`
- `AUTH-FE-01`
- `APP-FE-01`
- `USER-FE-01`
- `USER-FE-02`
- `GROUP-BE-01`
- `LIMIT-BE-01`
- `LIMIT-BE-02`
- `ADMIN-BE-03`
- `ADMIN-BE-04`
- `MODEL-BE-01`
- `CHAT-BE-01`
- `CHAT-BE-02`
- `CHAT-BE-03`
- `CHAT-BE-05`

### 4.2 已进入开发，但还未完全达到原始清单定义

- `INF-08`
  - 已有 Redis 配置与可关闭能力，但缓存/限流/冷却降级策略尚未完整落地
- `AUTH-BE-03`
  - 认证主链路已完成，登录失败限制与风控日志仍待补齐
- `ADMIN-BE-01`
  - 已完成列表/创建/更新，删除与更细的资源管理尚未补齐
- `ADMIN-BE-02`
  - 已完成列表/创建/更新与路由绑定保存，但更细的资源管理、删除与后续 P1 配置还未补齐
- `ADMIN-BE-05`
  - 已完成第一版：支持基于 upstream 的模型发现，并返回候选模型和本地已导入状态
- `ADMIN-BE-06`
  - 已完成第一版：支持模型导入接口，并在 admin 模型返回中补 `visible_user_groups`、hydrated `route_bindings` 和 human-readable summary
- `ADMIN-BE-07`
  - 已完成第一版：上游/渠道/模型/用户组详情接口已补齐；后续仍需继续补停用、删除前检查和编辑体验
- `CHAT-BE-04`
  - 当前已按数据库配置选择可用 route binding 并完成单路由调用，但还没有真正做多上游故障切换、冷却窗口和 Redis 缓存
- `CHAT-BE-05`
  - 第一版已完成：非流式、SSE 和客户端断连取消均已打通；独立 stop API 与更细的前端交互仍待后续补强
- `CHAT-BE-06`
  - 已实现请求消息与 assistant 消息落库，但消息状态机和中断场景仍待扩展
- `CHAT-BE-07`
  - 已优先使用上游 usage，估算仅作为回退；尚未进入正式计费闭环
- `CHAT-BE-08`
  - `llm_request_logs` 已写入成功/失败/超限结果，但路由明细、聚合报表与更完整错误码体系仍待补齐

## 5. 下一步建议

最顺的推进顺序现在是：

1. 继续推进 `CHAT-BE-04`，补多上游 Failover、冷却与 Redis 缓存
2. 完成 `CHAT-BE-07` + `BILL-BE-01`，把 usage 与计费账本打通
3. 补强 `CHAT-BE-08` 的路由日志、错误码和报表聚合能力
4. 继续完善前端对 `reasoning_content`、长流式消息和 stop 反馈的体验细节
