# MrChat 实现进度快照

- 状态：首轮实现进行中
- 日期：2026-03-17
- 当前阶段：`M0/M1` 已基本落地，`M2/M3` 正在推进

## 1. 当前结论

仓库已经不再停留在纯需求分析阶段，当前已经进入“可持续开发 + 可联调验证”的实现阶段。

截至 2026-03-17，项目状态可以概括为：

- 后端工程骨架、配置加载、goose 迁移入口、基础鉴权和用户中心 API 已可用
- 核心表结构已经落到 PostgreSQL，并支持服务启动自动迁移
- 管理侧已具备上游、模型、用户调额、审计日志的首版 API
- Chat 侧已具备模型列表、会话 CRUD、消息分页查询的首版 API
- 前端已提供无样式联调壳子，可直接对接登录、设置、用量、Chat 和部分 Admin 页面

## 2. 已落地能力

### 2.1 工程与数据层

- 后端入口：`cmd/api`
- 迁移入口：`cmd/migrate`
- goose SQL 已建立并应用到：
  - `users` / `auths` / `groups` / `group_members`
  - `upstreams` / `models` / `model_route_bindings`
  - `conversations` / `messages`
  - `quota_logs` / `redeem_codes` / `redeem_redemptions` / `audit_logs`
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
  - `GET /api/v1/admin/models`
  - `POST /api/v1/admin/models`
  - `PUT /api/v1/admin/models/:id`
  - `GET /api/v1/admin/users`
  - `PUT /api/v1/admin/users/:id/quota`
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
- `/admin/models`
- `/admin/users`
- `/admin/audit-logs`

## 3. 已验证结果

### 3.1 命令级验证

- `go test ./...` 通过
- `pnpm typecheck` 通过
- `pnpm build` 通过
- `go run ./cmd/migrate up` 可对局域网 PostgreSQL 执行迁移
- `go run ./cmd/api` 启动时可自动对齐数据库版本

### 3.2 真实链路验证

已在局域网 PostgreSQL 环境验证通过以下链路：

- 注册 -> 登录 -> 获取当前用户
- 用户资料 / 安全设置 / 用量摘要查询
- Admin 创建 upstream
- Admin 创建 model 并绑定 upstream
- 普通用户获取 `/models`
- 普通用户创建 conversation
- 普通用户查询 conversation messages
- Admin 调整用户 quota
- Admin 查询 audit logs

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
  - 已完成列表/创建/更新与路由绑定保存，删除与组管理配套能力仍待补齐
- `GROUP-BE-01`
  - 尚未正式开始
- `CHAT-BE-03` ~ `CHAT-BE-08`
  - 目前只到会话与消息读取层，尚未进入真实上游调用、SSE、路由、结算与请求追踪

## 5. 下一步建议

最顺的推进顺序仍然是：

1. 补 `GROUP-BE-01`，把模型可见性和用户组维护闭环补起来
2. 开始 `CHAT-BE-03`，先接通一个 OpenAI 兼容上游的非流式请求
3. 再推进 `CHAT-BE-04` ~ `CHAT-BE-08`，完成路由、SSE、持久化、usage 与日志闭环
4. 随后进入 `M5` 的兑换码与计费闭环
