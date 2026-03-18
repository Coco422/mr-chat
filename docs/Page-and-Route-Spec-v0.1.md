# MrChat v0.1 页面与路由规格

- 状态：实现设计草案
- 日期：2026-03-18
- 依赖基线：`docs/Requirements-Baseline-v0.1.md`

## 1. 目标

这份文档用于把 v0.1 需求基线拆成前端可执行的信息架构、页面清单与路由规则。

设计原则：

- 优先保证主路径清晰：登录 -> Chat -> 查看用量 -> 充值 -> 继续聊天
- 主聊天体验优先，避免把设置、调试、运营入口塞进同一页面
- URL 要能表达页面状态，尤其是对话上下文
- 外部子服务入口允许存在，但不干扰主 Chat 导航

## 2. 信息架构

### 2.1 用户侧一级导航

- `Chat`
- `Usage`
- `Settings`
- `Services`（P1，可按开关隐藏）

### 2.2 管理侧一级导航

- `Upstreams`
- `Channels`
- `Models`
- `User Groups`
- `Users`
- `Redeem Codes`
- `Audit Logs`
- `Service Entries`（P1）

### 2.3 Shell 结构

#### AuthShell

用于未登录页面。

- 品牌区
- 登录/注册表单区
- 轻量说明与错误提示

#### AppShell

用于普通用户登录后页面。

- 左侧主导航
- 顶部用户菜单
- 主内容区
- 全局 Toast / Modal / Loading 容器

#### AdminShell

用于管理后台。

- 左侧管理导航
- 顶部环境与当前管理员信息
- 主内容区
- 审计操作确认弹窗

## 3. 路由清单

| 路由 | 访问角色 | 优先级 | 用途 |
|---|---|---|---|
| `/login` | Guest | P0 | 登录页 |
| `/signup` | Guest | P0 | 注册页 |
| `/` | All | P0 | 根据登录态跳转到默认页 |
| `/chat` | User/Admin/Root | P0 | 空态聊天页或最近对话入口 |
| `/chat/:conversationId` | User/Admin/Root | P0 | 指定会话聊天页 |
| `/usage` | User/Admin/Root | P0 | 额度、用量、充值入口 |
| `/settings/profile` | User/Admin/Root | P0 | 个人资料设置 |
| `/settings/security` | User/Admin/Root | P0 | 密码与登录安全设置 |
| `/settings/invites` | User/Admin/Root | P1 | 邀请码与邀请记录 |
| `/services` | User/Admin/Root | P1 | 外部子服务入口列表 |
| `/services/:serviceEntryId` | User/Admin/Root | P1 | iframe 容器页或跳转过渡页 |
| `/admin` | Admin/Root | P0 | 管理后台默认入口，跳转到 `/admin/upstreams` |
| `/admin/upstreams` | Admin/Root | P0 | 上游管理 |
| `/admin/channels` | Admin/Root | P0 | 渠道与计费通道管理 |
| `/admin/models` | Admin/Root | P0 | 模型管理 |
| `/admin/user-groups` | Admin/Root | P0 | 用户分组与分组限额管理 |
| `/admin/users` | Admin/Root | P0 | 用户与额度管理 |
| `/admin/redeem-codes` | Admin/Root | P0 | 兑换码管理 |
| `/admin/audit-logs` | Admin/Root | P0 | 审计日志 |
| `/admin/service-entries` | Admin/Root | P1 | 外部子服务入口管理 |
| `/403` | All | P0 | 无权限页 |
| `/404` | All | P0 | 未找到页 |

## 4. 路由规则

### 4.1 默认跳转

- 未登录访问受保护页面时，跳转到 `/login`
- 已登录访问 `/login` 或 `/signup` 时，跳转到 `/chat`
- 访问 `/` 时：
  - 未登录 -> `/login`
  - 已登录普通用户 -> `/chat`
  - 已登录管理员仍默认进入 `/chat`，不直接跳后台

### 4.2 权限守卫

- `User` 不能访问任何 `/admin/*`
- `Admin` 与 `Root` 可以访问用户侧与管理侧全部页面
- `Services` 页面是否展示，取决于是否启用服务入口功能与用户是否有可见服务

### 4.3 对话 URL 规则

- `/chat` 表示“未选中对话”或“准备创建新对话”
- `/chat/:conversationId` 表示已选中会话
- 创建新对话后，前端应立即跳转到对应 `conversationId`
- 重命名、发送消息、停止生成都不改变当前路由

### 4.4 外部服务入口规则

