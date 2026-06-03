-- name: UpsertAIProviderConfig :one
INSERT INTO ai_provider_configs (provider_key, capability, encrypted_credentials, model, is_enabled, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (provider_key, capability) DO UPDATE
    SET encrypted_credentials = $3,
        model                 = $4,
        is_enabled            = $5,
        updated_at            = NOW()
RETURNING *;

-- name: GetAIProviderConfig :one
SELECT * FROM ai_provider_configs
WHERE provider_key = $1 AND capability = $2;

-- name: GetAIProviderConfigByID :one
SELECT * FROM ai_provider_configs WHERE id = $1;

-- name: ListAIProviderConfigs :many
SELECT * FROM ai_provider_configs
ORDER BY provider_key ASC, capability ASC;

-- name: ListAIProviderConfigsByProvider :many
SELECT * FROM ai_provider_configs
WHERE provider_key = $1
ORDER BY capability ASC;

-- name: ListEnabledAIProviderConfigs :many
SELECT * FROM ai_provider_configs
WHERE is_enabled = TRUE
ORDER BY provider_key ASC, capability ASC;

-- name: DeleteAIProviderConfig :exec
DELETE FROM ai_provider_configs WHERE id = $1;
