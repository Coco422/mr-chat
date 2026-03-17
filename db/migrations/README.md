# Migrations

This directory stores PostgreSQL schema changes managed by `goose`.

## Naming

- Use ordered numeric prefixes for v0.1, for example:
  - `00001_init_users.sql`
  - `00002_init_models.sql`

## Rules

- Use SQL migrations by default
- Avoid mixing unrelated schema changes in a single file
- Production schema changes must not rely on GORM `AutoMigrate`
- Run migrations outside application startup

