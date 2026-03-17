# MrChat 本地开发说明

- 状态：开发联调用
- 日期：2026-03-17

## 1. 目的

这份文档只记录团队当前用于本地开发和联调的环境信息，不属于产品需求文档。

## 2. 局域网 PostgreSQL

当前可用于开发测试的 PostgreSQL 信息如下：

- Host: `172.16.99.32`
- Port: `5432`
- Database: `mrchat`
- Username: `dev-ray`
- Password: `Mckj#2025ray`

说明：

- 该数据库仅限局域网访问
- 团队内联调时优先使用独立测试账号，不要直接改线上或公共环境数据

## 3. 推荐环境变量

推荐直接使用拆分字段，避免密码中的 `#` 在 URL 中带来转义问题：

```env
POSTGRES_ENABLED=true
POSTGRES_HOST=172.16.99.32
POSTGRES_PORT=5432
POSTGRES_USER=dev-ray
POSTGRES_PASSWORD=Mckj#2025ray
POSTGRES_DB=mrchat
POSTGRES_SSLMODE=disable
REDIS_ENABLED=false
```

如果希望使用单条 DSN，则需要对密码做 URL 编码：

```env
POSTGRES_DSN=postgres://dev-ray:Mckj%232025ray@172.16.99.32:5432/mrchat?sslmode=disable
REDIS_ENABLED=false
```

## 4. 本地启动建议

后端：

```bash
go run ./cmd/api
```

说明：

- 当前开发阶段默认 `POSTGRES_AUTO_MIGRATE=true`
- 服务启动时会自动执行 `db/migrations/` 下尚未应用的 goose SQL 迁移
- 这样仍然是在使用 goose，只是执行入口从 CLI 改成了 Go 服务启动流程
- 如果需要手工检查迁移状态，可直接运行 `go run ./cmd/migrate status` 或 `make migrate-status`

前端：

```bash
cd web
pnpm dev
```

前端默认请求：

- `http://127.0.0.1:8080/api/v1`

当前默认允许的前端开发源：

- `http://127.0.0.1:5173`
- `http://localhost:5173`

如果前端运行在其他端口或域名，需要同步设置：

```env
CORS_ALLOWED_ORIGINS=http://127.0.0.1:5173,http://localhost:5173,http://127.0.0.1:4173
```

如需改前端 API 地址，可设置：

```env
VITE_API_BASE_URL=http://127.0.0.1:8080
```

## 5. 当前联调范围

目前已适合做最小联调的页面和 API：

- `/login` <-> `POST /api/v1/auth/signin`
- `/signup` <-> `POST /api/v1/auth/signup`
- `/settings/profile` <-> `GET/PUT /api/v1/users/me`
- `/settings/security` <-> `GET /api/v1/users/me/security`、`PUT /api/v1/users/me/password`
- `/usage` <-> `GET /api/v1/users/me/quota`、`GET /api/v1/users/me/usage`、`GET /api/v1/billing/summary`、`GET /api/v1/billing/logs`
- `/chat`、`/chat/:conversationId` <-> `GET /api/v1/models`、`GET/POST/PUT/DELETE /api/v1/conversations`、`GET /api/v1/conversations/:id/messages`
- `/admin/upstreams` <-> `GET/POST/PUT /api/v1/admin/upstreams`
- `/admin/models` <-> `GET/POST/PUT /api/v1/admin/models`
- `/admin/users` <-> `GET /api/v1/admin/users`、`PUT /api/v1/admin/users/:id/quota`
- `/admin/audit-logs` <-> `GET /api/v1/admin/audit-logs`
