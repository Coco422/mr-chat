# ADR-0006: 默认 ORM 选型为 GORM

- 状态：通过
- 日期：2026-03-17
- 相关文档：
  - `docs/Requirements-Baseline-v0.1.md`
  - `docs/Data-Model-and-State-v0.1.md`
  - `docs/adr/ADR-0003-backend-framework-gin.md`
  - `docs/adr/ADR-0004-database-postgresql.md`
  - `docs/adr/ADR-0005-migration-tool-goose.md`

## 背景与问题

MrChat 已经收敛到 Go 后端、PostgreSQL 和 goose 迁移，但数据访问层仍未正式决定是否采用 ORM，以及采用哪一个 ORM。

当前约束：

- 业务里既有常规 CRUD，也有账本、调额、兑换码、显式锁和统计查询
- 团队需要尽快搭起稳定的模型、事务和仓储层，不想在基础 CRUD 上投入过多样板代码
- schema 变更已经决定统一由 goose 管理，因此 ORM 不需要承担迁移职责

## 选项

### 选项 A：GORM

- 优点：
  - Go 生态成熟，团队可参考资料多
  - 常规 CRUD、事务、关联、分页和模型映射都能较快落地
  - 与 PostgreSQL 组合常见，适合当前项目的开发速度要求
- 缺点：
  - 抽象层较厚，不适合把所有复杂查询都硬塞进 ORM
  - 若滥用模型关联和自动行为，容易产生不透明 SQL
- 风险与未知：
  - 如果没有明确边界，容易误用 `AutoMigrate` 或把关键账本逻辑写成不可控 ORM 链式调用
- 迁移/回滚成本：
  - 低，当前还未形成实际数据访问层实现

### 选项 B：sqlc / 手写 SQL 为主

- 优点：
  - SQL 可控性最高，适合复杂查询与性能敏感路径
  - 对 PostgreSQL 特性利用更直接
- 缺点：
  - CRUD 样板更多，项目初期交付速度较慢
  - 团队需要更早把 repository/query 组织方式一次性设计清楚
- 风险与未知：
  - 需求仍在演进时，修改成本会更高
- 迁移/回滚成本：
  - 中

### 选项 C：ent

- 优点：
  - 结构化建模能力强，代码生成体验好
  - 对关系建模一致性较强
- 缺点：
  - 需要接受更强的框架约束
  - 对当前项目来说引入成本高于 GORM
- 风险与未知：
  - 团队需要同时学习代码生成和新的建模方式
- 迁移/回滚成本：
  - 中

## 决策

选择 **GORM** 作为 MrChat 的默认 ORM。

理由：

- 它能在项目早期显著降低 CRUD 和事务样板成本
- 与 `Gin + PostgreSQL` 组合成熟，学习与落地成本都可控
- 它适合作为默认数据访问层，同时允许我们在关键路径回落到直接 SQL

## 影响

- 默认数据访问走 GORM 模型和事务接口
- 复杂查询、显式锁、账本结算、幂等更新等关键路径允许直接 SQL，不强求只用 ORM
- 不使用 GORM `AutoMigrate` 管理生产 schema；schema 变更统一走 goose

## 后续行动

1. 在工程骨架里初始化 GORM，并明确 repository/service 分层。
2. 为模型定义统一约定：主键、时间字段、软删除、JSON 字段映射。
3. 在代码规范中明确哪些场景优先直接 SQL，例如余额预扣、兑换码幂等、`FOR UPDATE` 锁定。
