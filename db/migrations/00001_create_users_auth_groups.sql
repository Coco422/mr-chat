-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL,
    description text NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    permissions_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_groups_name ON groups (LOWER(name));

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    username varchar(50) NOT NULL,
    email varchar(100) NOT NULL,
    display_name varchar(100) NOT NULL,
    avatar_url varchar(500) NULL,
    role varchar(16) NOT NULL DEFAULT 'user',
    status varchar(16) NOT NULL DEFAULT 'active',
    quota bigint NOT NULL DEFAULT 0,
    used_quota bigint NOT NULL DEFAULT 0,
    primary_group_id uuid NULL,
    settings_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    aff_code varchar(32) NULL,
    inviter_id uuid NULL,
    last_login_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX uq_users_username ON users (LOWER(username));
CREATE UNIQUE INDEX uq_users_email ON users (LOWER(email));
CREATE INDEX idx_users_role_status ON users (role, status);
CREATE INDEX idx_users_primary_group_id ON users (primary_group_id);
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

ALTER TABLE users
    ADD CONSTRAINT fk_users_primary_group
    FOREIGN KEY (primary_group_id) REFERENCES groups(id);

ALTER TABLE users
    ADD CONSTRAINT fk_users_inviter
    FOREIGN KEY (inviter_id) REFERENCES users(id);

CREATE TABLE auths (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    auth_type varchar(32) NOT NULL,
    provider varchar(50) NULL,
    provider_subject varchar(255) NULL,
    password_hash varchar(255) NULL,
    verified_at timestamptz NULL,
    last_login_at timestamptz NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_auths_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_auths_provider_subject UNIQUE (provider, provider_subject)
);

CREATE INDEX idx_auths_user_id ON auths (user_id);

CREATE TABLE group_members (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id uuid NOT NULL,
    user_id uuid NOT NULL,
    member_role varchar(16) NOT NULL DEFAULT 'member',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_group_members_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_group_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_group_members_group_user UNIQUE (group_id, user_id)
);

CREATE INDEX idx_group_members_user_id ON group_members (user_id);

-- +goose Down
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS auths;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_inviter;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_primary_group;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS groups;
