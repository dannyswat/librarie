-- AI provider capability configurations (credentials encrypted at rest)
CREATE TABLE ai_provider_configs (
    id                    uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    provider_key          text        NOT NULL,
    capability            text        NOT NULL,
    encrypted_credentials bytea       NOT NULL,
    model                 text        NOT NULL,
    is_enabled            boolean     NOT NULL DEFAULT TRUE,
    updated_at            timestamptz NOT NULL DEFAULT NOW(),
    UNIQUE (provider_key, capability)
);
