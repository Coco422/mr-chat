-- +goose Up
CREATE TABLE conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    title varchar(255) NOT NULL,
    model_id uuid NULL,
    status varchar(32) NOT NULL DEFAULT 'active',
    message_count integer NOT NULL DEFAULT 0,
    last_message_at timestamptz NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT fk_conversations_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_conversations_model FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE SET NULL
);

CREATE INDEX idx_conversations_user_status_last_message_at ON conversations (user_id, status, last_message_at DESC);
CREATE INDEX idx_conversations_deleted_at ON conversations (deleted_at);

CREATE TABLE messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid NOT NULL,
    user_id uuid NOT NULL,
    model_id uuid NULL,
    upstream_id uuid NULL,
    request_id varchar(64) NULL,
    role varchar(16) NOT NULL,
    content text NOT NULL DEFAULT '',
    reasoning_content text NULL,
    status varchar(32) NOT NULL DEFAULT 'pending',
    finish_reason varchar(50) NULL,
    usage_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    error_code varchar(100) NULL,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,
    CONSTRAINT fk_messages_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_model FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE SET NULL,
    CONSTRAINT fk_messages_upstream FOREIGN KEY (upstream_id) REFERENCES upstreams(id) ON DELETE SET NULL
);

CREATE INDEX idx_messages_conversation_created_at ON messages (conversation_id, created_at);
CREATE INDEX idx_messages_request_id ON messages (request_id);
CREATE INDEX idx_messages_status ON messages (status);
CREATE INDEX idx_messages_deleted_at ON messages (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
