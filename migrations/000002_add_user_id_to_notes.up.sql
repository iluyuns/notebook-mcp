ALTER TABLE notebook_notes
    ADD COLUMN IF NOT EXISTS user_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_notebook_notes_user_created_at
    ON notebook_notes (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notebook_notes_user_note_type_created_at
    ON notebook_notes (user_id, note_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notebook_notes_user_search
    ON notebook_notes USING GIN (
        to_tsvector('simple', user_id || ' ' || coalesce(title, '') || ' ' || coalesce(content, ''))
    );
