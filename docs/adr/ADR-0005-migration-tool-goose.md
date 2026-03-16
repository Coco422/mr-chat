# ADR-0005: 数据库迁移工具选型为 goose

- 状态：通过
- 日期：2026-03-17
- 相关文档：
  - `docs/Requirements-Baseline-v0.1.md`
  - `docs/Development-Task-Breakdown-v0.1.md`
  - `docs/adr/ADR-0004-database-postgresql.md`
  - `Plan.md`

## 背景与问题

MrChat 已经收敛到 `Gin + GORM + PostgreSQL`，但数据库迁移工具此前仍未明确。迁移工具如果长期悬空，会直接影响本地初始化、CI 校验、发布流程和 SQL 文件组织方式。

当前约束：

- 团队愿意引入 Go 生态内的工具，并接受为此建立新的迁移约定
- v0.1 以 SQL migration 为主，不需要大量用 Go 编写数据迁移逻辑
- 需要把迁移执行从应用启动流程中拆开，避免多实例启动时并发跑迁移
- 迁移工具最好和 Go 开发流程贴近，降低后续维护割裂感

## 选项

### 选项 A：goose

- 优点：
  - Go 生态常见，CLI 与库两种使用方式都成熟
  - SQL migration 工作流直接，支持 `status`、`validate`、`up`、`down`
  - 同时支持 SQL migration 与 Go migration，后续需要数据迁移时有余地
  - 支持嵌入式 migration，但不强迫项目一开始就内嵌
- 缺点：
  - SQL 文件格式采用 `-- +goose Up/Down` 注释分段，和 Flyway 的命名习惯不同
  - 功能面比极简迁移执行器更宽，需要团队约束只用需要的子集
- 风险与未知：
  - 需要尽早统一命名、是否使用顺序号、是否允许 Go migration，否则迁移风格会漂移
- 迁移/回滚成本：
  - 低，当前项目尚未开始真实迁移文件沉淀

### 选项 B：golang-migrate

- 优点：
  - Go 生态成熟，适合纯迁移工具场景
  - up/down 双文件模型清晰，工具边界简单
- 缺点：
  - CLI 工作流更偏底层执行器，不如 goose 接近日常开发体验
  - 对当前项目没有比 goose 更明显的组织优势
- 风险与未知：
  - 团队需要同时适应新的文件命名与更显式的执行方式
- 迁移/回滚成本：
  - 中

### 选项 C：Flyway

- 优点：
  - 团队在其他语言服务里已有经验
  - SQL-first，状态管理和迁移治理都成熟
- 缺点：
  - 不属于 Go 主流工具链，日常开发体验会和项目技术栈割裂
  - 对当前项目引入收益不足以覆盖“另带一套工具心智”的成本
- 风险与未知：
  - 仍需准备单独运行环境和额外文档约定
- 迁移/回滚成本：
  - 中

## 决策

选择 **goose** 作为 MrChat 的数据库迁移工具。

理由：

- 它更贴近 Go 项目的日常开发与运维方式，能减少技术栈割裂
- 相比 `golang-migrate`，它更接近日常团队工作流，学习成本更低
- 当前阶段以 SQL migration 为主，但 goose 也给后续 Go migration 留了空间

## 影响

- 数据库结构变更统一通过 goose migration 管理
- Gin 服务启动流程不负责自动跑迁移；迁移由本地脚本、CI/CD 或独立部署步骤执行
- GORM 不使用 `AutoMigrate` 管生产 schema；schema 变更统一走 goose
- 需要尽快定义目录、命名规范、`status` / `validate` / `up` / `down` 的执行约定

## 后续行动

1. 在任务清单中新增 goose 基建任务，并让首批表迁移显式依赖它。
2. 约定迁移目录和命名格式，例如 `db/migrations/00001_init_users.sql` 或时间戳方案。
3. 补本地与 CI 的迁移命令，至少覆盖 `status`、`validate` 和 `up`。
4. 在部署文档里明确“迁移先于应用发布执行”。
