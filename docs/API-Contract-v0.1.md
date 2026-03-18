# MrChat v0.1 API 契约

- 状态：实现设计草案
- 日期：2026-03-18
- 依赖基线：`docs/Requirements-Baseline-v0.1.md`

## 1. 目标

这份文档定义 v0.1 的 API 约定，供前后端并行开发时对齐请求、响应、鉴权、错误码与 SSE 流式格式。

## 2. 全局约定

### 2.1 基础路径

- 所有业务 API 统一挂在 `/api/v1`

### 2.2 鉴权

- 访问受保护接口时，使用 `Authorization: Bearer <access_token>`
- 登录成功后：
  - 响应体返回 `access_token`
  - 服务端设置 `refresh_token` 为 `httpOnly` Cookie
- `/auth/refresh` 用于刷新 `access_token`

### 2.3 通用响应格式

普通 JSON 接口统一返回：

```json
{
  "success": true,
  "data": {},
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 100
  },
  "request_id": "req_123"
}
```

失败时：

```json
{
  "success": false,
  "error": {
    "code": "AUTH_INVALID_CREDENTIALS",
    "message": "Invalid credentials",
    "details": null
  },
  "request_id": "req_123"
}
```

### 2.4 时间与 ID

- 所有资源 ID 使用 UUID 字符串
- 时间字段统一使用 ISO 8601 字符串或统一的 UTC 时间戳
- v0.1 推荐返回 ISO 8601 字符串，便于前端直接显示与调试

### 2.5 分页与排序

- 列表接口默认使用：
  - `page`
  - `page_size`
  - `sort_by`
  - `sort_order`
- 未显式说明时，默认按创建时间倒序

## 3. 资源模型

### 3.1 User

```json
{
  "id": "uuid",
  "username": "ray",
  "email": "ray@example.com",
  "display_name": "Ray",
  "avatar_url": null,
  "role": "user",
  "status": "active",
  "quota": 1200000,
  "used_quota": 450000,
  "settings": {
    "timezone": "Asia/Shanghai",
    "locale": "zh-CN"
  },
  "created_at": "2026-03-12T09:00:00Z",
  "updated_at": "2026-03-12T09:00:00Z"
}
```

### 3.2 Model

```json
{
  "id": "uuid",
  "model_key": "gpt-4o-mini",
  "display_name": "GPT-4o Mini",
  "provider_type": "openai_compatible",
  "context_length": 128000,
  "pricing": {
    "input_price": 0.15,
    "output_price": 0.6,
    "currency": "USD"
  },
  "capabilities": {
    "streaming": true,
    "vision": false,
    "function_calling": false
  },
  "visible_user_group_ids": ["uuid"],
  "status": "active"
}
```

### 3.3 Channel

```json
{
  "id": "uuid",
  "name": "default-openai",
  "description": "default billing channel",
  "status": "active",
  "billing_config": {
    "currency": "USD"
  }
}
```

### 3.4 UserGroup

```json
{
  "id": "uuid",
  "name": "free",
  "description": "free tier users",
  "status": "active",
  "permissions": {},
  "metadata": {}
}
```

### 3.5 UserGroupLimitPolicy

```json
{
  "id": "uuid",
  "user_group_id": "uuid",
  "model_id": null,
  "hour_request_limit": 30,
  "week_request_limit": 500,
  "lifetime_request_limit": null,
  "hour_token_limit": 200000,
  "week_token_limit": 1000000,
  "lifetime_token_limit": null,
  "status": "active"
}
```

### 3.6 ConversationSummary

```json
{
  "id": "uuid",
  "title": "New conversation",
  "model_id": "uuid",
  "last_message_at": "2026-03-12T09:00:00Z",
  "message_count": 12,
  "status": "active"
}
```

### 3.7 Message

```json
{
  "id": "uuid",
  "conversation_id": "uuid",
  "role": "assistant",
  "content": "Hello",
  "reasoning_content": "",
  "status": "completed",
  "finish_reason": "stop",
  "usage": {
    "prompt_tokens": 120,
    "completion_tokens": 80,
    "total_tokens": 200
  },
  "created_at": "2026-03-12T09:00:00Z"
}
```

