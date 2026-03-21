# MrChat v0.1 API 契约

- 状态：实现设计草案
- 日期：2026-03-18
- 依赖基线：`docs/Requirements-Baseline-v0.1.md`

## 1. 目标

这份文档定义 v0.1 的 API 约定，供前后端并行开发时对齐请求、响应、鉴权、错误码与 SSE 流式格式。

## 2. 全局约定

### 2.1 基础路径

- 所有业务 API 统一挂在 `/api/v1`
- 开发联调时可通过 `/swagger/index.html` 查看当前运行服务生成的 Swagger UI

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

这是用户侧主聊天入口。当前实现已经支持非流式和 `stream=true` 的 SSE 流式请求。

请求：

```json
{
  "conversation_id": "uuid",
  "model_id": "uuid",
  "stream": false,
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
- `model_id` 当前为必填；若传已有 `conversation_id` 且该会话已绑定模型，则服务端可回退使用会话上的模型
- `messages` 至少包含 1 条新的 user 消息
- 对于已有会话，前端可以只传本次新增消息；服务端负责结合持久化历史重建上下文
- 当前服务端会先做用户分组 + 模型限额校验，再按数据库中的 `route_bindings -> upstream` 配置发起 OpenAI 兼容请求

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

### 当前已实现行为

- 非流式请求会：
  - 校验模型可见性
  - 读取当前用户的有效限额模板与直接调整值
  - 结合历史消息重建上游上下文
  - 按数据库里的上游和路由绑定发起 OpenAI 兼容请求
  - 持久化 user/assistant 消息
  - 落库 `llm_request_logs`
- 流式请求会：
  - 在校验通过后立即返回 `text/event-stream`
  - 先发送 `response.start`
  - 按上游增量分别透传 `response.delta.content` 与 `reasoning.delta.reasoning_content`
  - 完成后发送 `response.completed` 和 `data: [DONE]`
  - 客户端主动断开时中止上游请求，并把消息与 `llm_request_logs` 标记为取消
- 账本结算字段当前固定返回 `0`，后续由 `BILL-BE-01` 接入
- 某些推理模型可能返回 `reasoning_content` 但 `content` 为空；前端需要同时兼容这两个字段

### 流式响应

响应头：

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- 如果在发出任何 SSE 事件之前就校验失败，则仍返回标准 JSON 错误包，而不是事件流

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
data: {"type":"response.completed","request_id":"req_123","conversation_id":"uuid","assistant_message_id":"uuid","usage":{"prompt_tokens":120,"completion_tokens":80,"total_tokens":200},"billing":{"pre_deducted":0,"final_charged":0,"refunded":0},"finish_reason":"stop"}

```

#### `response.failed`

```text
data: {"type":"response.failed","request_id":"req_123","conversation_id":"uuid","assistant_message_id":"uuid","error":{"code":"CHAT_UPSTREAM_UNAVAILABLE","message":"streaming failed"}}

```

结束标志：

```text
data: [DONE]

```

### 停止生成

- v0.1 以“客户端主动断开 SSE 连接”作为停止生成主机制
- 服务端当前已经在连接断开后中止上游请求，并把请求日志与消息状态更新为取消
- 独立的 stop API 仍属于后续演进项

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

## 11.6 `GET /api/v1/admin/upstreams/:id`

- 返回单个 upstream 详情
- 当前读接口会对 `auth_config` 中的敏感字段做脱敏展示，例如：
  - `api_key`
  - `token`
  - `password`

## 11.7 `GET /api/v1/admin/upstreams/:id/discovered-models`

- 从指定 upstream 的 `/v1/models` 拉取候选模型
- 当前第一版按 OpenAI 兼容接口实现
- 返回值会标记：
  - `already_imported`
  - `existing_model`

响应示例：

```json
{
  "success": true,
  "data": {
    "upstream": {
      "id": "uuid",
      "name": "LAN newapi",
      "provider_type": "openai",
      "status": "active"
    },
    "items": [
      {
        "model_key": "Qwen/Qwen3.5-122B-A10B",
        "display_name": "Qwen/Qwen3.5-122B-A10B",
        "provider_type": "openai",
        "supported_endpoint_types": ["openai"],
        "already_imported": true,
        "existing_model": {
          "id": "uuid",
          "model_key": "Qwen/Qwen3.5-122B-A10B",
          "display_name": "Qwen/Qwen3.5-122B-A10B",
          "status": "active"
        }
      }
    ],
    "fetched_at": "2026-03-21T08:19:41Z",
    "summary": {
      "total": 2,
      "already_imported": 1,
      "new_candidates": 1
    }
  },
  "request_id": "req_123"
}
```

## 11.8 `GET /api/v1/admin/channels`

- 返回渠道列表

## 11.9 `GET /api/v1/admin/channels/:id`

- 返回单个渠道详情

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

## 11.12 `GET /api/v1/admin/models`

