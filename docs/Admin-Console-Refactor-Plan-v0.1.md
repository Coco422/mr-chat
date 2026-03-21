# MrChat 管理控制台重构计划 v0.1

- 状态：工作计划
- 日期：2026-03-21
- 适用阶段：管理台二次收敛 -> 接口补齐 -> 前后端联调
- 依赖文档：
  - `docs/Requirements-Baseline-v0.1.md`
  - `docs/API-Contract-v0.1.md`
  - `docs/Data-Model-and-State-v0.1.md`
  - `docs/Development-Task-Breakdown-v0.1.md`

## 1. 背景

当前管理控制台已经具备首轮“能打通链路”的后端和页面骨架，但在实际查看前端页面后，已经暴露出一批影响可用性和理解成本的问题：

- 模型管理仍然是手工录入，不符合“先从上游发现，再导入平台模型目录”的运营习惯
- 模型页和部分管理页直接展示 UUID，不适合运营人员阅读
- `route_bindings`、`channel` 等概念过于工程化，UI 语义不够直观
- 上游、渠道、模型、用户组页面主要只有“列表 + 新增”，编辑、详情、停用/删除策略不完整
- 兑换码页目前仍是占位，不属于本轮管理台收敛的直接范围

这份计划的目标，是把管理台从“开发者可用”收敛到“运营可读、前后端可继续开发”的状态。

## 2. 这轮计划的核心结论

### 2.1 模型仍然要落本地，但来源改成“导入”

- `models` 表继续保留
- 平台模型不是上游原始模型名的镜像，而是平台面向用户展示的逻辑模型目录
- 平台模型仍然需要承载：
  - 展示名
  - 用户组可见性
  - 路由规则
  - 能力标记
  - 定价与后续计费扩展字段
- 但模型的创建方式应从“纯手工录入”改成“先从某个 upstream 发现候选模型，再选择导入”

一句话：

- v0.1 后续模型管理 = `上游模型发现 + 选择导入 + 本地补充业务配置`

### 2.2 `route_binding` 保留数据结构，但前台文案要改成人类可读

`route_binding` 的本意不是给运营人员直接理解数据库结构，而是表达：

- 某个逻辑模型在实际请求时，应该尝试哪些上游
- 每个上游处于什么渠道下
- 尝试顺序是什么

对运营侧更容易理解的表达应该是：

- “默认路由”
- “候选上游”
- “优先级”
- “主路由 / 备用路由”

也就是说：

- 数据结构上仍保留 `route_bindings`
- 管理台显示层统一改称“路由规则”或“候选上游”
- 列表页不再直接展示 `channel_id -> upstream_id`

### 2.3 `channel` 保留，但在 v0.1 中应降级为“高级配置”

`upstream` 和 `channel` 的职责区分如下：

- `upstream`
  - 真实外部服务地址
  - 负责 `base_url`、鉴权、超时、冷却、健康状态
- `channel`
  - 平台内部的业务/计费/路由维度
  - 用于区分不同运营通道、定价口径、报表维度

现状问题不是这个概念不合理，而是：

- 当前聊天主链路还没有把 `channel` 做成用户或运营明显感知到的能力
- 所以前端看到“渠道管理”会觉得抽象且不知用途

本轮建议：

- `channel` 继续保留
- 在文档和 UI 上明确其定义为“高级配置”
- 当系统只有一个默认 channel 时，可以在管理台上弱化其存在感，不要求运营频繁进入该页

### 2.4 管理台默认不直接展示 UUID

后续规则统一为：

- 除审计日志外，管理台主界面应优先展示 human-readable 信息
- 资源关联返回值应带嵌套对象：
  - `id`
  - `name` / `display_name`
  - 必要状态字段
- UUID 可保留在：
  - 调试抽屉
  - 复制按钮
  - Swagger
  - 审计日志

## 3. 页面级改造计划

### 3.1 上游配置页

本页目标改成“管理真实外部服务连接”。

应支持：

- 列表
- 创建
- 详情
- 编辑
- 停用
- 触发模型发现

建议展示字段：

- 名称
- Provider Type
- Base URL
- 状态
- 超时
- 冷却时间
- 失败阈值
- 最近一次模型发现时间

本页不建议默认直接显示 API Key 明文。

### 3.2 渠道管理页

本页目标改成“管理平台内部计费/运营通道”。

应支持：

- 列表
- 创建
- 详情
- 编辑
- 停用

建议展示字段：

- 名称
- 描述
- 状态
- 计费口径摘要
- 已关联模型数

如果系统只有一个默认 channel，可以在导航文案上提示其为“高级配置”。

### 3.3 模型管理页

这是本轮需要重点重构的页面。

目标流程应改成：

1. 选择一个 upstream
2. 拉取该 upstream 的候选模型列表
3. 勾选需要启用的模型导入到本地目录
4. 为导入后的模型补充：
   - 展示名
   - 用户组可见性
   - 路由规则
   - 状态
   - 后续定价或能力配置

页面结构建议拆成两个子区域：

- 模型发现 / 导入
- 本地模型目录管理

模型目录列表中，后续显示应改成：

- 模型名：`display_name`
- 模型标识：`model_key`
- 状态
- 可见用户组：用户组名称列表，而不是 UUID
- 路由规则：如 `默认路由 -> OpenAI Relay A (优先级 1)`

路由规则编辑器建议支持：

- 多行 route binding
- 选择 `channel`
- 选择 `upstream`
- 填写 `priority`
- 启停单条规则

### 3.4 用户组页