### 3.8 UserLimitAdjustment

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "model_id": null,
  "metric_type": "request_count",
  "window_type": "rolling_hour",
  "delta": 10,
  "expires_at": "2026-03-18T10:00:00Z",
  "reason": "support grant",
  "actor_user_id": "uuid",
  "created_at": "2026-03-18T09:00:00Z"
}
```

### 3.9 ServiceEntry（P1）

```json
{
  "id": "uuid",
  "name": "Grok Mirror",
  "slug": "grok-mirror",
  "description": "Experience service",
  "launch_mode": "iframe",
  "entry_url": "https://service.example.com",
  "icon_url": null,
  "badge": "trial",
  "status": "active"
}
```

## 4. Auth API

## 4.1 `POST /api/v1/auth/signup`

请求：

```json
{
  "username": "ray",
  "email": "ray@example.com",
  "password": "secret"
}
```

响应：

```json
{
  "success": true,
  "data": {
    "access_token": "jwt",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "username": "ray",
      "email": "ray@example.com",
      "role": "user"
    }
  },
  "request_id": "req_123"
}
```

## 4.2 `POST /api/v1/auth/signin`

请求：

```json
{
  "identifier": "ray@example.com",
  "password": "secret"
}
```

响应：

```json
{
  "success": true,
  "data": {
    "access_token": "jwt",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "username": "ray",
      "email": "ray@example.com",
      "role": "user"
    }
  },
  "request_id": "req_123"
}
```

## 4.3 `POST /api/v1/auth/signout`

- 清除刷新 Cookie
- 前端删除本地 access token

## 4.4 `POST /api/v1/auth/refresh`

响应：

```json
{
  "success": true,
  "data": {
    "access_token": "new_jwt",
    "expires_in": 3600,
    "user": {
      "id": "uuid",
      "username": "ray",
      "email": "ray@example.com",
      "role": "user"
    }
  },
  "request_id": "req_123"
}
```

## 5. 用户与用量 API

## 5.1 `GET /api/v1/users/me`

- 返回当前用户资料

## 5.2 `PUT /api/v1/users/me`

请求：

```json
{
  "display_name": "Ray",
  "avatar_url": "https://example.com/avatar.png",
  "settings": {
    "timezone": "Asia/Shanghai",
    "locale": "zh-CN"
  }
}
```

## 5.3 `GET /api/v1/users/me/security`

响应：

```json
{
  "success": true,
  "data": {
    "last_login_at": "2026-03-12T09:00:00Z",
    "password_updated_at": "2026-03-12T09:00:00Z",
    "has_password": true
  },
  "request_id": "req_123"
}
```

## 5.4 `PUT /api/v1/users/me/password`

请求：

```json
{
  "current_password": "secret-old",
  "new_password": "secret-new"
}
```

## 5.5 `GET /api/v1/users/me/quota`

响应：

```json
{
  "success": true,
  "data": {
    "quota": 1200000,
    "used_quota": 450000,
    "remaining_quota": 1200000
  },
  "request_id": "req_123"
}
```

说明：

- `quota` 表示当前余额
- `remaining_quota` 在 v0.1 作为余额别名返回，便于前端语义对齐

## 5.6 `GET /api/v1/users/me/usage`

查询参数：

- `range=7d|30d|month`

响应：

```json
{
  "success": true,
  "data": {
    "summary": {
      "total_spent_tokens": 450000,
      "spent_today": 12000,
      "spent_in_range": 87000
    },
    "daily": [
      {
        "date": "2026-03-12",
        "spent_tokens": 12000
      }
    ]
  },
  "request_id": "req_123"
}
```

## 6. 模型 API

## 6.1 `GET /api/v1/models`

- 返回当前用户有权限看到的模型列表

响应：

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "model_key": "gpt-4o-mini",
      "display_name": "GPT-4o Mini"
    }
  ],
  "request_id": "req_123"
}
```

## 7. Conversation API

## 7.1 `GET /api/v1/conversations`

查询参数：

- `page`
- `page_size`
- `status=active|archived`