- 返回模型与其路由配置摘要
- 当前返回值除了保留写路径字段：
  - `visible_user_group_ids`
  - `route_bindings[].channel_id`
  - `route_bindings[].upstream_id`
- 也会额外补充人类可读字段：
  - `visible_user_groups`
  - `visibility_summary`
  - `route_bindings[].channel`
  - `route_bindings[].upstream`
  - `route_bindings[].summary`
  - `route_rule_summaries`

响应片段：

```json
{
  "id": "uuid",
  "model_key": "Qwen/Qwen3.5-122B-A10B",
  "display_name": "Qwen/Qwen3.5-122B-A10B",
  "visibility_summary": "vip-users",
  "visible_user_group_ids": ["uuid"],
  "visible_user_groups": [
    {
      "id": "uuid",
      "name": "vip-users",
      "status": "active"
    }
  ],
  "route_bindings": [
    {
      "id": "uuid",
      "channel_id": "uuid",
      "upstream_id": "uuid",
      "priority": 1,
      "status": "active",
      "channel": {
        "id": "uuid",
        "name": "default"
      },
      "upstream": {
        "id": "uuid",
        "name": "LAN newapi",
        "status": "active"
      },
      "summary": "default -> LAN newapi (priority 1)"
    }
  ]
}
```

## 11.13 `GET /api/v1/admin/models/:id`

- 返回单个模型详情
- 返回体结构与 `GET /api/v1/admin/models` 中的单项对象一致

## 11.14 `POST /api/v1/admin/models`

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

## 11.15 `POST /api/v1/admin/models/import`

- 用于把“已发现的 upstream 模型”导入到本地模型目录
- 第一版由前端把已发现的候选模型字段回传给后端；后端不会在导入时再次强制请求 upstream
- 若 `model_key` 已存在，则不会报错，而是返回 `status = skipped_existing`

请求：

```json
{
  "upstream_id": "uuid",
  "items": [
    {
      "model_key": "Qwen/Qwen3.5-122B-A10B",
      "display_name": "Qwen/Qwen3.5-122B-A10B",
      "provider_type": "openai",
      "channel_id": "uuid",
      "visible_user_group_ids": ["uuid"],
      "status": "active",
      "capabilities": {
        "chat": true,
        "streaming": true
      },
      "priority": 1
    }
  ]
}
```

响应片段：

```json
{
  "success": true,
  "data": {
    "upstream": {
      "id": "uuid",
      "name": "LAN newapi"
    },
    "items": [
      {
        "requested_model_key": "Qwen/Qwen3.5-122B-A10B",
        "status": "skipped_existing",
        "existing_model": {
          "id": "uuid",
          "model_key": "Qwen/Qwen3.5-122B-A10B",
          "display_name": "Qwen/Qwen3.5-122B-A10B",
          "status": "active"
        },
        "model": null
      }
    ],
    "summary": {
      "requested": 1,
      "imported": 0,
      "skipped_existing": 1
    }
  },
  "request_id": "req_123"
}
```

## 11.16 `PUT /api/v1/admin/models/:id`

- 用于修改模型展示、定价、可见性与路由绑定

## 11.17 `GET /api/v1/admin/user-groups`

- 返回用户分组列表

## 11.18 `GET /api/v1/admin/user-groups/:id`

- 返回单个用户分组详情

## 11.19 `POST /api/v1/admin/user-groups`

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

## 11.20 `PUT /api/v1/admin/user-groups/:id`

- 用于修改分组名称、描述、状态和权限

## 11.21 `GET /api/v1/admin/user-groups/:id/limits`

- 返回该用户分组的默认模板与模型覆盖规则列表

## 11.22 `PUT /api/v1/admin/user-groups/:id/limits`

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

## 11.23 `PUT /api/v1/admin/users/:id/group`

请求：

```json
{
  "user_group_id": "uuid"
}
```

## 11.24 `GET /api/v1/admin/users/:id/limit-usage`

查询参数：

- `model_id`：可选；为空时按全部模型汇总

响应要点：

- 返回 `effective_policy`
- 返回 `usage.hour/week/lifetime`
- 返回 `adjustments.hour/week/lifetime`
- 返回 `remaining.hour/week/lifetime`

## 11.25 `GET /api/v1/admin/users/:id/limit-adjustments`

查询参数：

- `page`
- `page_size`
- `model_id`

## 11.26 `POST /api/v1/admin/users/:id/limit-adjustments`

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

## 11.27 `POST /api/v1/admin/redeem-codes/batch`

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

## 11.28 `GET /api/v1/admin/redeem-codes`

- 返回兑换码批次与统计信息

## 11.29 `GET /api/v1/admin/audit-logs`

查询参数：

- `page`
- `page_size`
- `actor_id`
- `action`
- `resource_type`
- `result`

## 11.30 `GET /api/v1/admin/service-entries`（P1）

- 返回服务入口配置列表

## 11.31 `POST /api/v1/admin/service-entries`（P1）

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

## 11.32 `PUT /api/v1/admin/service-entries/:id`（P1）

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
