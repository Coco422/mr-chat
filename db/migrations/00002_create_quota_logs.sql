-- +goose Up
CREATE TABLE quota_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    conversation_id uuid NULL,
    message_id uuid NULL,
    model_id uuid NULL,
    request_id varchar(64) NULL,
    log_type varchar(32) NOT NULL,
    delta_quota bigint NOT NULL,
    balance_after bigint NOT NULL,
    usage_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    operator_user_id uuid NULL,
    reason varchar(255) NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_quota_logs_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_quota_logs_operator FOREIGN KEY (operator_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_quota_logs_user_created_at ON quota_logs (user_id, created_at DESC);
CREATE INDEX idx_quota_logs_request_id ON quota_logs (request_id);
CREATE INDEX idx_quota_logs_log_type ON quota_logs (log_type);

-- +goose Down
DROP TABLE IF EXISTS quota_logs;
