# MrChat v0.1 剩余交付计划

- 状态：执行计划
- 日期：2026-04-19
- 适用阶段：从“首轮联调可用”推进到“v0.1 可发布”
- 依据：
  - `docs/Requirements-Baseline-v0.1.md`
  - `docs/Implementation-Progress.md`
  - `docs/Development-Task-Breakdown-v0.1.md`
  - `docs/Admin-Console-Refactor-Plan-v0.1.md`
  - `docs/API-Contract-v0.1.md`
  - `docs/Data-Model-and-State-v0.1.md`
  - 当前工作区实际扫描结果（后端模块、前端路由、页面与 Makefile）

## 1. 当前判断

`codex/mrchat` 分支已经不是“工程初始化期”，也不是“只有设计文档”的阶段。当前状态更准确的描述是：

- 后端主链路已具备：认证、用户中心、管理台基础资源、聊天会话 CRUD、`/api/v1/chat/completions` 非流式与 SSE、Swagger、请求日志与基础限额统计
- 前端已具备：登录/注册、设置页、用量页、管理台骨架，以及一版可直接联调的 Chat 工作区
- 管理台已从“纯骨架”推进到“可配置上游 / 发现模型 / 导入模型 / 查看详情”的阶段
- 真正阻塞 v0.1 发布的，不再是项目脚手架，而是“核心业务闭环”与“发布前验证”

一句话结论：

- 当前最重要的工作不是继续扩展页面数量，而是把 `路由可靠性 + 账本结算 + 兑换码闭环 + Redis 降级 + 自动化验证` 补到可发布标准。

## 2. v0.1 发布门槛

| 门槛 | 定义 | 当前缺口 |
|---|---|---|
| Chat 可用性 | 模型可见、会话可用、SSE 稳定、上游失败可自动切换 | 多上游 failover / cooldown 仍未真正闭环 |
| 账本可信 | 请求前预扣、完成后结算、失败/取消退款，`quota_logs` 为真相 | 目前聊天返回中的 `billing` 仍为 `0`，结算尚未落地 |
| 运营可维护 | 管理员能安全维护上游、渠道、模型、兑换码和用户额度 | 兑换码链路仍未完成，停用/删除前检查不足 |
| 安全与降级 | 登录限流、风控日志、Redis 故障不阻断核心链路 | `AUTH-BE-03`、`INF-08` 仍是未完成项 |
| 发布可信度 | 至少有一轮稳定的 smoke / regression / CI 基线 | 当前仓库几乎没有测试文件，也没有 CI 工作流 |

## 3. 剩余 P0 工作流

### 3.1 工作流 A：核心聊天链路闭环

目标：

- 让聊天从“能跑通”升级为“可发布”

对应任务：

- `CHAT-BE-04`
- `CHAT-BE-07`
- `BILL-BE-01`
- `CHAT-BE-08`

重点文件：

- `internal/modules/chat/service.go`
- `internal/modules/chat/streaming.go`
- `internal/modules/chat/openai_client.go`
- `internal/modules/limits/service.go`
- `internal/modules/account/repository.go`
- `docs/API-Contract-v0.1.md`

本工作流需要补齐的关键点：

- 按 `route_bindings` 真正做优先级 failover，而不是单上游可用即结束
- 引入服务商失败计数与 cooldown 读取/回写逻辑
- 把 usage 回退口径真正接入结算，而不是只落 `llm_request_logs`
- 在非流式与流式链路里补 `pre_deduct / final_charge / refund`
- 统一 `response.completed`、请求日志和账本流水里的 billing 字段口径
- 收敛错误码、路由尝试元数据与排障日志

完成标准：

- 上游失败时会自动尝试下一个可用候选
- `response.completed.billing` 不再固定为 `0`
- 取消、失败、完成三类请求的余额与 `quota_logs` 一致
- 路由日志能回答“命中了哪个上游、失败了几次、为什么切换”

### 3.2 工作流 B：Redis 降级与认证安全

目标：

- 让 Redis 真正成为“可丢失、可降级”的运行时增强层，而不是隐形依赖

对应任务：

- `INF-08`
- `AUTH-BE-03`

重点文件：

- `internal/platform/cache/redis.go`
- `internal/modules/auth/service.go`
- `internal/http/middleware/auth.go`
- `internal/modules/admin/service.go`
- `internal/modules/chat/service.go`
- `docs/adr/ADR-0007-redis-runtime-strategy.md`

本工作流需要补齐的关键点：

- Redis key 命名空间、TTL 和用途边界固定下来
- 登录失败次数限制、基础风控日志、限流策略补齐
- Redis 不可用时，登录限流 / upstream cooldown 退化为单实例内存策略
- 核心链路不能因 Redis ping 失败而整体不可用

完成标准：

- 关闭 Redis 后，登录、聊天、结算、兑换、审计仍可继续工作
- 登录失败限制与风控日志可验证
- 文档中明确每类 key 的 TTL、前缀和降级策略

### 3.3 工作流 C：运营与充值闭环

目标：

- 让管理员可以完成“发额度、发兑换码、查记录、做安全停用”的基础运营动作

