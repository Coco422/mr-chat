# 专题调研：流式输出（SSE / WebSocket）与对话体验

## 1. 我们的需求（对话站的“流式”到底要解决什么）

- 首 token 快：用户发出后尽快看到反馈
- 过程可控：可停止、可重试、可继续、可复盘
- 信息结构化：支持正文 + reasoning + tool calls + 引用/来源（未来）
- 稳定：代理层/反向代理不应该把流式缓存住或断流后 UI 崩掉

## 2. 协议选择：SSE 优先，WebSocket 作为补充

### 2.1 SSE（HTTP streaming）

适合：

- ChatCompletions/Responses 的 token 流式
- 兼容 OpenAI 生态（`data: ...` + `[DONE]`）
- 断线重连和中间件兼容性更好（比 WS 简单）

风险：

- 代理缓冲（Nginx/Cloudflare 等）需要明确禁用 buffering
- chunk 频率高时，服务端 flush 与前端渲染都要节流

### 2.2 WebSocket

适合：

- 多路复用：一个连接跑多条会话/推送（在线状态、协作、通知）
- OpenAI Realtime / 语音通话类低延迟交互

建议：先把“聊天流式”做成 SSE，实时协作/语音再上 WS。

## 3. 事件格式建议（以 OpenAI 兼容为主）

- 每个事件：`data: {json}\n\n`
- 结束标记：`data: [DONE]\n\n`
- delta 字段：
  - `delta.content` 正文
  - `delta.reasoning_content` / `delta.reasoning` 推理（不同上游可能不同）
  - tool calls（未来）：建议沿用 OpenAI 的 tool_calls delta 结构

## 4. 前端实现要点（决定体验与性能）

- “只更新最后一条 assistant”：流式阶段不要重建整个 messages 数组
- 节流：
  - 正文 append 可以按帧（rAF）或 33ms flush
  - debug/trace 更新更慢（200-500ms）
- 停止生成：
  - 关闭 SSE 连接（或 AbortController）
  - 把最后一条 incomplete 内容做收尾清理（尤其是 `<think>` 未闭合）
- 错误收敛：
  - JSON chunk parse 失败不应让整个 UI 崩溃
  - 正常关闭不要误报 error（readyState/status 判定）

参考实现：new-api 的 SSE hook 在 `new-api/web/src/hooks/playground/useApiRequest.jsx`。

## 5. 后端实现要点（决定稳定性与可观测）

- 上游请求要可取消：客户端 stop 时，`context.Cancel` 需要打断上游连接
- flush 策略：按 chunk flush，但要避免每个字都 flush（会拖垮 CPU）
- 计费/结算：
  - 预扣 + 最终结算（Plan.md/Architecture-Design-OUI-Integration.md 已提到）
  - 断流时的“部分结算”策略要明确（按已收到 token、或按上游 usage）
- 落库策略：
  - 流式过程中不要频繁写 DB（建议消息完成后一次写入）
  - 如需“中途可恢复”，用内存/Redis 写入增量，完成后 compaction

## 6. 对“更好的对话”的直接支撑点

- reasoning 与正文分离展示，默认折叠（减少噪音与渲染压力）
- `<think>` 标签兼容（很多模型会输出）
- 流式阶段只渲染必要内容，完成后再做重型 Markdown 渲染升级

