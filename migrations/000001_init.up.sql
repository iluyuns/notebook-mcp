CREATE TABLE IF NOT EXISTS notebook_users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notebook_users_status_created_at
    ON notebook_users (status, created_at DESC);

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

CREATE TABLE IF NOT EXISTS notebook_notes (
    id BIGSERIAL PRIMARY KEY,
    note_type VARCHAR(20) NOT NULL CHECK (note_type IN ('summary', 'context', 'note')),
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    session_id TEXT NOT NULL DEFAULT '',
    source_instruction TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    user_id TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notebook_notes_note_type ON notebook_notes (note_type);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_created_at ON notebook_notes (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_search ON notebook_notes USING GIN (
    to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(content, ''))
);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_user_created_at
    ON notebook_notes (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_user_note_type_created_at
    ON notebook_notes (user_id, note_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_user_search
    ON notebook_notes USING GIN (
        to_tsvector('simple', user_id || ' ' || coalesce(title, '') || ' ' || coalesce(content, ''))
    );

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

CREATE TABLE IF NOT EXISTS notebook_invitation_codes (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    max_uses INT NOT NULL DEFAULT 1 CHECK (max_uses > 0),
    used_count INT NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (used_count <= max_uses)
);

CREATE INDEX IF NOT EXISTS idx_notebook_invitation_codes_expires_at
    ON notebook_invitation_codes (expires_at);