对应任务：

- `ADMIN-BE-08`
- `BILL-BE-02`
- `BILL-BE-03`
- `BILL-BE-05`
- `ADMIN-FE-06`

重点文件：

- `internal/modules/admin/handler.go`
- `internal/modules/admin/service.go`
- `internal/modules/billing/handler.go`
- `internal/modules/billing/service.go`
- `internal/modules/account/repository.go`
- `web/src/pages/admin/AdminRedeemCodesPage.vue`
- `web/src/api/admin.ts`

本工作流需要补齐的关键点：

- 兑换码批量生成、兑换、批次列表、兑换记录
- 兑换过程使用 PostgreSQL 事务保证原子性与幂等
- 兑换码只存 hash，不回写明文
- 上游 / 渠道 / 模型 / 用户组的停用优先、删除受限、删除前做引用检查
- 管理台页面从占位状态切到真正可操作

完成标准：

- 管理员能生成兑换码批次并查询记录
- 普通用户能兑换成功且余额、账本、审计一致
- 高风险资源不能被无引用检查地直接删除

### 3.4 工作流 D：用户侧完成度

目标：

- 把已具雏形的用户体验补到“能稳定演示和验收”

对应任务：

- `USAGE-FE-01`
- `USAGE-FE-02`
- `CHAT-FE-09`

重点文件：

- `web/src/pages/UsagePage.vue`
- `web/src/api/user.ts`
- `web/src/pages/ChatPage.vue`
- `web/src/layouts/AppLayout.vue`
- `web/src/api/chat.ts`

本工作流需要补齐的关键点：

- 用量页接真实摘要 / 趋势 / 账单过滤，而不是只显示基础 quota
- 兑换入口进入用户路径
- Chat 长列表、流式输出、stop / retry 的体验做一次针对性性能与状态校验

完成标准：

- `/usage` 不再只是账本列表壳子
- 用户能在站内完成“查看额度 -> 兑换 -> 继续聊天”
- 长对话场景下流式体验可接受

### 3.5 工作流 E：验证、CI 与发布准备

目标：

- 让仓库从“开发联调可用”升级到“可持续回归验证”

对应任务：

- `INF-05`
- `QA-01`
- `QA-02`
- `QA-03`
- `QA-04`
- `QA-05`
- `OBS-01`
- `OPS-01`
- `DOC-01`

重点文件：

- `Makefile`
- `docs/Local-Development.md`
- `docs/Implementation-Progress.md`
- `docs/API-Contract-v0.1.md`
- `.github/workflows/*`（当前缺失，需要补）

本工作流需要补齐的关键点：

- 为 Go / 前端建立最小 CI
- 把“注册 -> 登录 -> Chat -> 兑换 -> 继续聊天”跑成固定 smoke 用例
- 验证 failover、stop/refund、权限、安全、Redis 降级
- 补基础观测与部署文档

完成标准：

- 仓库内有最小 CI 工作流
- 每个 P0 验收路径都有明确命令或步骤
- 文档能够指导新环境部署和回归验证

## 4. 推荐执行顺序

### Wave 1：先补后端发布阻塞

- `CHAT-BE-04`
- `CHAT-BE-07`
- `BILL-BE-01`
- `INF-08`
- `AUTH-BE-03`

原因：

- 这些任务决定聊天是否可信、计费是否可信、Redis 是否会拖垮核心链路

### Wave 2：再补运营与充值闭环

- `CHAT-BE-08`
- `ADMIN-BE-08`
- `BILL-BE-02`
- `BILL-BE-03`
- `BILL-BE-05`
- `ADMIN-FE-06`

原因：

- Wave 1 结束后，平台才值得把运营动作接进来，否则发出去的额度和请求消耗无法对账

### Wave 3：补用户路径与体验收口

- `USAGE-FE-01`
- `USAGE-FE-02`
- `CHAT-FE-09`

原因：

- 此时后端契约基本稳定，前端不会反复返工

### Wave 4：最后做发布前验证

- `INF-05`
- `QA-01` ~ `QA-05`
- `OBS-01`
- `OPS-01`
- `DOC-01`

原因：

- 这批任务要基于前面三波的稳定结果，否则 CI 和验收脚本会持续重写

## 5. 当前最该创建的 12 个 issue

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

## 6. 近期 Sprint 切法

### Sprint A：把聊天和账本补成“可信链路”

- `CHAT-BE-04`
- `CHAT-BE-07`
- `BILL-BE-01`
- `INF-08`
- `AUTH-BE-03`

目标：

- 上游故障能切换
- `billing` 字段不再为零
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
- 达到 v0.1 首发验收标准

## 7. 执行原则

- 不再把 `INF-01` / `DB-01` 这类已完成的启动任务放在当前优先级前列
- 每一轮都优先补“会影响真实运营和真实计费”的路径
- 任何涉及 `quota_logs`、兑换码、结算、余额的逻辑，都以 PostgreSQL 事务与显式数据写入为准
- Redis、前端样式、管理台展示都应服务于“发布阻塞项”清理，而不是反过来牵引主线
