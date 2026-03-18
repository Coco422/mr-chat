# MrChat 实现进度快照

- 状态：首轮实现进行中
- 日期：2026-03-18
- 当前阶段：`M0/M1` 已完成，`M2` 大部分已落地，`M3` 正在进入真实聊天能力实现

## 1. 当前结论

仓库已经不是“只有设计文档”的状态，当前已经进入“可持续开发 + 可联调验证”的实现阶段。

截至 2026-03-18，项目状态可以概括为：

- 后端工程骨架、配置加载、goose 迁移入口、基础鉴权和用户中心 API 已可用
- 核心表结构已经落到 PostgreSQL，并支持服务启动自动迁移
- 管理侧已具备上游、渠道、模型、用户组、用户调额、用户限额调整、审计日志的首版 API
- Chat 侧已具备模型列表、会话 CRUD、消息分页查询；限额引擎和请求日志结构已落地，但真实 `/chat/completions` 上游调用尚未接通
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
- Admin
  - `GET /api/v1/admin/upstreams`
  - `POST /api/v1/admin/upstreams`
  - `PUT /api/v1/admin/upstreams/:id`
  - `GET /api/v1/admin/channels`
  - `POST /api/v1/admin/channels`
  - `PUT /api/v1/admin/channels/:id`
  - `GET /api/v1/admin/models`
  - `POST /api/v1/admin/models`
  - `PUT /api/v1/admin/models/:id`
  - `GET /api/v1/admin/user-groups`
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
- `pnpm typecheck` 通过
- `pnpm build` 通过
- `go run ./cmd/migrate status` 能识别局域网 PostgreSQL 上的 pending migration
- `go run ./cmd/migrate up` 已将局域网 PostgreSQL 升到 version `6`
- `go run ./cmd/api` 启动时日志显示 `goose: no migrations to run. current version: 6`

### 3.2 局域网 PostgreSQL 烟雾验证

已在局域网 PostgreSQL 环境验证通过以下链路：

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

本轮烟雾验证的关键结果：

- `policy_count = 2`
- `policy_source = model_override`
- `remaining_hour_requests = 10`
- `adjustment_count = 1`
- `model_visible = true`
- `audit_total = 1`

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

### 4.2 已进入开发，但还未完全达到原始清单定义

- `INF-08`
  - 已有 Redis 配置与可关闭能力，但缓存/限流/冷却降级策略尚未完整落地
- `AUTH-BE-03`
  - 认证主链路已完成，登录失败限制与风控日志仍待补齐
- `ADMIN-BE-01`
  - 已完成列表/创建/更新，删除与更细的资源管理尚未补齐
- `ADMIN-BE-02`
  - 已完成列表/创建/更新与路由绑定保存，但更细的资源管理、删除与后续 P1 配置还未补齐
- `CHAT-BE-03` ~ `CHAT-BE-08`
  - 当前只到会话与消息读取层；真实上游调用、SSE、请求日志写入、usage 采集与结算闭环还在后续里程碑中

## 5. 下一步建议

最顺的推进顺序现在是：

1. 开始 `CHAT-BE-03`，先接通一个 OpenAI 兼容上游的非流式请求
2. 紧接着推进 `CHAT-BE-05`，落第一版 `/api/v1/chat/completions`
3. 把已完成的限额引擎和 `llm_request_logs` 正式接入聊天前置校验与请求结果落库
4. 再进入 `CHAT-BE-04`、`CHAT-BE-07`、`CHAT-BE-08`，补路由、usage 回退和日志闭环
