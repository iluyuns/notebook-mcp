CREATE TABLE IF NOT EXISTS notebook_oauth_clients (
    id BIGSERIAL PRIMARY KEY,
    client_id TEXT NOT NULL UNIQUE,
    client_name TEXT NOT NULL DEFAULT '',
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    scopes TEXT[] NOT NULL DEFAULT '{"mcp"}',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notebook_oauth_clients_active_created_at
    ON notebook_oauth_clients (is_active, created_at DESC);