响应：

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "title": "New conversation",
      "last_message_at": "2026-03-12T09:00:00Z",
      "message_count": 12,
      "status": "active"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 1
  },
  "request_id": "req_123"
}
```

## 7.2 `POST /api/v1/conversations`

请求：

```json
{
  "title": "New conversation",
  "model_id": "uuid"
}
```

## 7.3 `PUT /api/v1/conversations/:id`

请求：

```json
{
  "title": "Renamed conversation"
}
```

## 7.4 `DELETE /api/v1/conversations/:id`

- 软删除会话

## 7.5 `GET /api/v1/conversations/:id/messages`

查询参数：

- `page`
- `page_size`

响应：

```json
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "role": "user",
      "content": "Hello",
      "reasoning_content": "",
      "status": "completed",
      "created_at": "2026-03-12T09:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 50,
    "total": 20
  },
  "request_id": "req_123"
}
```

## 8. Chat API

## 8.1 `POST /api/v1/chat/completions`

这是用户侧主聊天入口，支持非流式与流式。

请求：

```json
{
  "conversation_id": "uuid",
  "model_id": "uuid",
  "stream": true,
  "messages": [
    {
      "role": "user",
      "content": "Explain SSE simply."
    }
  ],
  "metadata": {
    "source": "web"
  }
}
```

约束：

- `conversation_id` 可为空；为空时服务端创建新会话
- `messages` 至少包含 1 条新的 user 消息
- 对于已有会话，前端可以只传本次新增消息；服务端负责结合持久化历史重建上下文

### 非流式响应

```json
{
  "success": true,
  "data": {
    "conversation_id": "uuid",
    "message": {
      "id": "uuid",
      "role": "assistant",
      "content": "SSE is..."
    },
    "usage": {
      "prompt_tokens": 120,
      "completion_tokens": 80,
      "total_tokens": 200
    },
    "billing": {
      "pre_deducted": 300,
      "final_charged": 200,
      "refunded": 100
    }
  },
  "request_id": "req_123"
}
```

### 流式响应

响应头：

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

事件格式：

#### `response.start`

```text
data: {"type":"response.start","request_id":"req_123","conversation_id":"uuid","assistant_message_id":"uuid"}

```

#### `response.delta`

```text
data: {"type":"response.delta","delta":{"content":"SSE "}}

```

#### `reasoning.delta`

```text
data: {"type":"reasoning.delta","delta":{"reasoning_content":"thinking..."}}

```

#### `response.completed`

```text
data: {"type":"response.completed","usage":{"prompt_tokens":120,"completion_tokens":80,"total_tokens":200},"billing":{"pre_deducted":300,"final_charged":200,"refunded":100},"finish_reason":"stop"}

```

#### `error`

```text
data: {"type":"error","error":{"code":"CHAT_UPSTREAM_UNAVAILABLE","message":"Upstream unavailable"}}

```

结束标志：

```text
data: [DONE]

