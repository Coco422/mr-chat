# ADR-0002: Chat UI 策略选择自研（MVP 先上用户中心 + Chat）

- 状态：通过
- 日期：2026-03-01
- 相关文档：
  - `docs/Current-Design-Summary.md`
  - `docs/research/compare.md`
  - `docs/research/topics/frontend-performance.md`
  - `docs/research/topics/streaming.md`

## 背景与问题

MrChat 当前阶段目标是“快速上线一个可用的聚合对话站”，现阶段需求不复杂，优先级是：

- 用户中心（注册/登录、个人资料、配额/用量、API Key 等）
- Chat 模块（会话/消息、流式输出、基础 Markdown 渲染）

此前存在一种路线：通过 iframe/外部应用方式集成 OpenWebUI/HiveChat 等对话 UI（Portal 路线）。需要决定是否走该路线，还是直接自研一个 simple chat UI 先上线。

## 选项

### 选项 A：Portal/嵌入外部对话 UI（iframe/反代）

- 优点：
  - UI 上线更快（短期）
  - 能复用成熟产品的大量功能
- 缺点：
  - 体验不可控：导航、主题、快捷键、账户体系与我们难统一
  - 性能不可控：流式渲染/长对话性能问题可能无法按我们的节奏修
  - 集成复杂：鉴权/key 注入、跨域、埋点、回调与权限边界会反复踩坑
  - 许可证风险：例如 open-webui v0.6.6+ 的 Branding 限制，new-api 为 AGPLv3（详见 `docs/research/topics/licensing.md`）

### 选项 B：自研 simple Chat UI（Vue 3 + Vite）

- 优点：
  - 体验与性能完全可控，能围绕“更好的对话”做结构化呈现（reasoning/think/引用等）
  - 与用户中心/配额/计费/审计的集成最简单
  - 许可证风险最小（仅借鉴思路与协议兼容）
- 缺点：
  - 需要自己实现基础聊天 UI（会话列表、消息渲染、流式、停止/重试等）
  - 初期功能不如成熟 UI 丰富

## 决策

选择 **选项 B：自研 simple Chat UI** 作为 MrChat 的核心对话主路径。

补充说明（2026-03-12）：

- 本 ADR 约束的是“核心 Chat 主路径”，不是“平台上完全禁止外部服务入口”
- MrChat 仍可提供受控的外部子服务入口，用于体验或承载独立部署能力
- 外部子服务入口可采用 iframe 嵌入或新窗口跳转，但不替代主 Chat 的默认体验

## 决策范围（MVP 明确做什么/不做什么）

### MVP 要做

- 用户中心：
  - 注册/登录（基础）
  - 个人资料
  - 配额/用量查看（至少展示本月/总量）
  - API Key 管理（如果产品形态需要对外提供 OpenAI 兼容 API）
- Chat：
  - 会话 CRUD（新建、重命名、删除、列表）
  - 消息发送与流式输出（SSE）
  - 停止生成、失败重试
  - 基础 Markdown 渲染（先做轻量能力，重型渲染按需）

### MVP 明确不做（先延后）

- RAG、插件系统、复杂 Agent
- 支付/充值（可用“手动加余额/额度”替代）
- 不把多应用 Portal/iframe 集成作为主对话方案
- 复杂协作（多用户同会话编辑、在线状态等）

## 影响

- 前端将围绕“流式渲染 + 长对话性能”做设计，不受外部 UI 限制
- 后端 API 需要优先提供：
  - 会话/消息接口
  - SSE 流式接口（可取消、可结算）
  - 用量与配额接口
- 若提供外部子服务入口，需要补充：
  - 服务入口配置
  - iframe / 跳转模式控制
  - 用户组可见性与访问日志
- 文档对比与路线图需要同步，避免继续讨论 Portal 方案作为默认路线

## 后续行动

1. 更新文档：在 `docs/Current-Design-Summary.md` 与 `docs/research/compare.md` 标注已选择“自研 Chat UI”。
2. 定义 MVP 页面与路由（建议：`/login`、`/chat`、`/settings`、`/admin` 可后置）。
3. 为 Chat UI 建一个性能红线：
   - 流式阶段只更新最后一条 assistant
   - 节流 flush（按帧/固定间隔）
   - 长列表虚拟化（达到阈值立刻上）
