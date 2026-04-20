CREATE TABLE IF NOT EXISTS notebook_oauth_access_tokens (
    id BIGSERIAL PRIMARY KEY,
    access_token TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES notebook_users(id),
    client_id TEXT NOT NULL REFERENCES notebook_oauth_clients(client_id),
    scope TEXT NOT NULL DEFAULT 'mcp',
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notebook_oauth_access_tokens_user_client
    ON notebook_oauth_access_tokens (user_id, client_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notebook_oauth_access_tokens_expires_at
    ON notebook_oauth_access_tokens (expires_at);