```

### 停止生成

- v0.1 以“客户端主动断开 SSE 连接”作为停止生成主机制
- 服务端必须在连接断开后中止上游请求并进入结算/退款流程

## 9. Billing API

## 9.1 `POST /api/v1/billing/redeem`

请求：

```json
{
  "code": "ABCD-EFGH-IJKL"
}
```

响应：

```json
{
  "success": true,
  "data": {
    "redeemed_quota": 500000,
    "remaining_quota": 1250000
  },
  "request_id": "req_123"
}
```

## 9.2 `GET /api/v1/billing/logs`

查询参数：

- `page`
- `page_size`
- `type=consume|refund|redeem|admin_adjust`

## 9.3 `GET /api/v1/billing/summary`

响应：

```json
{
  "success": true,
  "data": {
    "remaining_quota": 1250000,
    "consumed_total": 450000,
    "redeemed_total": 1000000
  },
  "request_id": "req_123"
}
```

## 10. Service Entry API（P1）

## 10.1 `GET /api/v1/service-entries`

- 返回当前用户可见的服务入口

## 11. Admin API

## 11.1 `GET /api/v1/admin/users`

查询参数：

- `page`
- `page_size`
- `keyword`
- `status`

## 11.2 `PUT /api/v1/admin/users/:id/quota`

请求：

```json
{
  "delta": 500000,
  "reason": "manual_adjust"
}
```

约束：

- `delta` 可正可负
- 必须写入 `quota_logs` 与 `audit_logs`

## 11.3 `GET /api/v1/admin/upstreams`

- 返回上游列表

## 11.4 `POST /api/v1/admin/upstreams`

请求：

```json
{
  "name": "OpenAI Relay A",
  "provider_type": "openai_compatible",
  "base_url": "https://relay.example.com/v1",
  "auth_type": "bearer",
  "auth_config": {
    "api_key": "encrypted-or-writeonly"
  },
  "status": "active"
}
```

## 11.5 `PUT /api/v1/admin/upstreams/:id`

- 用于修改状态、密钥、备注、超时与冷却参数

## 11.6 `GET /api/v1/admin/models`

- 返回模型与其路由配置摘要

## 11.7 `POST /api/v1/admin/models`

请求：

```json
{
  "model_key": "gpt-4o-mini",
  "display_name": "GPT-4o Mini",
  "context_length": 128000,
  "pricing": {
    "input_price": 0.15,
    "output_price": 0.6,
    "currency": "USD"
  },
  "visible_user_group_ids": ["uuid"],
  "route_bindings": [
    {
      "channel_id": "uuid",
      "upstream_id": "uuid",
      "priority": 1
    }
  ]
}
```

## 11.8 `PUT /api/v1/admin/models/:id`

- 用于修改模型展示、定价、可见性与路由绑定

## 11.9 `GET /api/v1/admin/channels`

- 返回渠道列表

## 11.10 `POST /api/v1/admin/channels`

请求：

```json
{
  "name": "default-openai",
  "description": "default billing channel",
  "status": "active",
  "billing_config": {
    "currency": "USD"
  },
  "metadata": {}
}
```

## 11.11 `PUT /api/v1/admin/channels/:id`

- 用于修改渠道状态、描述与计费配置

## 11.12 `GET /api/v1/admin/user-groups`

- 返回用户分组列表

## 11.13 `POST /api/v1/admin/user-groups`

请求：

```json
{
  "name": "free",
  "description": "free tier users",
  "status": "active",
  "permissions": {},
  "metadata": {}
}
```

## 11.14 `PUT /api/v1/admin/user-groups/:id`

- 用于修改分组名称、描述、状态和权限

## 11.15 `GET /api/v1/admin/user-groups/:id/limits`

- 返回该用户分组的默认模板与模型覆盖规则列表

## 11.16 `PUT /api/v1/admin/user-groups/:id/limits`

请求：

```json
{
  "policies": [
    {
      "model_id": null,
      "hour_request_limit": 30,
      "week_request_limit": 500,
      "lifetime_request_limit": null,
      "hour_token_limit": 200000,
      "week_token_limit": 1000000,
      "lifetime_token_limit": null,
      "status": "active"
    }
  ]
}
```

## 11.17 `PUT /api/v1/admin/users/:id/group`

请求：

```json
{
  "user_group_id": "uuid"
}
```

## 11.18 `GET /api/v1/admin/users/:id/limit-usage`

查询参数：

- `model_id`：可选；为空时按全部模型汇总

响应要点：

- 返回 `effective_policy`
- 返回 `usage.hour/week/lifetime`
- 返回 `adjustments.hour/week/lifetime`
- 返回 `remaining.hour/week/lifetime`

## 11.19 `GET /api/v1/admin/users/:id/limit-adjustments`

查询参数：

- `page`
- `page_size`
- `model_id`

## 11.20 `POST /api/v1/admin/users/:id/limit-adjustments`

请求：

```json
{
  "model_id": null,
  "metric_type": "request_count",
  "window_type": "rolling_hour",
  "delta": 10,
  "reason": "support grant"
}
```

## 11.21 `POST /api/v1/admin/redeem-codes/batch`

请求：

```json
{
  "quota_amount": 500000,
  "count": 100,
  "valid_until": "2026-06-30T00:00:00Z",
  "max_redemptions": 1,
  "batch_no": "20260312-A"
}
```

## 11.22 `GET /api/v1/admin/redeem-codes`

- 返回兑换码批次与统计信息

## 11.23 `GET /api/v1/admin/audit-logs`

查询参数：

- `page`
- `page_size`
- `actor_id`
- `action`
- `resource_type`
- `result`

## 11.24 `GET /api/v1/admin/service-entries`（P1）

- 返回服务入口配置列表

## 11.25 `POST /api/v1/admin/service-entries`（P1）

请求：

```json
{
  "name": "Grok Mirror",
  "slug": "grok-mirror",
  "entry_url": "https://service.example.com",
  "launch_mode": "iframe",
  "visible_user_group_ids": ["uuid"],
  "status": "active"
}
```

## 11.26 `PUT /api/v1/admin/service-entries/:id`（P1）

- 用于修改地址、模式、状态和可见组

## 12. 错误码建议

### 12.1 Auth

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_TOKEN_EXPIRED`
- `AUTH_TOKEN_INVALID`
- `AUTH_FORBIDDEN`

### 12.2 Chat

- `CHAT_MODEL_NOT_AVAILABLE`
- `CHAT_UPSTREAM_UNAVAILABLE`
- `CHAT_STREAM_ABORTED`
- `CHAT_CONTEXT_INVALID`
- `CHAT_LIMIT_EXCEEDED`

### 12.3 Billing

- `BILLING_INSUFFICIENT_QUOTA`
- `BILLING_PREDEDUCT_FAILED`
- `BILLING_SETTLEMENT_FAILED`
- `REDEEM_INVALID_OR_USED`

### 12.4 Admin

- `ADMIN_RESOURCE_NOT_FOUND`
- `ADMIN_INVALID_STATUS_TRANSITION`
- `ADMIN_SENSITIVE_OPERATION_REQUIRES_AUDIT`
