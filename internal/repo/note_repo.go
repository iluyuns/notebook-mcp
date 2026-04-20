package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"notebook-mcp/internal/model"

	"github.com/jackc/pgx/v5/pgtype"
)

type NoteRepo struct {
	db *sql.DB
}

func NewNoteRepo(db *sql.DB) *NoteRepo {
	return &NoteRepo{db: db}
}

func (r *NoteRepo) Save(ctx context.Context, req model.SaveNoteRequest) (model.Note, error) {
	metaJSON, err := json.Marshal(req.Metadata)
	if err != nil {
		return model.Note{}, err
	}

	var note model.Note
	var tags pgtype.FlatArray[string]
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO notebook_notes
		    (note_type, title, content, tags, session_id, source_instruction, metadata)
		VALUES
		    ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, note_type, title, content, tags, session_id, source_instruction, metadata, created_at, updated_at
	`,
		req.NoteType,
		strings.TrimSpace(req.Title),
		req.Content,
		pgtype.FlatArray[string](req.Tags),
		strings.TrimSpace(req.SessionID),
		strings.TrimSpace(req.SourceInstruction),
		metaJSON,
	).Scan(
		&note.ID,
		&note.NoteType,
		&note.Title,
		&note.Content,
		&tags,
		&note.SessionID,
		&note.SourceInstruction,
		&metaJSON,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		return model.Note{}, err
	}
	note.Tags = []string(tags)
	if err := json.Unmarshal(metaJSON, &note.Metadata); err != nil {
		return model.Note{}, err
	}
	return note, nil
}

func (r *NoteRepo) Search(ctx context.Context, req model.SearchNotesRequest) ([]model.Note, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var (
		rows *sql.Rows
		err  error
	)

	if req.NoteType == "" {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, note_type, title, content, tags, session_id, source_instruction, metadata, created_at, updated_at
			FROM notebook_notes
			WHERE (title ILIKE '%' || $1 || '%' OR content ILIKE '%' || $1 || '%')
			ORDER BY created_at DESC
			LIMIT $2
		`, req.Keyword, limit)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, note_type, title, content, tags, session_id, source_instruction, metadata, created_at, updated_at
			FROM notebook_notes
			WHERE note_type = $1
			  AND (title ILIKE '%' || $2 || '%' OR content ILIKE '%' || $2 || '%')
			ORDER BY created_at DESC
			LIMIT $3
		`, req.NoteType, req.Keyword, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]model.Note, 0, limit)
	for rows.Next() {
		var (
			note     model.Note
			metaJSON []byte
			tags     pgtype.FlatArray[string]
		)
		if err := rows.Scan(
			&note.ID,
			&note.NoteType,
			&note.Title,
			&note.Content,
			&tags,
			&note.SessionID,
			&note.SourceInstruction,
			&metaJSON,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, err
		}
		note.Tags = []string(tags)
		if err := json.Unmarshal(metaJSON, &note.Metadata); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}