- `launch_mode = iframe` 时，进入 `/services/:serviceEntryId`
- `launch_mode = new_tab` 时，服务卡片点击后直接新窗口打开，当前页保留在 `/services`
- 若某服务配置为 iframe 但浏览器或服务端拒绝嵌入，前端应降级为新窗口打开

## 5. 页面规格

## 5.1 `/login`

目标：

- 让用户尽快完成登录

核心元素：

- 用户名/邮箱输入框
- 密码输入框
- 登录按钮
- 去注册入口
- 登录失败提示

依赖接口：

- `POST /api/v1/auth/signin`

成功后动作：

- 拉取当前用户信息
- 跳转到 `/chat`

## 5.2 `/signup`

目标：

- 创建新账号并进入系统

核心元素：

- 用户名
- 邮箱
- 密码
- 确认密码
- 邀请码输入框（P1，可隐藏）

依赖接口：

- `POST /api/v1/auth/signup`

成功后动作：

- 后端直接返回 `access_token` 并设置 `refresh_token` Cookie
- 前端拉取当前用户后跳转到 `/chat`

## 5.3 `/chat`

目标：

- 作为主工作区承载全部核心聊天行为

页面结构：

- 左栏：会话列表
- 主区头部：当前模型、会话标题、状态反馈
- 主区消息流：消息列表、加载更多、错误态
- 底部输入区：输入框、发送、停止生成、重试

关键交互：

- 新建会话
- 切换会话
- 发送消息
- 流式渲染最后一条 assistant 消息
- 停止生成
- 失败重试

依赖接口：

- `GET /api/v1/conversations`
- `POST /api/v1/conversations`
- `PUT /api/v1/conversations/:id`
- `DELETE /api/v1/conversations/:id`
- `GET /api/v1/conversations/:id/messages`
- `POST /api/v1/chat/completions`
- `GET /api/v1/models`

实现约束：

- 会话切换不应导致整页重建
- 流式阶段只更新最后一条 assistant 消息
- `conversationId` 作为 URL 状态源

## 5.4 `/usage`

目标：

- 给用户一个统一的额度、用量、充值入口

页面结构：

- 顶部摘要卡片：当前余额、近 7 天消耗、本月消耗
- 趋势图或简表：按天统计
- 流水表：最近消费、退款、兑换、管理员调额
- 兑换码表单

依赖接口：

- `GET /api/v1/users/me/quota`
- `GET /api/v1/users/me/usage`
- `GET /api/v1/billing/logs`
- `GET /api/v1/billing/summary`
- `POST /api/v1/billing/redeem`（M5，可按功能开关隐藏）

## 5.5 `/settings/profile`

目标：

- 管理个人资料与偏好

核心元素：

- `display_name`
- 头像 URL 或上传占位
- 时区/语言偏好
- 保存按钮

依赖接口：

- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`

## 5.6 `/settings/security`

目标：

- 管理密码与会话安全

核心元素：

- 当前密码
- 新密码
- 确认新密码
- 最近登录信息（可选）

依赖接口：

- `GET /api/v1/users/me/security`
- `PUT /api/v1/users/me/password`

## 5.7 `/services`（P1）

目标：

- 展示平台可访问的外部子服务入口

页面结构：

- 服务卡片列表
- 服务说明
- 标签：体验服务 / 独立子服务 / 需要新窗口

关键交互：

- 进入 iframe 服务
- 新窗口打开服务

依赖接口：

- `GET /api/v1/service-entries`

## 5.8 `/services/:serviceEntryId`（P1）

目标：

- 承载 iframe 服务的容器页，或作为外跳前的中转页

页面结构：

- 顶部栏：返回、服务名称、在新窗口打开按钮
- 内容区：iframe 或错误提示

异常处理：

- 服务不存在 -> 404
- 用户无权访问 -> 403
- iframe 被拒绝 -> 显示降级提示并允许新窗口打开

## 5.9 `/admin/upstreams`

目标：

- 管理上游地址、密钥、状态与备注

表格字段建议：

- 名称
- Provider 类型
- Base URL
- 状态
- 最近健康情况
- 更新时间

关键交互：

- 新建/编辑上游
- 启用/停用上游

依赖接口：

- `GET /api/v1/admin/upstreams`
- `POST /api/v1/admin/upstreams`
- `PUT /api/v1/admin/upstreams/:id`

## 5.10 `/admin/models`

目标：

- 管理模型与路由绑定

表格字段建议：

- 模型标识
- 显示名称
- 定价倍率
- 上下文长度
- 可见用户组
- 状态

关键交互：

- 新建/编辑模型
- 设置默认或按渠道的上游优先级
- 配置 `visible_user_group_ids`

依赖接口：

- `GET /api/v1/admin/models`
- `POST /api/v1/admin/models`
- `PUT /api/v1/admin/models/:id`

## 5.11 `/admin/channels`

目标：

- 管理模型调用渠道、计费口径和路由归属

表格字段建议：

- 渠道名称
- 状态
- Description
- Billing Config

关键交互：

- 新建/编辑渠道
- 查看渠道是否被模型路由使用

依赖接口：

- `GET /api/v1/admin/channels`
- `POST /api/v1/admin/channels`
- `PUT /api/v1/admin/channels/:id`

## 5.12 `/admin/user-groups`

目标：

- 管理用户运营分组和该分组的模型限额模板

表格字段建议：

- 分组名
- 状态
- 描述
- 默认限额模板
- 模型覆盖规则数

关键交互：

- 新建/编辑用户分组
- 配置分组默认限额
- 配置某模型覆盖规则

依赖接口：

- `GET /api/v1/admin/user-groups`
- `POST /api/v1/admin/user-groups`
- `PUT /api/v1/admin/user-groups/:id`
- `GET /api/v1/admin/user-groups/:id/limits`
- `PUT /api/v1/admin/user-groups/:id/limits`

## 5.13 `/admin/users`

目标：

- 查用户、看额度、做人工调额与限额调整

表格字段建议：

- 用户名
- 邮箱
- 角色
- 状态
- 用户分组
- 当前额度
- 累计消耗

关键交互：

- 搜索用户
- 调整用户分组
- 调整额度
- 查询某模型下的 hour/week/lifetime 限额使用情况
- 新增单用户 direct adjustment
- 启用/禁用用户

依赖接口：

- `GET /api/v1/admin/users`
- `PUT /api/v1/admin/users/:id/group`
- `PUT /api/v1/admin/users/:id/quota`
- `GET /api/v1/admin/users/:id/limit-usage`
- `GET /api/v1/admin/users/:id/limit-adjustments`
- `POST /api/v1/admin/users/:id/limit-adjustments`

## 5.14 `/admin/redeem-codes`

目标：

- 生成和管理兑换码批次

页面结构：

- 批量生成表单
- 批次列表
- 兑换记录表

依赖接口：

- `POST /api/v1/admin/redeem-codes/batch`
- `GET /api/v1/admin/redeem-codes`

## 5.15 `/admin/audit-logs`

目标：

- 查看关键操作审计日志

表格字段建议：

- 时间
- 操作人
- 动作
- 资源类型
- 资源 ID
- 结果
- 请求 ID

依赖接口：

- `GET /api/v1/admin/audit-logs`

## 5.16 `/admin/service-entries`（P1）

目标：

- 管理外部子服务入口配置

关键字段：

- 名称
- `slug`
- 地址
- 展示方式
- 状态
- 可见组

依赖接口：

- `GET /api/v1/admin/service-entries`
- `POST /api/v1/admin/service-entries`
- `PUT /api/v1/admin/service-entries/:id`

## 6. 页面与接口映射建议

| 页面 | 首屏必拉接口 |
|---|---|
| `/chat` | `GET /users/me`、`GET /models`、`GET /conversations` |
| `/chat/:conversationId` | `GET /users/me`、`GET /models`、`GET /conversations`、`GET /conversations/:id/messages` |
| `/usage` | `GET /users/me/quota`、`GET /users/me/usage`、`GET /billing/summary`、`GET /billing/logs` |
| `/settings/profile` | `GET /users/me` |
| `/settings/security` | `GET /users/me/security` |
| `/services` | `GET /service-entries` |
| `/admin/upstreams` | `GET /admin/upstreams` |
| `/admin/models` | `GET /admin/models` |
| `/admin/users` | `GET /admin/users` |
| `/admin/redeem-codes` | `GET /admin/redeem-codes` |
| `/admin/audit-logs` | `GET /admin/audit-logs` |

## 7. 前端实现优先级建议

### 7.1 第一批页面

- `/login`
- `/signup`
- `/chat`
- `/chat/:conversationId`
- `/usage`
- `/settings/profile`
- `/admin/upstreams`
- `/admin/models`
- `/admin/users`
- `/admin/redeem-codes`
- `/admin/audit-logs`

### 7.2 第二批页面

- `/settings/security`
- `/settings/invites`
- `/services`
- `/services/:serviceEntryId`
- `/admin/service-entries`
