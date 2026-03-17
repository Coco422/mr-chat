-- +goose Up
CREATE TABLE upstreams (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL,
    provider_type varchar(50) NOT NULL DEFAULT 'openai_compatible',
    base_url varchar(500) NOT NULL,
    auth_type varchar(32) NOT NULL DEFAULT 'bearer',
    auth_config_encrypted jsonb NOT NULL DEFAULT '{}'::jsonb,
    status varchar(32) NOT NULL DEFAULT 'active',
    timeout_seconds integer NOT NULL DEFAULT 60,
    cooldown_seconds integer NOT NULL DEFAULT 60,
    failure_threshold integer NOT NULL DEFAULT 3,
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_upstreams_name ON upstreams (LOWER(name));
CREATE INDEX idx_upstreams_status ON upstreams (status);
CREATE INDEX idx_upstreams_provider_type ON upstreams (provider_type);

CREATE TABLE models (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    model_key varchar(100) NOT NULL,
    display_name varchar(200) NOT NULL,
    provider_type varchar(50) NOT NULL DEFAULT 'openai_compatible',
    context_length integer NOT NULL DEFAULT 0,
    max_output_tokens integer NULL,
    pricing_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    capabilities_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    allowed_group_ids_json jsonb NOT NULL DEFAULT '[]'::jsonb,
    status varchar(32) NOT NULL DEFAULT 'active',
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_models_model_key ON models (LOWER(model_key));
CREATE INDEX idx_models_status ON models (status);

CREATE TABLE model_route_bindings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id uuid NOT NULL,
    group_id uuid NULL,
    upstream_id uuid NOT NULL,
    priority integer NOT NULL DEFAULT 1,
    status varchar(32) NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT fk_model_route_bindings_model FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    CONSTRAINT fk_model_route_bindings_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_model_route_bindings_upstream FOREIGN KEY (upstream_id) REFERENCES upstreams(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX uq_model_route_bindings_model_group_priority
    ON model_route_bindings (
        model_id,
        COALESCE(group_id, '00000000-0000-0000-0000-000000000000'::uuid),
        priority
    );
CREATE INDEX idx_model_route_bindings_upstream_id ON model_route_bindings (upstream_id);
CREATE INDEX idx_model_route_bindings_model_id ON model_route_bindings (model_id);

-- +goose Down
DROP TABLE IF EXISTS model_route_bindings;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS upstreams;
