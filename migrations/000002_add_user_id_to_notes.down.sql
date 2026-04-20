DROP INDEX IF EXISTS idx_notebook_notes_user_search;
DROP INDEX IF EXISTS idx_notebook_notes_user_note_type_created_at;
DROP INDEX IF EXISTS idx_notebook_notes_user_created_at;

ALTER TABLE notebook_notes
    DROP COLUMN IF EXISTS user_id;