当前方向基本正确，但后续应补：

- 分组详情
- 分组编辑
- 限额规则中的模型名展示
- 被哪些模型引用的摘要

### 3.5 用户管理页

当前方向基本正确，但后续应补：

- 更清晰的人类可读字段
- 详情信息入口
- 用户组、限额、调额、调整记录的分区展示

### 3.6 兑换码页

本页仍归属 `M5`，不纳入本轮管理台结构重构主线。

本轮只记录结论：

- 当前页面确实还未实现
- 后续仍按“批量生成 / 批次列表 / 兑换记录”推进

## 4. 后端接口改造计划

### 4.1 保留的现有接口

以下接口继续保留：

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

### 4.2 需要新增的接口

建议新增：

- `GET /api/v1/admin/upstreams/:id`
  - 单个上游详情
- `GET /api/v1/admin/upstreams/:id/discovered-models`
  - 请求指定 upstream 的模型发现接口，并返回标准化候选模型列表
- `GET /api/v1/admin/channels/:id`
  - 单个 channel 详情
- `GET /api/v1/admin/models/:id`
  - 单个模型详情，包含 hydrated route bindings
- `POST /api/v1/admin/models/import`
  - 根据“已发现模型”批量导入本地模型目录
- `GET /api/v1/admin/user-groups/:id`
  - 单个用户组详情

可选新增：

- `GET /api/v1/admin/references`
  - 返回前端常用选项字典，如 `upstreams / channels / user_groups / models`
  - 用于减少管理台多页面重复拉列表

### 4.3 返回结构的人类可读规则

后续管理接口应遵循：

- 写接口仍接受 ID
- 读接口尽量返回 hydrated 对象

例如模型列表返回结构建议改为：

```json
{
  "id": "uuid",
  "model_key": "Qwen/Qwen3.5-122B-A10B",
  "display_name": "Qwen 3.5 122B",
  "status": "active",
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
      }
    }
  ]
}
```

注意：

- `visible_user_group_ids` 可继续保留，方便前端写回
- 但页面主展示应优先使用 `visible_user_groups`
- `route_bindings` 内也应优先给出 `channel.name` 和 `upstream.name`

### 4.4 删除策略

这里不建议马上把所有资源都做成硬删除。

原因：

- 当前 `upstreams / channels / models / user_groups` 并没有统一的 `deleted_at`
- 这些资源已经被会话、消息、请求日志、限额规则和路由关系引用

本轮推荐策略：

- 第一阶段优先补 `详情 + 编辑 + 停用`
- 第二阶段再补“安全删除”，删除前必须做引用检查
- 若后续确认需要真正删除，再单独补 schema 变更与删除语义

一句话：

- 先把“能安全改、能停用”做好，再决定“是否允许删”

## 5. 数据与显示语义修正

### 5.1 `visible_user_group_ids`

这是写路径字段，不适合作为 UI 直接展示字段。

后续改成：

- 写入时继续传 ID 数组
- 列表展示时返回并显示 `visible_user_groups[].name`

### 5.2 `route_bindings`

这是内部结构字段，不适合作为用户第一眼看到的文案。

后续 UI 建议统一改叫：

- “路由规则”
- “候选上游”
- “备用上游”

### 5.3 `channel`

这是平台业务维度，不是用户分组，不是上游别名。

后续页面和文档里建议统一解释为：

- “计费 / 运营通道”

## 6. 分阶段实施建议

### Phase A：先补人类可读返回

目标：

- 不改数据库结构，先把管理接口返回变得适合前端展示

包含：

- 模型列表返回 `visible_user_groups`
- 模型列表返回 hydrated `route_bindings`
- 其他列表接口补充必要 `name` 字段

### Phase B：补模型发现与导入

目标：

- 管理员不再需要手动输入模型名来创建模型

包含：

- `GET /admin/upstreams/:id/discovered-models`
- `POST /admin/models/import`

### Phase C：补详情与编辑

目标：

- 管理台从“新增 + 列表”变成“可维护资源”

包含：

- 单资源详情接口
- 前端编辑表单
- 路由规则编辑器

### Phase D：补停用与安全删除

目标：

- 完成资源生命周期管理

包含：

- 停用
- 删除前引用检查
- 必要时补 schema 以支持 archive / soft delete

### Phase E：推进兑换码页

目标：

- 单独完成 `M5` 里的兑换码闭环

## 7. 任务拆解建议

建议把这轮新增任务补成以下 issue：

- `ADMIN-BE-05`
  - 实现上游模型发现接口与标准化结果返回
- `ADMIN-BE-06`
  - 实现模型导入接口与本地模型初始化逻辑
- `ADMIN-BE-07`
  - 改造 admin 列表/详情接口，返回 human-readable 关联对象
- `ADMIN-BE-08`
  - 补单资源详情接口与安全停用/删除策略
- `ADMIN-FE-07`
  - 实现模型发现、勾选导入和路由规则编辑
- `ADMIN-FE-08`
  - 实现上游/渠道/模型/用户组详情与编辑
- `ADMIN-FE-09`
  - 清理管理台中的 UUID 直出，统一用人类可读字段展示

## 8. 当前结论

这轮不是要推翻现有后端结构，而是把现有“能跑”的骨架收成一套更适合运营使用的管理模型。

最关键的三点是：

- 模型管理从“手工录入”升级为“上游发现 + 导入”
- 管理台默认不再展示 UUID，而是展示名称和状态
- `upstream / channel / model / route_binding` 的语义要在接口和 UI 上同时变得可理解
