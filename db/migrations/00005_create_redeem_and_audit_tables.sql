-- +goose Up
CREATE TABLE redeem_codes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_no varchar(64) NOT NULL,
    code_hash varchar(255) NOT NULL,
    quota_amount bigint NOT NULL,
    max_redemptions integer NOT NULL DEFAULT 1,
    redeemed_count integer NOT NULL DEFAULT 0,
    status varchar(32) NOT NULL DEFAULT 'active',
    valid_from timestamptz NULL,
    valid_until timestamptz NULL,
    created_by uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_redeem_codes_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE UNIQUE INDEX uq_redeem_codes_code_hash ON redeem_codes (code_hash);
CREATE INDEX idx_redeem_codes_batch_no ON redeem_codes (batch_no);
CREATE INDEX idx_redeem_codes_status_valid_until ON redeem_codes (status, valid_until);

CREATE TABLE audit_logs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid NULL,
    actor_role varchar(20) NULL,
    action varchar(100) NOT NULL,
    resource_type varchar(100) NOT NULL,
    resource_id varchar(100) NULL,
    target_user_id uuid NULL,
    request_id varchar(64) NULL,
    ip_address varchar(64) NULL,
    user_agent text NULL,
    result varchar(20) NOT NULL,
    detail_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_audit_logs_actor_user FOREIGN KEY (actor_user_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_audit_logs_target_user FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_audit_logs_actor_user_created_at ON audit_logs (actor_user_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource_type_id ON audit_logs (resource_type, resource_id);
CREATE INDEX idx_audit_logs_action_created_at ON audit_logs (action, created_at DESC);

CREATE TABLE redeem_redemptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    redeem_code_id uuid NOT NULL,
    user_id uuid NOT NULL,
    quota_amount bigint NOT NULL,
    quota_log_id uuid NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_redeem_redemptions_code FOREIGN KEY (redeem_code_id) REFERENCES redeem_codes(id) ON DELETE CASCADE,
    CONSTRAINT fk_redeem_redemptions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_redeem_redemptions_code_user ON redeem_redemptions (redeem_code_id, user_id);
CREATE INDEX idx_redeem_redemptions_user_created_at ON redeem_redemptions (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS redeem_redemptions;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS redeem_codes;
