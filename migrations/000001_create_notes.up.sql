CREATE TABLE IF NOT EXISTS notebook_notes (
    id BIGSERIAL PRIMARY KEY,
    note_type VARCHAR(20) NOT NULL CHECK (note_type IN ('summary', 'context', 'note')),
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    session_id TEXT NOT NULL DEFAULT '',
    source_instruction TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notebook_notes_note_type ON notebook_notes (note_type);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_created_at ON notebook_notes (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notebook_notes_search ON notebook_notes USING GIN (
    to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(content, ''))
);
