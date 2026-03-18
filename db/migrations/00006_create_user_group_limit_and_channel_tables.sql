-- +goose Up
CREATE TABLE user_groups (
    id uuid PRIMARY KEY,
    name varchar(100) NOT NULL,
    description text NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    permissions_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO user_groups (id, name, description, status, permissions_json, metadata_json, created_at, updated_at)
SELECT
    id,
    name,
    description,
    status,
    permissions_json,
    '{}'::jsonb,
    created_at,
    updated_at
FROM groups
ON CONFLICT (id) DO NOTHING;

CREATE UNIQUE INDEX uq_user_groups_name ON user_groups (LOWER(name));

ALTER TABLE users
    ADD COLUMN user_group_id uuid NULL;

UPDATE users
SET user_group_id = primary_group_id
WHERE user_group_id IS NULL
  AND primary_group_id IS NOT NULL;

ALTER TABLE users
    ADD CONSTRAINT fk_users_user_group
    FOREIGN KEY (user_group_id) REFERENCES user_groups(id);

CREATE INDEX idx_users_user_group_id ON users (user_group_id);

CREATE TABLE channels (
    id uuid PRIMARY KEY,
    name varchar(100) NOT NULL,
    description text NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    billing_config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO channels (id, name, description, status, billing_config_json, metadata_json, created_at, updated_at)
SELECT DISTINCT
    g.id,
    g.name,
    g.description,
    g.status,
    '{}'::jsonb,
    '{}'::jsonb,
    now(),
    now()
FROM groups g
JOIN model_route_bindings mrb ON mrb.group_id = g.id
ON CONFLICT (id) DO NOTHING;

CREATE UNIQUE INDEX uq_channels_name ON channels (LOWER(name));
CREATE INDEX idx_channels_status ON channels (status);

ALTER TABLE model_route_bindings
    DROP CONSTRAINT IF EXISTS fk_model_route_bindings_group;

ALTER TABLE model_route_bindings
    RENAME COLUMN group_id TO channel_id;

ALTER TABLE model_route_bindings
    ADD CONSTRAINT fk_model_route_bindings_channel
    FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS uq_model_route_bindings_model_group_priority;

CREATE UNIQUE INDEX uq_model_route_bindings_model_channel_priority
    ON model_route_bindings (
        model_id,
        COALESCE(channel_id, '00000000-0000-0000-0000-000000000000'::uuid),
        priority
    );

CREATE INDEX idx_model_route_bindings_channel_id ON model_route_bindings (channel_id);

ALTER TABLE models
    RENAME COLUMN allowed_group_ids_json TO visible_user_group_ids_json;

CREATE TABLE user_group_model_limit_policies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_group_id uuid NOT NULL,
    model_id uuid NULL,
    hour_request_limit bigint NULL,
    week_request_limit bigint NULL,
    lifetime_request_limit bigint NULL,
    hour_token_limit bigint NULL,
    week_token_limit bigint NULL,
    lifetime_token_limit bigint NULL,
    status varchar(16) NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_user_group_model_limit_policies_user_group
        FOREIGN KEY (user_group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_group_model_limit_policies_model
        FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_user_group_model_limit_policies_group_model
    ON user_group_model_limit_policies (
        user_group_id,
        COALESCE(model_id, '00000000-0000-0000-0000-000000000000'::uuid)
    );
CREATE INDEX idx_user_group_model_limit_policies_status
    ON user_group_model_limit_policies (status);

CREATE TABLE user_limit_adjustments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    model_id uuid NULL,
    metric_type varchar(32) NOT NULL,
    window_type varchar(32) NOT NULL,
    delta bigint NOT NULL,
    expires_at timestamptz NULL,
    reason varchar(255) NULL,
    actor_user_id uuid NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_user_limit_adjustments_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_limit_adjustments_model
        FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_limit_adjustments_actor
        FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_user_limit_adjustments_user_created_at
    ON user_limit_adjustments (user_id, created_at DESC);
CREATE INDEX idx_user_limit_adjustments_user_model
    ON user_limit_adjustments (user_id, model_id);
CREATE INDEX idx_user_limit_adjustments_expires_at
    ON user_limit_adjustments (expires_at);

CREATE TABLE llm_request_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id varchar(64) NOT NULL,
    user_id uuid NOT NULL,
    user_group_id uuid NULL,
    conversation_id uuid NULL,
    message_id uuid NULL,
    model_id uuid NULL,
    channel_id uuid NULL,
    prompt_tokens bigint NOT NULL DEFAULT 0,
    completion_tokens bigint NOT NULL DEFAULT 0,
    total_tokens bigint NOT NULL DEFAULT 0,
    billed_quota bigint NOT NULL DEFAULT 0,
    status varchar(32) NOT NULL DEFAULT 'pending',
    error_code varchar(100) NULL,
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT uq_llm_request_logs_request_id UNIQUE (request_id),
    CONSTRAINT fk_llm_request_logs_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_llm_request_logs_user_group
        FOREIGN KEY (user_group_id) REFERENCES user_groups(id) ON DELETE SET NULL,
    CONSTRAINT fk_llm_request_logs_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE SET NULL,
    CONSTRAINT fk_llm_request_logs_message
        FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE SET NULL,
    CONSTRAINT fk_llm_request_logs_model
        FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE SET NULL,
    CONSTRAINT fk_llm_request_logs_channel
        FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE SET NULL
);

CREATE INDEX idx_llm_request_logs_user_started_at
    ON llm_request_logs (user_id, started_at DESC);
CREATE INDEX idx_llm_request_logs_user_model_started_at
    ON llm_request_logs (user_id, model_id, started_at DESC);
CREATE INDEX idx_llm_request_logs_status
    ON llm_request_logs (status);
CREATE INDEX idx_llm_request_logs_channel_id
    ON llm_request_logs (channel_id);

-- +goose Down
DROP TABLE IF EXISTS llm_request_logs;
DROP TABLE IF EXISTS user_limit_adjustments;
DROP TABLE IF EXISTS user_group_model_limit_policies;

ALTER TABLE models
    RENAME COLUMN visible_user_group_ids_json TO allowed_group_ids_json;

DROP INDEX IF EXISTS idx_model_route_bindings_channel_id;
DROP INDEX IF EXISTS uq_model_route_bindings_model_channel_priority;

ALTER TABLE model_route_bindings
    DROP CONSTRAINT IF EXISTS fk_model_route_bindings_channel;

ALTER TABLE model_route_bindings
    RENAME COLUMN channel_id TO group_id;

ALTER TABLE model_route_bindings
    ADD CONSTRAINT fk_model_route_bindings_group
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX uq_model_route_bindings_model_group_priority
    ON model_route_bindings (
        model_id,
        COALESCE(group_id, '00000000-0000-0000-0000-000000000000'::uuid),
        priority
    );

DROP TABLE IF EXISTS channels;

DROP INDEX IF EXISTS idx_users_user_group_id;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS fk_users_user_group;

ALTER TABLE users
    DROP COLUMN IF EXISTS user_group_id;

DROP TABLE IF EXISTS user_groups;
