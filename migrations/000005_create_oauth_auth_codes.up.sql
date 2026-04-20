CREATE TABLE IF NOT EXISTS notebook_oauth_auth_codes (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES notebook_users(id),
    client_id TEXT NOT NULL REFERENCES notebook_oauth_clients(client_id),
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    code_challenge_method VARCHAR(10) NOT NULL DEFAULT 'S256',
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notebook_oauth_auth_codes_user_client
    ON notebook_oauth_auth_codes (user_id, client_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notebook_oauth_auth_codes_expires_at
    ON notebook_oauth_auth_codes (expires_at);
